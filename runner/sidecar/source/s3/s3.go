package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/loadbalanced"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/opentracing/opentracing-go"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type message struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceName, sourceURN string, x dfv1.S3Source, process source.Process, leadReplica bool) (source.HasPending, error) {
	logger := sharedutil.NewLogger().WithValues("source", x.Name, "bucket", x.Bucket)
	var accessKeyID string
	{
		secretName := x.Credentials.AccessKeyID.Name
		secret, err := secretInterface.Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get secret %q: %w", secretName, err)
		}
		accessKeyID = string(secret.Data[x.Credentials.AccessKeyID.Key])
	}
	var secretAccessKey string
	{
		secretName := x.Credentials.SecretAccessKey.Name
		secret, err := secretInterface.Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get secret %q: %w", secretName, err)
		}
		secretAccessKey = string(secret.Data[x.Credentials.SecretAccessKey.Key])
	}
	var sessionToken string
	{
		secretName := x.Credentials.SessionToken.Name
		secret, err := secretInterface.Get(ctx, secretName, metav1.GetOptions{})
		if err == nil {
			sessionToken = string(secret.Data[x.Credentials.SessionToken.Key])
		} else {
			// it is okay for sessionToken to be missing
			if !apierr.IsNotFound(err) {
				return nil, err
			}
		}
	}
	options := s3.Options{
		Region: x.Region,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey, SessionToken: sessionToken}, nil
		}),
	}
	if e := x.Endpoint; e != nil {
		options.EndpointResolver = s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
			return aws.Endpoint{URL: e.URL, SigningRegion: region, HostnameImmutable: true}, nil
		})
	}

	dir := filepath.Join(dfv1.PathVarRun, "sources", sourceName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create %q: %w", dir, err)
	}

	client := s3.New(options)

	return loadbalanced.New(ctx, loadbalanced.NewReq{
		Logger:       logger,
		PipelineName: pipelineName,
		StepName:     stepName,
		SourceName:   sourceName,
		SourceURN:    sourceURN,
		LeadReplica:  leadReplica,
		Concurrency:  int(x.Concurrency),
		PollPeriod:   x.PollPeriod.Duration,
		Process: func(ctx context.Context, msg []byte) error {
			span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("s3-source-%s", sourceName))
			defer span.Finish()
			key := string(msg)
			path := filepath.Join(dir, key)
			output, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &x.Bucket, Key: &key})
			if err != nil {
				return fmt.Errorf("failed to get object %q %q: %w", x.Bucket, key, err)
			}
			defer output.Body.Close()
			if err := syscall.Mkfifo(path, 0o600); sharedutil.IgnoreExist(err) != nil {
				return fmt.Errorf("failed to create fifo %q: %w", path, err)
			}
			defer os.Remove(path)
			go func() {
				defer runtime.HandleCrash()
				logger.Info("opening file", "key", key)
				file, err := os.OpenFile(path, os.O_WRONLY, os.ModeNamedPipe)
				if err != nil {
					logger.Error(err, "failed to open file", "path", path)
				}
				defer file.Close()
				if _, err := io.Copy(file, output.Body); err != nil {
					logger.Error(err, "failed to copy object to FIFO", "path", path)
				}
			}()
			return process(
				dfv1.ContextWithMeta(ctx, dfv1.Meta{
					Source: sourceURN,
					ID:     key,
					Time:   *output.LastModified,
				}),
				[]byte(sharedutil.MustJSON(message{Key: key, Path: path})),
			)
		},
		ListItems: func() ([]interface{}, error) {
			list, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(x.Bucket)})
			if err != nil {
				return nil, err
			}
			keys := make([]interface{}, len(list.Contents))
			for i, obj := range list.Contents {
				keys[i] = *obj.Key
			}
			return keys, nil
		},
		RemoveItem: func(item interface{}) error {
			key := item.(string)
			_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &x.Bucket, Key: &key})
			return err
		},
	})
}

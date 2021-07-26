package s3

import (
	"bytes"
	"context"
	"fmt"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/source"
	httpsource "github.com/argoproj-labs/argo-dataflow/runner/sidecar/source/http"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const concurrency = 4

var logger = sharedutil.NewLogger()

type job string

type s3Source struct {
	httpSource source.Interface
	jobs       chan job
}

func New(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace, pipelineName, stepName, sourceName string, x dfv1.S3Source, f source.Func, leadReplica bool) (source.Interface, error) {
	secretInterface := kubernetesInterface.CoreV1().Secrets(namespace)
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

	options := s3.Options{
		Region: x.Region,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey}, nil
		}),
	}
	if e := x.Endpoint; e != nil {
		options.EndpointResolver = s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
			return aws.Endpoint{URL: e.URL, SigningRegion: region, HostnameImmutable: true}, nil
		})
	}

	client := s3.New(options)
	bucket := x.Bucket
	jobs := make(chan job, concurrency)
	if leadReplica {
		endpoint := "http://" + pipelineName + "-" + stepName + "/sources/" + sourceName
		logger.Info("starting lead workers", "source", sourceName)
		// create N workers to support concurrency
		for w := 0; w < concurrency; w++ {
			go func() {
				defer runtime.HandleCrash()
				for j := range jobs {
					key := string(j)
					logger.Info("dispatching", "key", key)
					resp, err := http.Post(endpoint, "application/octet-stream", bytes.NewBufferString(key))
					if err != nil {
						logger.Error(err, "failed to process object", "key", key)
					} else {
						body, _ := io.ReadAll(resp.Body)
						_ = resp.Body.Close()
						if resp.StatusCode >= 300 {
							err := fmt.Errorf("%q: %q", resp.Status, body)
							logger.Error(err, "failed to process object", "key", key)
						} else {
							logger.Info("deleting object", "key", key)
							_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{
								Bucket: &bucket,
								Key:    &key,
							})
							if err != nil {
								logger.Error(err, "failed to delete object", "key", key)
							}
						}
					}
				}
			}()
		}
		// create leader Goroutine to poll for new files
		go func() {
			defer runtime.HandleCrash()
			ticker := time.NewTicker(x.PollPeriod)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					list, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &bucket})
					if err != nil {
						logger.Error(err, "failed to list bucket", "source", sourceName, "bucket", bucket)
					} else {
						for _, obj := range list.Contents {
							jobs <- job(*obj.Key)
						}
					}
				}
			}
		}()
	}
	return &s3Source{
		httpsource.New(sourceName, func(ctx context.Context, msg []byte) error {
			key := string(msg)
			logger.Info("received", "key", key)
			output, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
			if err != nil {
				return fmt.Errorf("failed to get object %q %q: %w", bucket, key, err)
			}
			defer output.Body.Close()
			path := filepath.Join(dfv1.PathVarRun, key)
			if err := syscall.Mkfifo(path, 0o666); sharedutil.IgnoreExist(err) != nil {
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
				logger.Info("copying data", "key", key)
				if _, err := io.Copy(file, output.Body); err != nil {
					logger.Error(err, "failed to copy object to FIFO", "path", path)
				}
			}()
			logger.Info("invoking f", "key", key)
			return f(ctx, []byte(path))
		}),
		jobs,
	}, nil
}

func (s *s3Source) Close() error {
	close(s.jobs)
	return s.httpSource.Close()
}

package s3

import (
	"bytes"
	"context"
	"crypto/tls"
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
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const concurrency = 4

var logger = sharedutil.NewLogger()

type s3Source struct {
	httpSource source.Interface
	jobs       workqueue.Interface
}

type message struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

func New(ctx context.Context, secretInterface corev1.SecretInterface, pipelineName, stepName, sourceName string, x dfv1.S3Source, f source.Func, leadReplica bool) (source.Interface, error) {
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
	dir := filepath.Join(dfv1.PathVarRun, "sources", sourceName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create %q: %w", dir, err)
	}

	client := s3.New(options)
	bucket := x.Bucket
	jobs := workqueue.New()
	authorization := sharedutil.RandString()
	if leadReplica {
		endpoint := "https://" + pipelineName + "-" + stepName + "/sources/" + sourceName
		logger.Info("starting lead workers", "source", sourceName, "endpoint", endpoint)
		// create N workers to support concurrency
		httpClient := &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
		for w := 0; w < concurrency; w++ {
			go func() {
				defer runtime.HandleCrash()
				for {
					item, shutdown := jobs.Get()
					if shutdown {
						return
					}
					func() {
						defer jobs.Done(item)
						key := item.(string)
						req, err := http.NewRequest("POST", endpoint, bytes.NewBufferString(key))
						if err != nil {
							panic(err)
						}
						req.Header.Set("Authorization", authorization)
						resp, err := httpClient.Do(req)
						if err != nil {
							logger.Error(err, "failed to process object", "key", key)
						} else {
							body, _ := io.ReadAll(resp.Body)
							_ = resp.Body.Close()
							if resp.StatusCode >= 300 {
								err := fmt.Errorf("%q: %q", resp.Status, body)
								logger.Error(err, "failed to process object", "bucket", bucket, "key", key)
							} else {
								logger.Info("deleting object", "bucket", bucket, "key", key)
								_, err := client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: &bucket, Key: &key})
								if err != nil {
									logger.Error(err, "failed to delete object", "bucket", bucket, "key", key)
								}
							}
						}
					}()
				}
			}()
		}
		// create leader Goroutine to poll for new files
		go func() {
			defer runtime.HandleCrash()
			ticker := time.NewTicker(x.PollPeriod.Duration)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					list, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &bucket})
					if err != nil {
						logger.Error(err, "failed to list bucket", "bucket", bucket)
					} else {
						for _, obj := range list.Contents {
							jobs.Add(*obj.Key)
						}
					}
				}
			}
		}()
	}
	return &s3Source{
		httpsource.New(sourceName, authorization, func(ctx context.Context, msg []byte) error {
			key := string(msg)
			output, err := client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &key})
			if err != nil {
				return fmt.Errorf("failed to get object %q %q: %w", bucket, key, err)
			}
			defer output.Body.Close()
			path := filepath.Join(dir, key)
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
			return f(ctx, []byte(sharedutil.MustJSON(message{Key: key, Path: path})))
		}),
		jobs,
	}, nil
}

func (s *s3Source) Close() error {
	s.jobs.ShutDown()
	return s.httpSource.Close()
}

package s3

import (
	"bytes"
	"context"
	"fmt"
	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/argoproj-labs/argo-dataflow/runner/util"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type s3Sink struct {
	client *s3.Client
	bucket string
	prog   *vm.Program
}

func New(ctx context.Context, kubernetesInterface kubernetes.Interface, namespace string, x dfv1.S3Sink) (sink.Interface, error) {
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
	prog, err := expr.Compile(x.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to compile %q: %w", x.Key, err)
	}
	return s3Sink{
		client: s3.New(options),
		bucket: x.Bucket,
		prog:   prog,
	}, nil
}

func (h s3Sink) Sink(msg []byte) error {
	res, err := expr.Run(h.prog, util.ExprEnv(msg))
	if err != nil {
		return err
	}
	key, ok := res.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", res)
	}
	_, err = h.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        &h.bucket,
		Key:           &key,
		Body:          bytes.NewBuffer(msg),
		ContentLength: int64(len(msg)),
	}, s3.WithAPIOptions(
		// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#unseekable-streaming-input
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	return err
}

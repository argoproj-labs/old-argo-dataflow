package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/argoproj-labs/argo-dataflow/runner/sidecar/sink"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/opentracing/opentracing-go"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/utils/pointer"
)

type s3Sink struct {
	sinkName string
	client   *s3.Client
	bucket   string
}

type message struct {
	Key  string `json:"key"`
	Path string `json:"path"`
}

func New(ctx context.Context, sinkName string, secretInterface v1.SecretInterface, x dfv1.S3Sink) (sink.Interface, error) {
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
	return s3Sink{sinkName, s3.New(options), x.Bucket}, nil
}

func (h s3Sink) Sink(ctx context.Context, msg []byte) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, fmt.Sprintf("s3-sink-%s", h.sinkName))
	defer span.Finish()
	message := &message{}
	if err := json.Unmarshal(msg, message); err != nil {
		return err
	}
	f, err := os.Open(message.Path)
	if err != nil {
		return fmt.Errorf("failed to open %q: %w", message.Path, err)
	}
	m, err := dfv1.MetaFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = h.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:  &h.bucket,
		Key:     &message.Key,
		Body:    f,
		Tagging: pointer.StringPtr(fmt.Sprintf("%s=%s,%s=%s", dfv1.MetaSource, m.Source, dfv1.MetaID, m.ID)),
	}, s3.WithAPIOptions(
		// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#unseekable-streaming-input
		v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
	))
	return err
}

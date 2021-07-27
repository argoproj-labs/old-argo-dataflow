package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
)

func init() {
	// https://stackoverflow.com/questions/67575681/is-aws-go-sdk-v2-integrated-with-local-minio-server
	http.HandleFunc("/minio/create-object", func(w http.ResponseWriter, r *http.Request) {
		region := "us-west-2"
		client := s3.New(s3.Options{
			Region: region,
			Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{AccessKeyID: "admin", SecretAccessKey: "password"}, nil
			}),
			EndpointResolver: s3.EndpointResolverFunc(func(region string, options s3.EndpointResolverOptions) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://minio:9000", SigningRegion: region, HostnameImmutable: true}, nil
			}),
		})
		bucket := "my-bucket"
		key := "my-key"
		content := "my-contents"
		_, err := client.PutObject(r.Context(), &s3.PutObjectInput{
			Bucket:        &bucket,
			Key:           &key,
			Body:          bytes.NewBufferString(content),
			ContentLength: int64(len(content)),
		}, s3.WithAPIOptions(
			// https://aws.github.io/aws-sdk-go-v2/docs/sdk-utilities/s3/#unseekable-streaming-input
			v4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware,
		))
		if err != nil {
			w.WriteHeader(400)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to create  object %q %q: %v", bucket, key, err)))
			return
		}
		w.WriteHeader(201)
	})
}

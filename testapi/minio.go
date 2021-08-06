package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func init() {
	region := "us-west-2"
	// https://stackoverflow.com/questions/67575681/is-aws-go-sdk-v2-integrated-with-local-minio-server
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
	http.HandleFunc("/minio/empty-bucket", func(w http.ResponseWriter, r *http.Request) {
		list, err := client.ListObjectsV2(r.Context(), &s3.ListObjectsV2Input{Bucket: &bucket})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to list objects %q: %v", bucket, err)))
			return
		}
		objects := make([]types.ObjectIdentifier, 0)
		for _, x := range list.Contents {
			objects = append(objects, types.ObjectIdentifier{Key: x.Key})
		}
		_, err = client.DeleteObjects(r.Context(), &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &types.Delete{
				Objects: objects,
			},
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(fmt.Sprintf("failed to delete objects %q: %v", bucket, err)))
			return
		}
		w.WriteHeader(204)
	})
	http.HandleFunc("/minio/create-object", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		content := "my-content"
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
			_, _ = w.Write([]byte(fmt.Sprintf("failed to create object %q %q: %v", bucket, key, err)))
			return
		}
		w.WriteHeader(201)
	})
}

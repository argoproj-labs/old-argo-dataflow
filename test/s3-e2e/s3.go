// +build test

package s3_e2e

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreateBucket(bucket string) func() {
	ctx := context.Background()
	cfg := getAWSCfg(ctx)

	s3Svc := s3.NewFromConfig(cfg)
	_, err := s3Svc.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		panic(err)
	}

	return func() {
		purgeBucket(bucket, s3Svc)
		_, err := s3Svc.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
		if err != nil {
			panic(err)
		}
	}
}

func purgeBucket(bucket string, s3Svc *s3.Client) {
	// Setup BatchDeleteIterator to iterate through a list of objects.
	deleteObject := func(bucket, key string) {
		_, err := s3Svc.DeleteObject(context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			panic(err)
		}
	}

	in := &s3.ListObjectsV2Input{Bucket: aws.String(bucket)}
	for {
		out, err := s3Svc.ListObjectsV2(context.Background(), in)
		if err != nil {
			panic(err)
		}

		for _, item := range out.Contents {
			deleteObject(bucket, *item.Key)
		}

		if out.IsTruncated {
			in.ContinuationToken = out.ContinuationToken
		} else {
			break
		}
	}
}

func PutS3Object(bucket string, key, value string) func() {
	ctx := context.Background()
	cfg := getAWSCfg(ctx)

	s3Svc := s3.NewFromConfig(cfg)
	CreateBucket(bucket)

	_, err := s3Svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(value),
	})
	if err != nil {
		panic(err)
	}

	return func() {
		purgeBucket(bucket, s3Svc)
		_, err := s3Svc.DeleteBucket(ctx, &s3.DeleteBucketInput{Bucket: aws.String(bucket)})
		if err != nil {
			panic(err)
		}
	}
}

func getAWSCfg(ctx context.Context) aws.Config {
	opts := []func(*awscfg.LoadOptions) error{
		awscfg.WithRegion("us-west-2"),
	}

	opts = append(opts, awscfg.WithCredentialsProvider(credentials.StaticCredentialsProvider{
		Value: aws.Credentials{AccessKeyID: "testing", SecretAccessKey: "testing", SessionToken: "testing"},
	}))
	opts = append(opts,
		awscfg.WithEndpointResolver(aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: "http://127.0.0.1:5000", HostnameImmutable: true, SigningRegion: region,
			}, nil
		},
		)))

	cfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		panic(err)
	}
	return cfg
}

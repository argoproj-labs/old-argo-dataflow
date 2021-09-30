package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestS3_GenURN(t *testing.T) {
	urn := S3{Bucket: "my-bucket"}.GenURN(cluster, namespace)
	assert.Equal(t, "urn:dataflow:s3:my-bucket", urn)
}

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPSource_GenURN(t *testing.T) {
	urn := HTTPSource{
		ServiceName: "my-name",
	}.GenURN(cluster, namespace)
	assert.Equal(t, "urn:dataflow:http:https://my-name.svc.my-ns.my-cluster", urn)
}

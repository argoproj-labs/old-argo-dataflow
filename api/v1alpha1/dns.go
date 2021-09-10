package v1alpha1

import (
	"context"
	"fmt"
	"strings"
)

// https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/
func dnsName(ctx context.Context, kind, name string) string {
	cluster, namespace := GetMetaCluster(ctx), GetMetaNamespace(ctx)
	switch kind {
	case "Service":
		return fmt.Sprintf("%s.svc.%s.%s", name, namespace, cluster)
	default:
		return fmt.Sprintf("%s.%s.%s.%s", name, strings.ToLower(kind), namespace, cluster)
	}
}

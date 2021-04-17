package bus

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//go:generate kustomize build --load_restrictor=none ../../config/kafka -o manifests/kafka-default.yaml
//go:generate kustomize build --load_restrictor=none ../../config/stan -o manifests/stan-default.yaml

//go:embed manifests
var manifests embed.FS

var (
	logger           = zap.New()
	restConfig       = ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
)

func Install(ctx context.Context, name, namespace string) error {
	v, err := readManifests(name)
	if err != nil {
		panic(err.Error())
	}
	if os.IsNotExist(err) {
		return nil
	}
	logger.Info("installing", "name", name, "namespace", namespace)
	list, err := util.SplitYAML(v)
	if err != nil {
		return fmt.Errorf("failed to split YAML: %w", err)
	}
	for _, item := range list.Items {
		if err := apply(ctx, namespace, &item); err != nil {
			return err
		}
	}
	return nil
}

func readManifests(name string) ([]byte, error) {
	return manifests.ReadFile(filepath.Join("manifests", name+".yaml"))
}

func apply(ctx context.Context, namespace string, item *unstructured.Unstructured) error {

	if item.GetAnnotations() == nil {
		item.SetAnnotations(map[string]string{})
	}
	x := item.GetAnnotations()
	x[dfv1.KeyHash] = util.MustHash(item)
	item.SetAnnotations(x)
	item.SetNamespace(namespace)

	switch item.GetKind() {
	case "ClusterRoleBinding":
		x := &rbacv1.ClusterRoleBinding{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, x); err != nil {
			return fmt.Errorf("failed to convert from unstructured: %w", err)
		}
		for _, s := range x.Subjects {
			s.Namespace = namespace
		}
		if v, err := runtime.DefaultUnstructuredConverter.ToUnstructured(x); err != nil {
			return fmt.Errorf("failed to convert to unstructured: %w", err)
		} else {
			item.Object = v
		}
	case "RoleBinding":
		x := &rbacv1.RoleBinding{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, x); err != nil {
			return fmt.Errorf("failed to convert from unstructured: %w", err)
		}
		for _, s := range x.Subjects {
			s.Namespace = namespace
		}
		if v, err := runtime.DefaultUnstructuredConverter.ToUnstructured(x); err != nil {
			return fmt.Errorf("failed to convert to unstructured: %w", err)
		} else {
			item.Object = v
		}
	}

	resourceInterface := resourceInterface(item, namespace)
	if _, err := resourceInterface.Create(ctx, item, metav1.CreateOptions{}); err != nil {
		if apierr.IsAlreadyExists(err) {
			old, err := resourceInterface.Get(ctx, item.GetName(), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get %s/%s: %w", item.GetKind(), item.GetName(), err)
			}
			if old.GetAnnotations()[dfv1.KeyHash] != item.GetAnnotations()[dfv1.KeyHash] {
				logger.Info("resource already exists, but hash has changed, deleting and re-creating")
				if err := resourceInterface.Delete(ctx, item.GetName(), metav1.DeleteOptions{}); err != nil {
					return fmt.Errorf("failed to delete changed %s/%s: %w", item.GetKind(), item.GetName(), err)
				}
				if _, err := resourceInterface.Create(ctx, item, metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("failed to re-create %s/%s: %w", item.GetKind(), item.GetName(), err)
				}
			}
		} else {
			return fmt.Errorf("failed to create %s/%s: %w", item.GetKind(), item.GetName(), err)
		}
	} else {
		logger.Info("created", "kind", item.GetKind(), "name", item.GetName())
	}
	return nil
}

func resourceInterface(item *unstructured.Unstructured, namespace string) dynamic.ResourceInterface {
	gvr := item.GroupVersionKind().GroupVersion().WithResource(util.Resource(item.GetKind()))
	var resourceInterface dynamic.ResourceInterface
	if strings.HasPrefix(item.GetKind(), "Cluster") {
		resourceInterface = dynamicInterface.Resource(gvr)
	} else {
		resourceInterface = dynamicInterface.Resource(gvr).Namespace(namespace)
	}
	return resourceInterface
}

func Uninstall(ctx context.Context, name, namespace string) error {
	v, err := readManifests(name)
	if os.IsNotExist(err) {
		return nil
	}
	logger.Info("un-installing", "name", name, "namespace", namespace)
	list, err := util.SplitYAML(v)
	if err != nil {
		return fmt.Errorf("failed to split YAML: %w", err)
	}
	for _, item := range list.Items {
		if err := _delete(ctx, namespace, &item); err != nil {
			return err
		}
	}
	return nil
}

func _delete(ctx context.Context, namespace string, item *unstructured.Unstructured) error {
	resourceInterface := resourceInterface(item, namespace)
	if err := resourceInterface.Delete(ctx, item.GetName(), metav1.DeleteOptions{}); err != nil {
		if dfv1.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete %s/%s: %w", item.GetKind(), item.GetName(), err)
		}
	} else {
		logger.Info("deleted", "kind", item.GetKind(), "name", item.GetName())
	}
	return nil
}
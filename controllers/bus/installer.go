package bus

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/rest"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger  = zap.New()
	images  = make(map[string]string)
	enabled = os.Getenv(dfv1.EnvInstaller) == "true"
)

type Installer interface {
	Install(ctx context.Context, name, namespace string) error
}

func NewInstaller() Installer {
	if enabled {
		restConfig := ctrl.GetConfigOrDie()
		return &installer{
			restConfig:       restConfig,
			dynamicInterface: dynamic.NewForConfigOrDie(restConfig),
		}
	}
	return nil
}

type installer struct {
	restConfig       *rest.Config
	dynamicInterface dynamic.Interface
}

func init() {
	v := os.Getenv(dfv1.EnvInstallerImages)
	if v == "" {
		v = `
{
  "nats-streaming": "docker.io/nats-streaming",
  "nats": "docker.io/nats",
  "quay.io/argoproj/dataflow-runner": "quay.io/argoproj/dataflow-runner",
  "solsson/kafka-initutils": "docker.io/solsson/kafka-initutils",
  "solsson/kafka": "docker.io/solsson/kafka"
}
`
	}
	util.MustUnJSON(v, &images)
	logger.Info("installer config",
		"enabled", enabled,
		"images", images,
	)
}

func imageName(x string) (string, error) {
	parts := strings.SplitN(x, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("expected format image:name, got %q", x)
	}
	name := parts[0]
	tag := parts[1]
	if y, ok := images[name]; ok {
		parts := strings.SplitN(y, ":", 2)
		newName := parts[0]
		newTag := tag
		if len(parts) == 2 {
			newTag = parts[1]
		}
		return newName + ":" + newTag, nil
	} else {
		return x, nil
	}
}

func (i installer) Install(ctx context.Context, name, namespace string) error {
	filename := filepath.Join("config", name+".yaml")
	v, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		logger.Info("bus not found", "name", name)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	logger.Info("installing", "name", name, "namespace", namespace)
	list, err := util.SplitYAML(v)
	if err != nil {
		return fmt.Errorf("failed to split YAML: %w", err)
	}
	for _, item := range list.Items {
		if err := i.apply(ctx, namespace, &item); err != nil {
			return err
		}
	}
	return nil
}

func (i installer) apply(ctx context.Context, namespace string, item *unstructured.Unstructured) error {
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
	case "StatefulSet":
		x := &appsv1.StatefulSet{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(item.Object, x); err != nil {
			return fmt.Errorf("failed to convert from unstructured: %w", err)
		}
		for i, c := range x.Spec.Template.Spec.Containers {
			if v, err := imageName(c.Image); err != nil {
				return err
			} else {
				c.Image = v
			}
			x.Spec.Template.Spec.Containers[i] = c
		}
		if v, err := runtime.DefaultUnstructuredConverter.ToUnstructured(x); err != nil {
			return fmt.Errorf("failed to convert to unstructured: %w", err)
		} else {
			item.Object = v
		}
	}

	if item.GetAnnotations() == nil {
		item.SetAnnotations(map[string]string{})
	}
	if item.GetLabels() == nil {
		item.SetLabels(map[string]string{})
	}

	labels := item.GetLabels()
	labels["app.kubernetes.io/managed-by"] = "argo-dataflow"
	labels["app.kubernetes.io/created-by"] = "controller-manager"
	item.SetLabels(labels)

	item.SetNamespace(namespace)

	annotations := item.GetAnnotations()
	annotations[dfv1.KeyHash] = util.MustHash(item)
	item.SetAnnotations(annotations)

	resourceInterface := i.resourceInterface(item, namespace)
	if _, err := resourceInterface.Create(ctx, item, metav1.CreateOptions{}); err != nil {
		if apierr.IsAlreadyExists(err) {
			old, err := resourceInterface.Get(ctx, item.GetName(), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get %s/%s: %w", item.GetKind(), item.GetName(), err)
			}
			if old.GetAnnotations()[dfv1.KeyHash] != item.GetAnnotations()[dfv1.KeyHash] {
				logger.Info("resource already exists, but hash has changed, deleting and re-creating", "kind", item.GetKind(), "name", item.GetName())
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

func (i installer) resourceInterface(item *unstructured.Unstructured, namespace string) dynamic.ResourceInterface {
	gvr := item.GroupVersionKind().GroupVersion().WithResource(util.Resource(item.GetKind()))
	var resourceInterface dynamic.ResourceInterface
	if strings.HasPrefix(item.GetKind(), "Cluster") {
		resourceInterface = i.dynamicInterface.Resource(gvr)
	} else {
		resourceInterface = i.dynamicInterface.Resource(gvr).Namespace(namespace)
	}
	return resourceInterface
}

func (i installer) Uninstall(ctx context.Context, name, namespace string) error {
	filename := filepath.Join("config", name+".yaml")
	v, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	logger.Info("un-installing", "name", name, "namespace", namespace)
	list, err := util.SplitYAML(v)
	if err != nil {
		return fmt.Errorf("failed to split YAML: %w", err)
	}
	for _, item := range list.Items {
		if err := i._delete(ctx, namespace, &item); err != nil {
			return err
		}
	}
	return nil
}

func (i installer) _delete(ctx context.Context, namespace string, item *unstructured.Unstructured) error {
	resourceInterface := i.resourceInterface(item, namespace)
	if err := resourceInterface.Delete(ctx, item.GetName(), metav1.DeleteOptions{}); err != nil {
		if util.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete %s/%s: %w", item.GetKind(), item.GetName(), err)
		}
	} else {
		logger.Info("deleted", "kind", item.GetKind(), "name", item.GetName())
	}
	return nil
}

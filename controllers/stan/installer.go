package stan

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/argoproj-labs/argo-dataflow/api/util"
	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

//go:generate kustomize build --load_restrictor=none ../../config/nats -o nats.yaml
//go:embed nats.yaml
var nats string

//go:generate kustomize build --load_restrictor=none ../../config/stan -o stan.yaml
//go:embed stan.yaml
var stan string

var (
	logger           = zap.New()
	restConfig       = ctrl.GetConfigOrDie()
	dynamicInterface = dynamic.NewForConfigOrDie(restConfig)
)

func Install(ctx context.Context, namespace string) error {

	if cancel, ok := cancelUninstall[namespace]; ok {
		logger.Info("cancelling un-install", "namespace", namespace)
		cancel()
		delete(cancelUninstall, namespace)
	}

	logger.Info("installing", "namespace", namespace)
	for _, data := range []string{nats, stan} {
		list, err := util.SplitYAML(data)
		if err != nil {
			return fmt.Errorf("failed to split YAMl: %w", err)
		}
		for _, item := range list.Items {
			if err := apply(ctx, namespace, &item); err != nil {
				return err
			}
		}
	}
	return nil
}

func apply(ctx context.Context, namespace string, item *unstructured.Unstructured) error {

	if item.GetAnnotations() == nil {
		item.SetAnnotations(map[string]string{})
	}
	x := item.GetAnnotations()
	x[dfv1.KeyHash] = util.MustHash(item)
	item.SetAnnotations(x)

	gvr := item.GroupVersionKind().GroupVersion().WithResource(util.Resource(item.GetKind()))
	resourceInterface := dynamicInterface.Resource(gvr).Namespace(namespace)
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

func Uninstall(ctx context.Context, namespace string) error {
	logger.Info("un-installing", "namespace", namespace)
	for _, data := range []string{nats, stan} {
		list, err := util.SplitYAML(data)
		if err != nil {
			return fmt.Errorf("failed to split YAML: %w", err)
		}
		for _, item := range list.Items {
			if err := _delete(ctx, namespace, &item); err != nil {
				return err
			}
		}
	}
	return nil
}

var cancelUninstall = make(map[string]func())

func UninstallAfter(ctx context.Context, namespace string, after time.Duration) {
	if _, ok := cancelUninstall[namespace]; ok {
		logger.Info("already un-installing")
		return
	}
	logger.Info("un-installing", "after", after.String())
	ctx, cancelUninstall[namespace] = context.WithCancel(ctx)
	time.AfterFunc(after, func() {
		if err := Uninstall(ctx, namespace); err != nil {
			logger.Error(err, "failed to un-installed after")
		}
	})
}

func _delete(ctx context.Context, namespace string, item *unstructured.Unstructured) error {
	gvr := item.GroupVersionKind().GroupVersion().WithResource(util.Resource(item.GetKind()))
	resourceInterface := dynamicInterface.Resource(gvr).Namespace(namespace)
	if err := resourceInterface.Delete(ctx, item.GetName(), metav1.DeleteOptions{}); err != nil {
		if dfv1.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete %s/%s: %w", item.GetKind(), item.GetName(), err)
		}
	} else {
		logger.Info("deleted", "kind", item.GetKind(), "name", item.GetName())
	}
	return nil
}

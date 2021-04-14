package v1alpha1

import (
	"os"
	"strings"

	apierr "k8s.io/apimachinery/pkg/api/errors"
)

func IgnoreIsExist(err error) error {
	if os.IsExist(err) {
		return nil
	}
	return err
}

func IgnoreAlreadyExists(err error) error {
	if apierr.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func IgnoreContainerNotFound(err error) error {
	if err != nil && strings.Contains(err.Error(), "container not found") {
		return nil
	}
	return err
}

func IgnoreConflict(err error) error {
	if apierr.IsConflict(err) {
		return nil
	}
	return err
}

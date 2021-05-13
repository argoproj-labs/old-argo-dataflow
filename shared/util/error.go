package util

import (
	"errors"
	"os"
	"strings"

	apierr "k8s.io/apimachinery/pkg/api/errors"
)

func IgnorePermission(err error) error {
	if errors.Is(err, os.ErrPermission) {
		return nil
	}
	return err
}

func IgnoreExist(err error) error {
	if errors.Is(err, os.ErrExist) {
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

func IgnoreNotFound(err error) error {
	if apierr.IsNotFound(err) {
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

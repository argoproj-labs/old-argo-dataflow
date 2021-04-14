package util

import (
	"os"

	"github.com/go-git/go-git/v5"
)

func IgnoreIsExist(err error) error {
	if os.IsExist(err) {
		return nil
	}
	return err
}

func IgnoreErrRepositoryAlreadyExists(err error) error {
	if err == git.ErrRepositoryAlreadyExists {
		return nil
	}
	return err
}

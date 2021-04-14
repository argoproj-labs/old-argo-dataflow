package init

import (
	"github.com/go-git/go-git/v5"
)

func IgnoreErrRepositoryAlreadyExists(err error) error {
	if err == git.ErrRepositoryAlreadyExists {
		return nil
	}
	return err
}

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/utils/strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func Init() error {
	if err := unmarshallSpec(); err != nil {
		return err
	}
	info.Info("creating in fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOIn, 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	info.Info("creating out fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOOut, 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if g := spec.Git; g != nil {
		info.Info("cloning", "url", g.URL, "checkout", dfv1.PathCheckout)
		if _, err := git.PlainClone(dfv1.PathCheckout, false, &git.CloneOptions{
			URL:           g.URL,
			Progress:      os.Stdout,
			SingleBranch:  true, // checkout faster
			Depth:         1,    // checkout faster
			ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
		}); IgnoreErrRepositoryAlreadyExists(err) != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
		path := filepath.Join(dfv1.PathCheckout, g.Path)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}
		info.Info("moving checked out code", "path", path, "wd", dfv1.PathWorkingDir)
		if err := os.Rename(path, dfv1.PathWorkingDir); IgnoreIsExist(err) != nil {
			return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
		}
	} else if h := spec.Handler; h != nil {
		info.Info("setting up handler", "runtime", h.Runtime, "code", strings.ShortenString(h.Code, 32)+"...")
		if err := ioutil.WriteFile(dfv1.PathHandlerFile, []byte(h.Code), 0600); err != nil {
			return fmt.Errorf("failed to create code file: %w", err)
		}
	}
	return nil
}

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

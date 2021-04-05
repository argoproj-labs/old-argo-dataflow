package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/otiai10/copy"
	"k8s.io/utils/strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func Init() error {
	if err := unmarshallSpec(); err != nil {
		return err
	}
	log.Info("creating in fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOIn, 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	log.Info("creating out fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOOut, 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if g := spec.Git; g != nil {
		log.Info("cloning", "url", g.URL, "checkout", dfv1.PathCheckout)
		if _, err := git.PlainClone(dfv1.PathCheckout, false, &git.CloneOptions{
			URL:           g.URL,
			Progress:      os.Stdout,
			SingleBranch:  true, // checkout faster
			Depth:         1,    // checkout faster
			ReferenceName: plumbing.NewBranchReferenceName(g.GetBranch()),
		}); IgnoreErrRepositoryAlreadyExists(err) != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
		path := g.GetPath()
		log.Info("moving checked out code", "path", path)
		if err := os.Rename(filepath.Join(dfv1.PathCheckout, path), dfv1.PathWorkingDir); IgnoreIsExist(err) != nil {
			return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
		}
	} else if h := spec.Handler; h != nil {
		log.Info("setting up handler", "runtime", h.Runtime)
		workingDir := filepath.Join(dfv1.PathRuntimes, string(h.Runtime))
		if err := os.Mkdir(filepath.Dir(workingDir), 0700); IgnoreIsExist(err) != nil {
			return fmt.Errorf("failed to create runtimes dir: %w", err)
		}
		log.Info("moving runtime", "runtime", h.Runtime)
		if err := copy.Copy(filepath.Join("runtimes", string(h.Runtime)), workingDir); err != nil {
			return fmt.Errorf("failed to move runtimes: %w", err)
		}
		log.Info("creating code file", "code", strings.ShortenString(h.Code, 32)+"...")
		if err := ioutil.WriteFile(filepath.Join(workingDir, h.Runtime.HandlerFile()), []byte(h.Code), 0600); err != nil {
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

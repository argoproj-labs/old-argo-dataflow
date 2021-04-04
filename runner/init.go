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
	if err := syscall.Mkfifo(filepath.Join(dfv1.PathVarRun, "in"), 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	log.Info("creating out fifo")
	if err := syscall.Mkfifo(filepath.Join(dfv1.PathVarRun, "out"), 0600); IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if h := spec.Handler; h != nil {
		runtime := h.Runtime
		log.Info("setting up handler", "runtime", runtime)
		workingDir := filepath.Join(dfv1.PathVarRunRuntimes, string(runtime))
		if err := os.Mkdir(filepath.Dir(workingDir), 0700); IgnoreIsExist(err) != nil {
			return fmt.Errorf("failed to create runtimes dir: %w", err)
		}
		if url := h.URL; url != "" {
			checkout := filepath.Join(dfv1.PathVarRun, "checkout",)
			log.Info("cloning", "url", url, "checkout", checkout)
			if _, err := git.PlainClone(checkout, false, &git.CloneOptions{
				URL:           url,
				Progress:      os.Stdout,
				SingleBranch:  true, // checkout faster
				Depth:         1,    // checkout faster
				ReferenceName: plumbing.NewBranchReferenceName(h.GetBranch()),
			}); IgnoreErrRepositoryAlreadyExists(err) != nil {
				return fmt.Errorf("failed to clone repo: %w", err)
			}
			log.Info("moving checked out code", "path", h.Path)
			if err := os.Rename(filepath.Join(checkout, h.GetPath()), workingDir); IgnoreIsExist(err) != nil {
				return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
			}
		} else if code := h.Code; code != "" {
			log.Info("moving runtime", "runtime", runtime)
			if err := copy.Copy(filepath.Join("runtimes", string(runtime)), workingDir); err != nil {
				return fmt.Errorf("failed to move runtimes: %w", err)
			}
			log.Info("creating code file", "code", strings.ShortenString(code, 32)+"...")
			if err := ioutil.WriteFile(filepath.Join(workingDir, runtime.HandlerFile()), []byte(code), 0600); err != nil {
				return fmt.Errorf("failed to create code file: %w", err)
			}
		} else {
			panic("invalid handler")
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
	if err == git.ErrRepositoryAlreadyExists{
		return nil
	}
	return err
}

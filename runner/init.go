package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/otiai10/copy"
	"k8s.io/utils/strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

func initCmd() error {
	if err := unmarshallFn(); err != nil {
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
	if h := fn.Spec.Handler; h != nil {
		log.Info("setting up handler", "runtime", h.Runtime)
		workingDir := filepath.Join(dfv1.PathVarRunRuntimes, string(h.Runtime))
		if err := os.Mkdir(filepath.Dir(workingDir), 0700); err != nil {
			return fmt.Errorf("failed to create runtimes dir: %w", err)
		}
		if err := copy.Copy(filepath.Join("runtimes", string(h.Runtime)), workingDir); err != nil {
			return fmt.Errorf("failed to move runtimes: %w", err)
		}
		if url := h.URL; url != "" {
			log.Info("cloning", "url", url)
			if _, err := git.PlainClone(filepath.Join(dfv1.PathVarRun, "code"), false, &git.CloneOptions{
				URL:          url,
				Progress:     os.Stdout,
				SingleBranch: true,
			});
				err != nil {
				return fmt.Errorf("failed to clone handle: %w", err)
			}
		} else if code := h.Code; code != "" {
			log.Info("creating code file", "code", strings.ShortenString(code, 32)+"...")
			if err := ioutil.WriteFile(filepath.Join(workingDir, h.Runtime.HandlerFile()), []byte(code), 0600); err != nil {
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

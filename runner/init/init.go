package init

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/argoproj-labs/argo-dataflow/runner/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/klog/klogr"
	"k8s.io/utils/strings"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
)

var logger = klogr.New()

func Exec() error {
	spec, err := util.UnmarshallSpec()
	if err != nil {
		return err
	}
	logger.Info("creating in fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOIn, 0600); util.IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create input FIFO: %w", err)
	}
	logger.Info("creating out fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOOut, 0600); util.IgnoreIsExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if g := spec.Git; g != nil {
		logger.Info("cloning", "url", g.URL, "checkout", dfv1.PathCheckout)
		if _, err := git.PlainClone(dfv1.PathCheckout, false, &git.CloneOptions{
			URL:           g.URL,
			Progress:      os.Stdout,
			SingleBranch:  true, // checkout faster
			Depth:         1,    // checkout faster
			ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
		}); util.IgnoreErrRepositoryAlreadyExists(err) != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
		path := filepath.Join(dfv1.PathCheckout, g.Path)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}
		logger.Info("moving checked out code", "path", path, "wd", dfv1.PathWorkingDir)
		if err := os.Rename(path, dfv1.PathWorkingDir); util.IgnoreIsExist(err) != nil {
			return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
		}
	} else if h := spec.Handler; h != nil {
		logger.Info("setting up handler", "runtime", h.Runtime, "code", strings.ShortenString(h.Code, 32)+"...")
		if err := ioutil.WriteFile(dfv1.PathHandlerFile, []byte(h.Code), 0600); err != nil {
			return fmt.Errorf("failed to create code file: %w", err)
		}
	}
	return nil
}

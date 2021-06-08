package init

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	util2 "github.com/argoproj-labs/argo-dataflow/shared/util"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/utils/strings"
)

var logger = util2.NewLogger()

// due to main container crashing, the init container may be started many times, so each operation we perform should be
// idempontent, i.e. if we copy a file to shared volume, and it already exists, we should ignore that error
func Exec() error {
	for _, name := range []string{dfv1.PathKill, dfv1.PathPreStop} {
		logger.Info("copying binary", "name", name)
		a := filepath.Join("/bin", filepath.Base(name))
		src, err := os.Open(a)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", a, err)
		}
		dst, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o500)
		if util2.IgnorePermission(util2.IgnoreExist(err)) != nil {
			return fmt.Errorf("failed to open %s: %w", name, err)
		} else if err == nil {
			if _, err := io.Copy(dst, src); util2.IgnoreExist(err) != nil {
				return fmt.Errorf("failed to create input FIFO: %w", err)
			}
		}
	}
	step := dfv1.Step{}
	util2.MustUnJSON(os.Getenv(dfv1.EnvStep), &step)

	if step.Spec.GetIn().FIFO {
		logger.Info("creating in fifo")
		if err := syscall.Mkfifo(dfv1.PathFIFOIn, 0o600); util2.IgnoreExist(err) != nil {
			return fmt.Errorf("failed to create input FIFO: %w", err)
		}
	}
	logger.Info("creating out fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOOut, 0o600); util2.IgnoreExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if g := step.Spec.Git; g != nil {
		logger.Info("cloning", "url", g.URL, "checkout", dfv1.PathCheckout)
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
		logger.Info("moving checked out code", "path", path, "wd", dfv1.PathWorkingDir)
		if err := os.Rename(path, dfv1.PathWorkingDir); util2.IgnoreExist(err) != nil {
			return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
		}
	} else if h := step.Spec.Handler; h != nil {
		logger.Info("setting up handler", "runtime", h.Runtime, "code", strings.ShortenString(h.Code, 32)+"...")
		if err := ioutil.WriteFile(dfv1.PathHandlerFile, []byte(h.Code), 0o600); err != nil {
			return fmt.Errorf("failed to create code file: %w", err)
		}
	}
	return nil
}

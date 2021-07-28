package init

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"

	dfv1 "github.com/argoproj-labs/argo-dataflow/api/v1alpha1"
	sharedutil "github.com/argoproj-labs/argo-dataflow/shared/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/utils/strings"
)

var logger = sharedutil.NewLogger()

// due to main container crashing, the init container may be started many times, so each operation we perform should be
// idempontent, i.e. if we copy a file to shared volume, and it already exists, we should ignore that error
func Exec(ctx context.Context) error {
	for _, name := range []string{dfv1.PathKill, dfv1.PathPreStop} {
		logger.Info("copying binary", "name", name)
		a := filepath.Join("/bin", filepath.Base(name))
		src, err := os.Open(a)
		if err != nil {
			return fmt.Errorf("failed to open %s: %w", a, err)
		}
		dst, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0o500)
		if sharedutil.IgnorePermission(sharedutil.IgnoreExist(err)) != nil {
			return fmt.Errorf("failed to open %s: %w", name, err)
		} else if err == nil {
			if _, err := io.Copy(dst, src); sharedutil.IgnoreExist(err) != nil {
				return fmt.Errorf("failed to create input FIFO: %w", err)
			}
		}
	}
	step := dfv1.Step{}
	sharedutil.MustUnJSON(os.Getenv(dfv1.EnvStep), &step)

	if step.Spec.GetIn().FIFO {
		logger.Info("creating in fifo")
		if err := syscall.Mkfifo(dfv1.PathFIFOIn, 0o600); sharedutil.IgnoreExist(err) != nil {
			return fmt.Errorf("failed to create input FIFO: %w", err)
		}
	}
	logger.Info("creating out fifo")
	if err := syscall.Mkfifo(dfv1.PathFIFOOut, 0o600); sharedutil.IgnoreExist(err) != nil {
		return fmt.Errorf("failed to create output FIFO: %w", err)
	}
	if g := step.Spec.Git; g != nil {
		logger.Info("cloning", "url", g.URL, "checkout", dfv1.PathCheckout)
		var auth transport.AuthMethod

		if k := g.UsernameSecret; k != nil {
			if v := g.PasswordSecret; v != nil {
				logger.Info("getting secret for auth", "UsernameSecret", k, "PasswordSecret", v)
				secretInterface := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie()).CoreV1().Secrets(os.Getenv(dfv1.EnvNamespace))

				if usernameSecret, err := secretInterface.Get(ctx, k.Name, metav1.GetOptions{}); err != nil {
					return fmt.Errorf("failed to get secret %q: %w", k.Name, err)
				} else {
					if pwdSecret, err := secretInterface.Get(ctx, v.Name, metav1.GetOptions{}); err != nil {
						return fmt.Errorf("failed to get secret %q: %w", v.Name, err)
					} else {
						auth = &http.BasicAuth{
							Username: string(usernameSecret.Data[v.Key]),
							Password: string(pwdSecret.Data[k.Key]),
						}
					}
				}
			}
		} else {
			logger.Info("No username & password found for git auth.")
		}

		if k := g.SSHPrivateKeySecret; k != nil {
			logger.Info("getting secret for auth", "SSHPrivateKeySecret", k)
			secretInterface := kubernetes.NewForConfigOrDie(ctrl.GetConfigOrDie()).CoreV1().Secrets(os.Getenv(dfv1.EnvNamespace))

			if sshPrivateKey, err := secretInterface.Get(ctx, k.Name, metav1.GetOptions{}); err != nil {
				return fmt.Errorf("failed to get secret %q: %w", k.Name, err)
			} else {
				if v, err := ssh.NewPublicKeys("git", sshPrivateKey.Data[k.Key], ""); err != nil {
					return fmt.Errorf("failed to get create public keys: %w", err)
				} else {
					auth = v
				}
			}
		} else {
			logger.Info("No ssh private key found for git auth.")
		}

		if _, err := git.PlainClone(dfv1.PathCheckout, false, &git.CloneOptions{
			Auth:          auth,
			Depth:         1, // checkout faster
			Progress:      os.Stdout,
			ReferenceName: plumbing.NewBranchReferenceName(g.Branch),
			SingleBranch:  true, // checkout faster
			URL:           g.URL,
		}); IgnoreErrRepositoryAlreadyExists(err) != nil {
			return fmt.Errorf("failed to clone repo: %w", err)
		}
		path := filepath.Join(dfv1.PathCheckout, g.Path)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}
		logger.Info("moving checked out code", "path", path, "wd", dfv1.PathWorkingDir)
		if err := os.Rename(path, dfv1.PathWorkingDir); sharedutil.IgnoreExist(err) != nil {
			return fmt.Errorf("failed to moved checked out path to working dir: %w", err)
		}
	} else if h := step.Spec.Code; h != nil {
		logger.Info("setting up handler", "runtime", h.Runtime, "code", strings.ShortenString(h.Source, 32)+"...")
		if err := ioutil.WriteFile(dfv1.PathHandlerFile, []byte(h.Source), 0o600); err != nil {
			return fmt.Errorf("failed to create code file: %w", err)
		}
	}
	return nil
}

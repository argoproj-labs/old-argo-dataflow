//go:build test
// +build test

package stress

import (
	"log"
	"os"

	"k8s.io/client-go/tools/clientcmd"
)

var currentContext string

func init() {
	home, _ := os.UserHomeDir()
	path := home + "/.kube/config"
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		path = ""
	}
	r, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{
		ExplicitPath: path,
	}, &clientcmd.ConfigOverrides{}).RawConfig()
	if err != nil {
		panic(err)
	}
	currentContext = r.CurrentContext
	log.Printf("currentContext=%s\n", currentContext)
}

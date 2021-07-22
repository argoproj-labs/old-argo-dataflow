// +build test

package stress

import (
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

var (
	currentContext string
)

func init() {
	home, _ := os.UserHomeDir()
	r, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&clientcmd.ClientConfigLoadingRules{
		ExplicitPath: home + "/.kube/config",
	}, &clientcmd.ConfigOverrides{}).RawConfig()
	if err != nil {
		panic(err)
	}
	currentContext = r.CurrentContext
	log.Printf("currentContext=%s\n", currentContext)
}

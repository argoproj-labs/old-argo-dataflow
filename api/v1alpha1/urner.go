package v1alpha1

// what a strange name - "URN" + "er"
// why? https://golang.org/doc/effective_go#interface-names
type urner interface { // only private to defeat codegen
	GenURN(cluster, namespace string) string
}

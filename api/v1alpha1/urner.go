package v1alpha1

import "context"

// what a strange name - "URN" + "er"
// why? https://golang.org/doc/effective_go#interface-names
type urner interface { // only private to defeat codegen
	GenURN(ctx context.Context) string
}

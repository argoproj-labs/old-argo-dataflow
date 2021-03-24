package v1alpha1

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/types"
)

type Message struct {
	ID   types.UID `json:"id"`
	Data []byte    `json:"data"`
}

func (in *Message) String() string {
	return fmt.Sprintf("(%s,%s)", in.ID, strconv.Quote(string(in.Data)))
}

package v1alpha1

import (
	"encoding/json"
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

func (in *Message) Json() string {
	data, _ := json.Marshal(in)
	return string(data)
}

func NewMessage(data []byte) Message {
	m := Message{}
	_ = json.Unmarshal(data, &m)
	return m
}

package golang

import "log"

func HandleCrash() {
	if r := recover(); r != nil {
		log.Printf("recovered from crash: %v\n", r)
	}
}

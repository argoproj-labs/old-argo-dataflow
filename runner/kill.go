package main

import "io/ioutil"

const (
	killFile = "/tmp/kill"
)

func Kill() error {
	return ioutil.WriteFile(killFile, nil, 0600)
}

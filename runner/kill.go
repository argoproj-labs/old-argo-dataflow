package main

import "io/ioutil"

func Kill() error {
	return ioutil.WriteFile(killFile, nil, 0600)
}

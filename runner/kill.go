package main

import "io/ioutil"

func killCmd() error {
	return ioutil.WriteFile(killFile, nil, 0600)
}

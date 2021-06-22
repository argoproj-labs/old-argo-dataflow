package test

import "log"

func WaitForever() {
	log.Printf("waiting forever\n")
	select {}
}

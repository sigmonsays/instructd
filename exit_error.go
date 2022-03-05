package main

import (
	"fmt"
	"os"
	"time"
)

func ExitError(s string, args ...interface{}) {
	fmt.Printf("\nERROR: "+s+"\n", args...)
	time.Sleep(time.Duration(1) * time.Second)
	os.Exit(1)
}

func ExitIfError(err error, s string, args ...interface{}) {
	if err == nil {
		return
	}
	ExitError(s, args...)
}

package main

import (
	"io/ioutil"
	"log"
	"strings"
)

const configFileName = "dir_list.cf"

var cfgCh = make(chan int) // channel for signal

// parse config file and return path array
func ConfigInit() []string {
	var dirs []string

	contents, err := ioutil.ReadFile(configFileName)
	if err != nil {
		log.Fatal(err.Error())
	}

	dirs = strings.Split(string(contents), "\n")

	return dirs
}

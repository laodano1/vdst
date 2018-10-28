package main

import (
	"io/ioutil"
	"log"
	"strings"
)

func ConfigInit() []string {
	contents, err := ioutil.ReadFile("dir_list.cf")
	if err != nil {
		log.Fatal(err.Error())
	}

	dirs := strings.Split(string(contents), "\n")

	return dirs
}

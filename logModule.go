package main

import (
	"time"
	"fmt"
	"os"
	"log"
)

func LogModuleInit() {
	year, month, day := time.Now().Date()
	logFileName := fmt.Sprintf("Server-%d-%d-%d.log", year, month, day)
	file, err := os.Create(logFileName)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.SetOutput(file)
}

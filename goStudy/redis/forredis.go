package  main

import (
	"github.com/garyburd/redigo/redis"
	"log"
	"time"
	"fmt"
)

func main() {
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Println("dial redis server error:", err)
	}
	defer c.Close()

	_, err = c.Do("SET", "passwd", "123456", "EX", "20")
	if err != nil {
		log.Println("set error:", err)
	}

	time.Sleep(2 * time.Second)

	password, err := redis.String(c.Do("GET", "password"))
	if err != nil {
		fmt.Println("redis get failed:", err)
	} else {
		fmt.Printf("Got password %v \n", password)
	}


}
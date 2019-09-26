package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"time"
)

type User struct {
	Name string
	Age  int
	Birthday time.Time
}

func main() {
	user := User{
		Name:     "TestUser",
		Age:      10,
		Birthday: time.Now(),
	}

	//"user:password@/dbname?charset=utf8&parseTime=True&loc=Local"
	db, err := gorm.Open("mysql", "root:root@/?charset=utf8&parseTime=True&loc=Local")
	//db, err := gorm.Open("mysql", "root:root@/testDB?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Println("open mysql error", err)
	}
	defer db.Close()

	// https://github.com/jinzhu/gorm/blob/master/main_test.go
	//
	log.Println("check & create db and table")

	//var result []struct{ string }
	//db.Exec("show databases").Scan(&result)
	//fmt.Printf("dbs: %v\n", result)

	fmt.Printf("database amount: %d\n", db.Exec("show databases").RowsAffected)
	//rs, err := db.Exec("show databases").Rows()
	//rs.Scan(&result)

	//if !strings.Contains(err.Error(), "database exists") {
	//	db.Exec("CREATE DATABASE testDB")
	//	log.Println("use testDB")
	//	db.Exec("USE testDB")
	//	db.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&User{})
	//} else {
		log.Println("use testDB")
		db.Exec("USE testDB")
	//}

	// add index
	//db.Model(&User{}).AddUniqueIndex("idx_user_name", "name")

	log.Println("add table record")
	db.Create(&user)

	//http://gorm.io/docs/query.html
	log.Println("select record")
	var arrUser []User
	db.Where("Name = ?", "TestUser").Find(&user).Scan(&arrUser)
	for k, item := range arrUser {
		fmt.Printf("id: %d | %s |\n", k, item.Birthday)
	}
	//fmt.Printf("users: %d \n", len(arr_user))
	log.Println("bye bye!")
}




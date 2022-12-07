package db

import (
	"fmt"

	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
)

func GormConnect() *gorm.DB {
	DBMS := "mysql"
	USER := "b5c7eb6d281e5d"
	PASS := "fb1ec773"
	PROTOCOL := "us-cdbr-east-06.cleardb.net"
	DBNAME := "heroku_d0a101037db1ece"
	CONNECT := DBMS + "://" + USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME + "?reconnect=true"

	db, err := gorm.Open(CONNECT)

	if err != nil {
		fmt.Println("データベース接続失敗")
	} else {
		fmt.Println("データベース接続成功")
	}
	return db
}

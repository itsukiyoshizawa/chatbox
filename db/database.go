package db

import (
	"fmt"

	"github.com/jinzhu/gorm"

	_ "github.com/go-sql-driver/mysql"
)

func GormConnect() *gorm.DB {
	DBMS := "mysql"
	USER := "baae2706f3809d"
	PASS := "41e4ceff"
	PROTOCOL := "us-cdbr-east-06.cleardb.net"
	DBNAME := "heroku_669397f69b32401"
	CONNECT := USER + ":" + PASS + "@tcp(" + PROTOCOL + ":3306)/" + DBNAME + "?parseTime=true"

	db, err := gorm.Open(DBMS, CONNECT)

	if err != nil {
		fmt.Println("データベース接続失敗")
	} else {
		fmt.Println("データベース接続成功")
	}
	return db
}

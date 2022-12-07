package db

import (
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type User struct {
	Id        int
	User_name string
	Message   string
	Password  string
}

type Post struct {
	Id        int
	User_name string
	Text      string
	Date      time.Time
	User_id   int
}

var LoginSession SessionInfo

type SessionInfo struct {
	User_name interface{}
}

// ログインチェック
func UserSelect(user_name string, password string) User {
	d := GormConnect()
	loginData := User{}
	d.First(&loginData, "user_name=? and password = ?", user_name, password)
	defer d.Close()
	return loginData
}

// ユーザー名からユーザー情報を取得
func UserRegister(user_name string) User {
	d := GormConnect()
	selData := User{}
	d.First(&selData, "user_name=?", user_name)
	defer d.Close()
	return selData
}

// ユーザーIDからユーザー情報を取得
func UserRegisterID(id int) User {
	d := GormConnect()
	selData := User{}
	d.First(&selData, "id=?", id)
	defer d.Close()
	return selData
}

//セッションにデータを格納する
func Login(ctx *gin.Context, User_name string) {
	session := sessions.Default(ctx)
	session.Set("User_name", User_name)
	session.Save()
	fmt.Println("セッション登録成功")
}

// ログアウト時のセッション削除
func Logout(ctx *gin.Context) {
	session := sessions.Default(ctx)
	fmt.Println("セッションを取得")
	session.Clear()
	session.Save()
}

// 新規投稿ユーザーが重複していないか確認
func UserCreate(user_name string) User {
	d := GormConnect()
	selData := User{}
	d.First(&selData, "user_name=?", user_name)
	defer d.Close()
	return selData
}

// 投稿IDから投稿情報を取得
func PostFound(id int) Post {
	d := GormConnect()
	selData := Post{}
	d.Table("posts").First(&selData, "id=?", id)
	defer d.Close()
	return selData
}

// 投稿IDから投稿情報を取得
func PostFoundOne(id int) Post {
	d := GormConnect()
	selData := Post{}
	d.Table("posts").Find(&selData, "id=?", id)
	defer d.Close()
	return selData
}

// ユーザーIDからユーザーの投稿一覧を取得
// func ChangeName(id int) Post {
// 	d := GormConnect()
// 	selData := Post{}
// 	d.Table("posts").Find(&selData, "user_id=?", id)
// 	defer d.Close()
// 	return selData
// }

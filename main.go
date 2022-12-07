package main

import (
	"fmt"
	"html/template"
	"main/db"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

// var jstLocation *time.Location
// var jstOnce sync.Once
// func jst() *time.Location {
// 	if jstLocation == nil {
// 		jstOnce.Do(func() {
// 			l, err := time.LoadLocation("Asia/Tokyo")
// 			if err != nil {
// 				l = time.FixedZone("JST2", +9*60*60)
// 			}
// 			jstLocation = l
// 		})
// 	}
// 	return jstLocation
// }
// t := time.Now().In(jst())

func nl2br(text string) template.HTML {
	return template.HTML(strings.Replace(template.HTMLEscapeString(text), "\n", "<br />", -1))
}

func main() {
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	router.SetFuncMap(template.FuncMap{
		"nl2br": nl2br,
	})

	router.LoadHTMLGlob("templates/*.html")
	// http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css/"))))
	router.Static("/css", "css/")

	var Err string
	var Errs string

	// ログインページ
	router.GET("/login", func(ctx *gin.Context) {
		Err = ""
		ctx.HTML(http.StatusOK, "login.html", gin.H{
			"Err": Err,
		})
	})

	// ログイン確認
	router.POST("/login", func(ctx *gin.Context) {
		var List []Post
		Db := db.GormConnect()
		Db.Table("posts").Order("id desc").Find(&List)

		User_name, _ := ctx.GetPostForm("user_name")
		Password, _ := ctx.GetPostForm("password")

		if User_name == "" || Password == "" {
			Err = "ユーザー名またはパスワードが入力されていません"
			ctx.HTML(http.StatusOK, "login.html", gin.H{
				"Err": Err,
			})
		} else {
			loginInfo := db.UserSelect(User_name, Password)

			if loginInfo.User_name != "" {
				db.Login(ctx, User_name)
				ctx.Redirect(http.StatusFound, "/")
			} else {
				Err = "ユーザー名またはパスワードが正しくありません"
				ctx.HTML(http.StatusOK, "login.html", gin.H{
					"Err": Err,
				})
			}
		}
	})

	// 一覧ページ
	router.GET("/", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		var List []Post
		Db := db.GormConnect()
		Db.Table("posts").Order("id desc").Find(&List)

		if LoginSession.User_name != nil {
			fmt.Println("セッション取得成功")
			userLog := LoginSession.User_name
			userinfo := db.UserRegister(userLog.(string))

			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"User":     LoginSession.User_name,
				"List":     List,
				"User_log": userinfo.Id,
			})
		} else {
			fmt.Println("セッション取得失敗")
			ctx.HTML(http.StatusOK, "index.html", gin.H{
				"List": List,
			})
		}
	})

	// 新規登録ページ
	router.GET("/registration", func(ctx *gin.Context) {
		Errs = ""

		ctx.HTML(http.StatusOK, "registration.html", gin.H{
			"Err": Errs,
		})
	})

	// 新規登録確認ページ
	router.POST("/registration", func(ctx *gin.Context) {
		User_name, _ := ctx.GetPostForm("user_name")
		Password, _ := ctx.GetPostForm("password")
		Len := len(Password)

		if User_name == "" || Password == "" {
			Errs := "ユーザー名またはパスワードが入力されていません"

			ctx.HTML(http.StatusOK, "registration.html", gin.H{
				"Err": Errs,
			})
		} else {
			userinfo := db.UserCreate(User_name)

			if userinfo.User_name != "" {
				Errs = "このユーザー名は既に作成してあります"
				ctx.HTML(http.StatusOK, "registration.html", gin.H{
					"Err": Errs,
				})
			} else {

				if User_name == "" || Len < 4 {
					Errs := "ユーザー名またはパスワードが正しくありません"

					ctx.HTML(http.StatusOK, "registration.html", gin.H{
						"Err": Errs,
					})
				} else {
					fmt.Println("登録成功")
					Db := db.GormConnect()
					user := User{
						User_name: User_name,
						Password:  Password,
					}

					Db.Create(&user)
					defer Db.Close()

					db.Login(ctx, User_name)
					ctx.Redirect(http.StatusFound, "/")
				}
			}
		}
	})

	// 投稿ページ
	router.GET("/create", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		if LoginSession.User_name != nil {
			user_name := LoginSession.User_name.(string)
			userinfo := db.UserRegister(user_name)

			ctx.HTML(http.StatusOK, "create.html", gin.H{
				"User_name": LoginSession.User_name,
				"User_log":  userinfo.Id,
			})
		} else {
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// 投稿作成ページ
	router.POST("/create", func(ctx *gin.Context) {
		Text, _ := ctx.GetPostForm("text")

		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")
		user_name := LoginSession.User_name.(string)
		userinfo := db.UserRegister(user_name)

		if Text == "" {
			Err := "投稿内容を入力してください"

			ctx.HTML(http.StatusOK, "create.html", gin.H{
				"User_name": LoginSession.User_name,
				"User_log":  userinfo.Id,
				"Err":       Err,
			})
		} else {

			if userinfo.User_name != "" {
				db := db.GormConnect()

				t := time.Now().UTC()
				tokyo, err := time.LoadLocation("Asia/Tokyo")
				if err != nil {
					fmt.Println("時刻取得失敗")
				}
				Time := t.In(tokyo)

				post := Post{
					User_name: userinfo.User_name,
					Text:      Text,
					Date:      Time,
					User_id:   userinfo.Id,
				}
				db.Create(&post)
			} else {
				Err := "投稿が失敗しました"

				ctx.HTML(http.StatusOK, "create.html", gin.H{
					"User_name": LoginSession.User_name,
					"User_log":  userinfo.Id,
					"Err":       Err,
				})
			}
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// ログアウトページ
	router.GET("/logout", func(ctx *gin.Context) {
		db.Logout(ctx)
		var List []Post
		db := db.GormConnect()
		db.Table("posts").Order("id desc").Find(&List)
		ctx.Redirect(http.StatusFound, "/")
	})

	// 投稿編集ページ
	router.GET("/edit", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		id := ctx.Query("id")
		ids, _ := strconv.Atoi(id)
		PostInfo := db.PostFound(ids)

		if LoginSession.User_name != nil {
			user_name := LoginSession.User_name.(string)
			Userinfo := db.UserRegister(user_name)

			if PostInfo.User_name == Userinfo.User_name {
				Err := ""

				ctx.HTML(http.StatusOK, "edit.html", gin.H{
					"User_name": LoginSession.User_name,
					"Text":      PostInfo.Text,
					"Id":        PostInfo.Id,
					"User":      Userinfo.User_name,
					"Err":       Err,
				})
			} else {
				ctx.Redirect(http.StatusFound, "/")

			}
		} else {
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// 投稿編集登録画面
	router.POST("/edit", func(ctx *gin.Context) {
		Text, _ := ctx.GetPostForm("text")
		Id, _ := ctx.GetPostForm("id")

		Db := db.GormConnect()
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")
		user_name := LoginSession.User_name.(string)
		Users := db.UserRegister(user_name)

		ids, _ := strconv.Atoi(Id)
		userinfo := db.PostFound(ids)

		if Text == "" {
			Err := "投稿内容が入力されていません"

			ctx.HTML(http.StatusOK, "edit.html", gin.H{
				"Err":      Err,
				"User":     Users.User_name,
				"User_log": Users.Id,
				"Id":       userinfo.Id,
			})
		} else {

			if LoginSession.User_name != nil {
				userinfo.Text = Text
				Db.Save(&userinfo)

				ctx.Redirect(http.StatusFound, "/")
			} else {
				Err := "ユーザーが一致しません"

				ctx.HTML(http.StatusOK, "edit.html", gin.H{
					"Err": Err,
				})
			}
		}
	})

	// 投稿削除ページ
	router.GET("/delete", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		id := ctx.Query("id")
		ids, _ := strconv.Atoi(id)
		PostInfo := db.PostFound(ids)

		if LoginSession.User_name != nil {
			user_name := LoginSession.User_name.(string)
			Userinfo := db.UserRegister(user_name)

			if PostInfo.User_name == LoginSession.User_name {
				Err := ""

				ctx.HTML(http.StatusOK, "delete.html", gin.H{
					"User_name": LoginSession.User_name,
					"Text":      PostInfo.Text,
					"Id":        PostInfo.Id,
					"Date":      PostInfo.Date,
					"User":      Userinfo.User_name,
					"Err":       Err,
				})
			} else {
				ctx.Redirect(http.StatusFound, "/")
			}
		} else {
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// 投稿削除登録画面
	router.POST("/delete", func(ctx *gin.Context) {
		Id, _ := ctx.GetPostForm("id")

		Db := db.GormConnect()
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		ids, _ := strconv.Atoi(Id)
		userinfo := db.PostFound(ids)

		if LoginSession.User_name != nil {
			Db.Delete(&userinfo)

			ctx.Redirect(http.StatusFound, "/")
		} else {
			Err := "投稿の削除に失敗しました"

			ctx.HTML(http.StatusOK, "delete.html", gin.H{
				"Err": Err,
			})
		}
	})

	// プロフィールページ
	router.GET("/profile", func(ctx *gin.Context) {
		Id := ctx.Query("id")
		ids, _ := strconv.Atoi(Id)
		PostInfo := db.UserRegisterID(ids)

		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		Db := db.GormConnect()

		if PostInfo.User_name != "" {

			if LoginSession.User_name != nil {
				user_name := LoginSession.User_name.(string)
				Userinfo := db.UserRegister(user_name)

				var List []Post
				Db.Table("posts").Order("id desc").Where("user_id=?", ids).Find(&List)

				ctx.HTML(http.StatusOK, "profile.html", gin.H{
					"User_name": PostInfo.User_name,
					"Post_log":  PostInfo.Id,
					"User_log":  Userinfo.Id,
					"List":      List,
					"Message":   PostInfo.Message,
					"User":      Userinfo.User_name,
				})
			} else {
				var List []Post
				Db.Table("posts").Order("id desc").Where("user_id=?", ids).Find(&List)

				ctx.HTML(http.StatusOK, "profile.html", gin.H{
					"List":      List,
					"User_name": PostInfo.User_name,
					"Post_log":  PostInfo.Id,
					"Message":   PostInfo.Message,
				})
			}
		} else {

			if Id != "" {
				Err := "ユーザーが存在しません"
				Nil := ""

				ctx.HTML(http.StatusOK, "profile.html", gin.H{
					"Err":       Err,
					"User_name": Nil,
				})
			} else {

				if LoginSession.User_name != nil {
					user_name := LoginSession.User_name.(string)
					Userinfo := db.UserRegister(user_name)

					var List []Post
					Db.Table("posts").Order("id desc").Where("user_id=?", Userinfo.Id).Find(&List)

					ctx.HTML(http.StatusOK, "profile.html", gin.H{
						"User_name": Userinfo.User_name,
						"Post_log":  Userinfo.Id,
						"User_log":  Userinfo.Id,
						"List":      List,
						"Message":   Userinfo.Message,
						"User":      Userinfo.User_name,
					})
				}
			}
		}

	})

	// プロフィール編集ページ
	router.GET("/profile_edit", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		id := ctx.Query("id")
		ids, _ := strconv.Atoi(id)
		PostInfo := db.UserRegisterID(ids)

		if PostInfo.User_name == LoginSession.User_name {
			user_name := LoginSession.User_name.(string)
			Userinfo := db.UserRegister(user_name)
			ctx.HTML(http.StatusOK, "profile_edit.html", gin.H{
				"User_name": PostInfo.User_name,
				"Message":   PostInfo.Message,
				"Id":        PostInfo.Id,
				"User":      Userinfo.User_name,
			})
		} else {
			fmt.Println("プロフィールの編集に失敗しました")
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// プロフィール編集ページ
	router.POST("/profile_edit", func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		Message, _ := ctx.GetPostForm("message")
		Id, _ := ctx.GetPostForm("id")
		Ids, _ := strconv.Atoi(Id)

		Db := db.GormConnect()

		userinfo := db.UserRegisterID(Ids)

		if userinfo.User_name != "" {
			userinfo.Message = Message
			Db.Save(&userinfo)

			ctx.Redirect(http.StatusFound, "/profile")
		} else {
			ctx.Redirect(http.StatusFound, "/")
		}
	})

	// 個別記事ページ
	router.GET("/post", func(ctx *gin.Context) {
		id := ctx.Query("id")
		ids, _ := strconv.Atoi(id)
		PostInfo := db.PostFoundOne(ids)

		session := sessions.Default(ctx)
		LoginSession.User_name = session.Get("User_name")

		if PostInfo.User_name != "" {
			if LoginSession.User_name != nil {
				user_name := LoginSession.User_name.(string)
				Userinfo := db.UserRegister(user_name)
				ctx.HTML(http.StatusOK, "post.html", gin.H{
					"Id":        PostInfo.Id,
					"User_name": PostInfo.User_name,
					"Date":      PostInfo.Date,
					"Text":      PostInfo.Text,
					"User_log":  Userinfo.Id,
					"User_id":   PostInfo.User_id,
					"User":      Userinfo.User_name,
				})
			} else {
				ctx.HTML(http.StatusOK, "post.html", gin.H{
					"Id":        PostInfo.Id,
					"User_name": PostInfo.User_name,
					"Date":      PostInfo.Date,
					"Text":      PostInfo.Text,
					"User_id":   PostInfo.User_id,
				})
			}
		} else {
			Err := "この投稿は存在しないか削除されました"
			Nil := ""

			ctx.HTML(http.StatusOK, "post.html", gin.H{
				"Err":       Err,
				"User_name": Nil,
			})
		}
	})

	// router.Run(":8080")
	router.Run(":" + os.Getenv("PORT"))
}

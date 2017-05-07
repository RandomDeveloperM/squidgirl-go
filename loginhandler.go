package main

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"

	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

const (
	//TokenLimitHour トークンの有効時間
	TokenLimitHour = 48
)

//LoginRequest ログインリクエストデータ
type LoginRequest struct {
	UserName string `json:"username" xml:"username" form:"username" query:"username"`
	Password string `json:"password" xml:"password" form:"password" query:"password"`
}

//LoginResponce ログインレスポンスデータ
type LoginResponce struct {
	Token string
}

//LoginHandler ユーザーログインハンドラ
func LoginHandler(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	token, err := login(req.UserName, req.Password)
	if err != nil {
		return err
	}

	res := new(LoginResponce)
	res.Token = token
	return c.JSON(http.StatusOK, res)
}

//ログインチェックを行い、トークンを返す
func login(userName string, password string) (string, error) {
	//DB検索を行う
	user, err := db.SelectUser(userName)
	if err != nil {
		fmt.Printf("login ERROR ユーザー検索エラー\n")
		return "", err
	}
	if user.Name != userName {
		fmt.Printf("login ERROR ユーザー未登録\n")
		return "", nil
	}

	passHash := db.CreatePasswordHash(password)
	if user.PassHash != passHash {
		fmt.Printf("login ERROR パスワード不正 %s <> %s\n", user.PassHash, passHash)
		return "", nil
	}

	token, err := createToken(userName, user.Permission)
	if err != nil {
		return "", nil
	}

	fmt.Printf("login OK\n")
	return token, nil
}

//ログイントークンを生成して返す
func createToken(userName string, permission int) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	//マップに設定
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = userName
	claims["admin"] = permission == db.UserPermissionAdmin
	claims["exp"] = time.Now().Add(time.Hour * TokenLimitHour).Unix()

	//トークン取得
	t, err := token.SignedString([]byte(config.GetConfig().Login.TokenSalt))
	if err != nil {
		return "", err
	}

	return t, nil
}

//CreateUserHandler ユーザーを作成する
func CreateUserHandler(c echo.Context) error {
	err := createUser("test", "testpassword", db.UserPermissionAdmin)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, "OK")
}

//指定したユーザー名・パスワードでユーザーを作成する。既に作成されているときは更新する
func createUser(userName string, password string, permission int) error {
	if err := db.InsertUser(userName, password, permission); err != nil {
		return err
	}

	return nil
}

//GetLoginUserName ログイントークンからユーザー名を取得する
func GetLoginUserName(c echo.Context) string {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	return claims["name"].(string)
}

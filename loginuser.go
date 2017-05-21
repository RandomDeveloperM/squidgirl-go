package main

import (
	"time"

	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

const (
	//TokenLimitHour トークンの有効時間
	tokenLimitHour = 48
)

//LoginUser はログインユーザーのログイン情報を保持する
type LoginUser struct {
	UserName  string
	AuthLevel int
}

//NewLoginUserFromRequest はechoのJWTリクエストヘッダ情報からログイン情報を取得して返す
func NewLoginUserFromRequest(c echo.Context) *LoginUser {
	tokenUser := c.Get("user").(*jwt.Token)
	claims := tokenUser.Claims.(jwt.MapClaims)

	loginUser := new(LoginUser)
	loginUser.UserName = claims["name"].(string)
	loginUser.AuthLevel = claims["authlevel"].(int)
	return loginUser
}

//NewLoginUserFromDB は指定したユーザー情報からDBを検索してログイン情報を取得して返す
func NewLoginUserFromDB(userName string, password string) (*LoginUser, error) {
	if userName == "" || password == "" {
		return nil, fmt.Errorf("ユーザー名またはパスワード入力なし")
	}

	user, err := db.SelectUser(userName)
	if err != nil {
		return nil, fmt.Errorf("指定されたユーザー名が見つからない")
	}

	passHash := db.CreatePasswordHash(password)
	if user.PassHash != passHash {
		return nil, fmt.Errorf("パスワードが違う")
	}

	loginUser := new(LoginUser)
	loginUser.UserName = userName
	loginUser.AuthLevel = user.Permission
	return loginUser, nil
}

//CreateLoginToken 現在のログイン情報からログイントークを生成して返す
func (loginUser *LoginUser) CreateLoginToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	//マップに設定
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = loginUser.UserName
	claims["authlevel"] = loginUser.AuthLevel
	claims["exp"] = time.Now().Add(time.Hour * tokenLimitHour).Unix()

	//トークン取得
	t, err := token.SignedString([]byte(config.GetConfig().Login.TokenSalt))
	if err != nil {
		return "", err
	}

	return t, nil
}

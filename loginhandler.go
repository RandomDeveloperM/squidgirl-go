package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"github.com/mryp/squidgirl-go/db"
)

//LoginRequest ログインリクエストデータ
type LoginRequest struct {
	UserName string `json:"username" xml:"username" form:"username" query:"username"`
	Password string `json:"password" xml:"password" form:"password" query:"password"`
}

//LoginResponce ログインレスポンスデータ
type LoginResponce struct {
	Token string `json:"token" xml:"token"`
}

type CreateUserRequest struct {
	UserName  string `json:"username" xml:"username" form:"username" query:"username"`
	Password  string `json:"password" xml:"password" form:"password" query:"password"`
	AuthLevel int    `json:"authlevel" xml:"authlevel" form:"authlevel" query:"authlevel"`
}

type CreateUserResponce struct {
	Status int `json:"status" xml:"status"`
}

type DeleteUserRequest struct {
	UserName string `json:"username" xml:"username" form:"username" query:"username"`
}

type DeleteUserResponce struct {
	Status int `json:"status" xml:"status"`
}

type UserListResponce struct {
	Count int                     `json:"count" xml:"count"`
	Users []UserListUsersResponce `json:"users" xml:"users"`
}

type UserListUsersResponce struct {
	UserName  string `json:"username" xml:"username"`
	AuthLevel int    `json:"authlevel" xml:"authlevel"`
}

//LoginHandler ユーザーログインハンドラ
func LoginHandler(c echo.Context) error {
	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	loginUser, err := NewLoginUserFromDB(req.UserName, req.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	token, err := loginUser.CreateLoginToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	res := new(LoginResponce)
	res.Token = token
	return c.JSON(http.StatusOK, res)
}

//CreateUserHandler ユーザーを作成する
func CreateUserHandler(c echo.Context) error {
	req := new(CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	err := db.InsertUser(req.UserName, req.Password, req.AuthLevel)
	if err != nil {
		return err
	}

	res := new(CreateUserResponce)
	res.Status = 0
	return c.JSON(http.StatusOK, res)
}

func DeleteUserHandler(c echo.Context) error {
	req := new(DeleteUserRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	res := new(DeleteUserResponce)
	res.Status = 0
	return c.JSON(http.StatusOK, res)
}

func UserListHandler(c echo.Context) error {
	res := new(UserListResponce)
	res.Count = 0
	return c.JSON(http.StatusOK, res)
}

//指定したユーザー名・パスワードでユーザーを作成する。既に作成されているときは更新する
func createUser(userName string, password string, permission int) error {
	if err := db.InsertUser(userName, password, permission); err != nil {
		return err
	}

	return nil
}

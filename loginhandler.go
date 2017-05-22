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

	fmt.Printf("token=%s\n", token)
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

	user, err := db.SelectUser(req.UserName)
	if err == nil && user.ID != 0 {
		return fmt.Errorf("すでにユーザーが存在する")
	}

	err = db.InsertUser(req.UserName, req.Password, req.AuthLevel)
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

	user, err := db.SelectUser(req.UserName)
	if err != nil || user.ID == 0 {
		return fmt.Errorf("削除するユーザーが見つからない")
	}

	//ユーザーを削除
	err = db.DeleteUser(user.ID)
	if err != nil {
		return err
	}

	//履歴も削除
	err = db.DeleteHistory(user.Name)
	if err != nil {
		return err
	}

	res := new(DeleteUserResponce)
	res.Status = 0
	return c.JSON(http.StatusOK, res)
}

func UserListHandler(c echo.Context) error {
	userList, err := db.SelectUserAll()
	if err != nil {
		return err
	}

	res := new(UserListResponce)
	res.Count = len(userList)
	userResponceList := make([]UserListUsersResponce, 0)
	for _, user := range userList {
		userResponceList = append(userResponceList, UserListUsersResponce{
			UserName:  user.Name,
			AuthLevel: user.Permission,
		})
	}
	res.Users = userResponceList
	return c.JSON(http.StatusOK, res)
}

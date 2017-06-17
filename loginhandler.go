package main

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"

	"github.com/mryp/squidgirl-go/db"
)

//LoginRequest はログインリクエストデータ構造体
type LoginRequest struct {
	UserName string `json:"username" xml:"username" form:"username" query:"username"`
	Password string `json:"password" xml:"password" form:"password" query:"password"`
}

//LoginResponce はログインレスポンスデータ構造体
type LoginResponce struct {
	Token string `json:"token" xml:"token"`
}

//CreateUserRequest はユーザー作成リクエストデータ構造体
type CreateUserRequest struct {
	UserName  string `json:"username" xml:"username" form:"username" query:"username"`
	Password  string `json:"password" xml:"password" form:"password" query:"password"`
	AuthLevel int    `json:"authlevel" xml:"authlevel" form:"authlevel" query:"authlevel"`
}

//CreateUserResponce はユーザー作成レスポンス構造体
type CreateUserResponce struct {
	Status int `json:"status" xml:"status"`
}

//DeleteUserRequest はユーザー削除リクエスト構造体
type DeleteUserRequest struct {
	UserName string `json:"username" xml:"username" form:"username" query:"username"`
}

//DeleteUserResponce はユーザー削除レスポンス構造体
type DeleteUserResponce struct {
	Status int `json:"status" xml:"status"`
}

//UserListResponce はユーザー一覧取得リクエスト構造体
type UserListResponce struct {
	Count int                     `json:"count" xml:"count"`
	Users []UserListUsersResponce `json:"users" xml:"users"`
}

//UserListUsersResponce はユーザー一覧取得レスポンス構造体
type UserListUsersResponce struct {
	UserName  string `json:"username" xml:"username"`
	AuthLevel int    `json:"authlevel" xml:"authlevel"`
}

//LoginHandler はユーザーログインを行い、トークンを返す
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

//CreateUserHandler は指定したユーザーを作成する
func CreateUserHandler(c echo.Context) error {
	req := new(CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	loginUser := NewLoginUserFromRequest(c)
	if loginUser.AuthLevel != db.UserPermissionAdmin {
		return fmt.Errorf("ログインユーザーが管理者権限を持っていない")
	}

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

//DeleteUserHandler は指定したユーザーを削除する。削除するには管理者権限が必要
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

	//トークンからユーザー名を取得
	loginUser := NewLoginUserFromRequest(c)
	if loginUser.AuthLevel != db.UserPermissionAdmin {
		return fmt.Errorf("ログインユーザーが管理者権限を持っていない")
	}
	if loginUser.UserName == req.UserName {
		return fmt.Errorf("現在ログインしているユーザーは削除できない")
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

//UserListHandler はユーザー一覧を取得する
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

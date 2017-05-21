package main

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/mryp/squidgirl-go/db"
)

//SaveBookRequest はアーカイブ情報を保存するリクエストデータを保持する
type SaveBookRequest struct {
	Hash     string `json:"hash" xml:"hash" form:"hash" query:"hash"`
	Index    int    `json:"index" xml:"index" form:"index" query:"index"`
	Reqction int    `json:"reaction" xml:"reaction" form:"reaction" query:"reaction"`
}

//SaveBookResponse はアーカイブ情報を保持する処理のレスポンスデータを保持する
type SaveBookResponse struct {
	Status int `json:"status" xml:"status"`
}

//SaveBookHandler はアーカイブ情報を保存してレスポンスを返す
func SaveBookHandler(c echo.Context) error {
	req := new(SaveBookRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("SaveBookHandler request=%v\n", *req)

	//トークンからユーザー名を取得
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userName := claims["name"].(string)

	//データの追加
	err := db.InsertHistory(userName, req.Hash, req.Index, req.Reqction, true)
	if err != nil {
		return err
	}

	response := new(SaveBookResponse)
	response.Status = 0
	return c.JSON(http.StatusOK, response)
}

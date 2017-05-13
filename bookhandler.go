package main

import (
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/mryp/squidgirl-go/db"
)

type SaveBookRequest struct {
	Hash     string `json:"hash" xml:"hash" form:"hash" query:"hash"`
	Index    int    `json:"index" xml:"index" form:"index" query:"index"`
	Reqction int    `json:"reaction" xml:"reaction" form:"reaction" query:"reaction"`
}

type SaveBookResponse struct {
	Status int `json:"status" xml:"status"`
}

func SaveBookHandler(c echo.Context) error {
	req := new(SaveBookRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	//トークンからユーザー名を取得
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userName := claims["name"].(string)

	//データの追加
	history, _ := db.SelectHistory(userName, req.Hash)
	if history.BookHash == "" {
		err := db.InsertHistory(userName, req.Hash, req.Index, req.Reqction)
		if err != nil {
			return err
		}
	} else {
		err := db.UpdateHistory(userName, req.Hash, req.Index, req.Reqction)
		if err != nil {
			return err
		}
	}

	response := new(SaveBookResponse)
	response.Status = 0
	return c.JSON(http.StatusOK, response)
}

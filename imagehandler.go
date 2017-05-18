package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

const (
	GetImageCacheCount = 3
)

type ThumbnailRequest struct {
	Base64 bool `json:"base64" xml:"base64" form:"base64" query:"base64"`
}

type PageRequest struct {
	Index     int  `json:"index" xml:"index" form:"index" query:"index"`
	MaxHeight uint `json:"maxheight" xml:"maxheight" form:"maxheight" query:"maxheight"`
	MaxWidth  uint `json:"maxwidth" xml:"maxwidth" form:"maxwidth" query:"maxwidth"`
	Base64    bool `json:"base64" xml:"base64" form:"base64" query:"base64"`
}

func ThumbnailHandler(c echo.Context) error {
	hash := c.Param("hash")
	fmt.Printf("hash=%s\n", hash)

	req := new(ThumbnailRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	thum := NewThumbnail()
	thumImagePath := thum.GetFilePathFromHash(hash)
	_, err := os.Stat(thumImagePath)
	if os.IsNotExist(err) {
		//画像なしを返却する
		if req.Base64 {
			return c.String(http.StatusOK, "")
		}
		return c.File("assets/noimage.jpg")
	}

	if req.Base64 {
		imageBase64, err := convertImageToBase64(thumImagePath)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, imageBase64)
	}

	return c.File(thumImagePath)
}

func convertImageToBase64(filePath string) (string, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func PageHandler(c echo.Context) error {
	hash := c.Param("hash")
	fmt.Printf("hash=%s\n", hash)

	req := new(PageRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	exist, filePath := IsExistPageFile(hash, req.Index, req.MaxHeight, req.MaxWidth)
	if filePath == "" {
		return c.NoContent(http.StatusBadRequest)
	} else if !exist {
		//ZIPから対象のファイルを作成する
		_, err := UnzipPageFile(hash, req.Index, 1, req.MaxHeight, req.MaxWidth)
		if err != nil {
			return err
		}
	}

	//トークンからユーザー名を取得
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userName := claims["name"].(string)

	//現在の読み込み位置を保存
	err := updateHistory(userName, hash, req.Index, -1)
	if err != nil {
		fmt.Printf("PageHandler 1 err=%s\n", err)
		return err
	}

	//次のページ以降の展開キャッシュを行う（非同期）
	go UnzipPageFileMutex(hash, req.Index+1, GetImageCacheCount, req.MaxHeight, req.MaxWidth)

	//データを返却
	if req.Base64 {
		imageBase64, err := convertImageToBase64(filePath)
		if err != nil {
			fmt.Printf("PageHandler 2 err=%s\n", err)
			return err
		}
		return c.String(http.StatusOK, imageBase64)
	}

	return c.File(filePath)
}

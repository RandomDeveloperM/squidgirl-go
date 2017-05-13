package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"os"

	"github.com/labstack/echo"
)

type ThumbnailRequest struct {
	Hash string `json:"hash" xml:"hash" form:"hash" query:"hash"`
}

type PageRequest struct {
	Index     int  `json:"index" xml:"index" form:"index" query:"index"`
	MaxHeight int  `json:"maxheight" xml:"maxheight" form:"maxheight" query:"maxheight"`
	MaxWidth  int  `json:"maxwidth" xml:"maxwidth" form:"maxwidth" query:"maxwidth"`
	Base64    bool `json:"base64" xml:"base64" form:"base64" query:"base64"`
}

func ThumbnailHandler(c echo.Context) error {
	hash := c.Param("hash")
	fmt.Printf("hash=%s\n", hash)

	thumImagePath := CreateThumFilePathFromHash(hash)
	_, err := os.Stat(thumImagePath)
	if err != nil {
		return c.File("assets/noimage.jpg")
	}
	return c.File(thumImagePath)
}

func ThumbnailBase64Handler(c echo.Context) error {
	hash := c.Param("hash")
	fmt.Printf("hash=%s\n", hash)

	thumImagePath := CreateThumFilePathFromHash(hash)
	_, err := os.Stat(thumImagePath)
	if os.IsNotExist(err) {
		//画像なしを返却する
		return c.String(http.StatusOK, "")
	}

	imageBase64, err := convertImageToBase64(thumImagePath)
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, imageBase64)
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

	filePath, err := CreatePageFilePathFromHash(hash, req.Index, req.MaxHeight, req.MaxWidth)
	if err != nil {
		return err
	}

	if req.Base64 {
		imageBase64, err := convertImageToBase64(filePath)
		if err != nil {
			return err
		}
		return c.String(http.StatusOK, imageBase64)
	}

	return c.File(filePath)
}

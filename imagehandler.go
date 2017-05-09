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
	if err != nil {
		thumImagePath = "assets/noimage.jpg"
	}

	data, err := ioutil.ReadFile(thumImagePath)
	if err != nil {
		return err
	}
	imageBase64 := base64.StdEncoding.EncodeToString(data)
	return c.String(http.StatusOK, imageBase64)
}

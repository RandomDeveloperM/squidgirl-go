package main

import (
	"fmt"

	"github.com/labstack/echo"
)

type ThumbnailRequest struct {
	Hash string `json:"hash" xml:"hash" form:"hash" query:"hash"`
}

func ThumbnailHandler(c echo.Context) error {
	req := new(FileListRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	thumImagePath := CreateThumFilePathFromHash(req.Hash)
	return c.File(thumImagePath)
}

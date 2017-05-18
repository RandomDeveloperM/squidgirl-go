package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mryp/squidgirl-go/db"
)

const (
	defThumbnailDirPath     = "_temp/thumbnail"
	defThumbnailWidth       = 512
	defThumbnailJpegQuality = 70
)

type Thumbnail struct {
	dirPath     string
	width       uint
	jpegQuality int
}

type ThumbnailFunc interface {
	GetFilePathFromHash(hash string)
	GetFilePath(bookPath string)
	IsExist(filePath string)
	CreateFile(bookPath string)
}

func NewThumbnail() *Thumbnail {
	thum := new(Thumbnail)
	thum.dirPath = defThumbnailDirPath
	thum.width = defThumbnailWidth
	thum.jpegQuality = defThumbnailJpegQuality
	return thum
}

func (thum *Thumbnail) GetFilePathFromHash(hash string) string {
	return filepath.Join(thum.dirPath, hash+".jpg")
}

func (thum *Thumbnail) GetFilePath(bookPath string) string {
	return thum.GetFilePathFromHash(db.CreateBookHash(bookPath))
}

func (thum *Thumbnail) IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func (thum *Thumbnail) CreateFile(bookPath string) error {
	//ZIPファイルを開く
	r, err := zip.OpenReader(bookPath)
	if err != nil {
		fmt.Printf("ZIPファイルオープンエラー err:%s\n", err)
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		//ZIPファイル内のファイルを開く
		rc, err := f.Open()
		if err != nil {
			fmt.Printf("ZIP内ファイルオープンエラー err:%s\n", err)
			continue
		}
		defer rc.Close()

		if !f.FileInfo().IsDir() {
			//最初のページファイルをサムネイル画像として作成する
			resize := NewResize(0, thum.width, thum.jpegQuality)
			resize.ResizeFile(rc, thum.GetFilePath(bookPath))
			break
		}
	}

	return nil
}

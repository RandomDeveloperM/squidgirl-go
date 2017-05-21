package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

//Thumbnail はサムネイル変換情報を保持する
type Thumbnail struct {
	dirPath     string
	width       uint
	jpegQuality int
}

//NewThumbnail はサムネイル構造体をデフォルト値をセットして返す
func NewThumbnail() *Thumbnail {
	thum := new(Thumbnail)
	thum.dirPath = config.GetConfig().File.ThumbnailDirPath
	thum.width = uint(config.GetConfig().File.ThumbnailWidth)
	thum.jpegQuality = config.GetConfig().File.ThumbnailJpegQuality
	return thum
}

//GetFilePathFromHash はアーカイブハッシュからサムネイルのファイルパスを取得する
func (thum *Thumbnail) GetFilePathFromHash(hash string) string {
	return filepath.Join(thum.dirPath, hash+".jpg")
}

//GetFilePath はアーカイブのファイルパスからサムネイルのファイルパスを取得する
func (thum *Thumbnail) GetFilePath(bookPath string) string {
	return thum.GetFilePathFromHash(db.CreateBookHash(bookPath))
}

//IsExist は指定したサムネイルファイルが存在するかどうかを返す
func (thum *Thumbnail) IsExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

//CreateFile はアーカイブファイルの先頭ファイル画像をサムネイル画像として保存する
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

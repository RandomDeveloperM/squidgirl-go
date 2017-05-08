package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"

	"io"

	_ "github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

const (
	ThumbnailDirPath     = "_temp/thumbnail"
	ThumbnailWidth       = 512
	ThumbnailJpegQuality = 70
)

func CreateThumbnailFile(filePath string) error {
	fmt.Printf("CreateThumbnailFile filePath=%s\n", filePath)
	return createThumbnailFileFromZip(filePath)
}

func GetArchivePageCount(filePath string) (int, error) {
	fmt.Printf("GetArchivePageCount filePath=%s\n", filePath)
	return getZipFileCount(filePath)
}

func createThumbnailFileFromZip(filePath string) error {
	//ZIPファイルを開く
	r, err := zip.OpenReader(filePath)
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
			saveResizeImage(rc, ThumbnailWidth, 0, ThumbnailJpegQuality, CreateThumFilePath(filePath))
			break
		}
	}

	return nil
}

func saveResizeImage(r io.Reader, width uint, height uint, jpegQuality int, outputPath string) error {
	//画像読み込み
	image, _, err := image.Decode(r)
	if err != nil {
		fmt.Printf("画像読み込みエラー err:%s\n", err)
		return err
	}
	resizeImage := resize.Resize(width, height, image, resize.Lanczos3)

	//書き込み用ファイル作成
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("ファイル作成エラー err:%s\n", err)
		return err
	}
	defer outFile.Close()

	//JPEGとして保存
	opts := &jpeg.Options{Quality: jpegQuality}
	jpeg.Encode(outFile, resizeImage, opts)
	return nil
}

func CreateThumFilePath(filePath string) string {
	return CreateThumFilePathFromHash(db.CreateBookHash(filePath))
}

func CreateThumFilePathFromHash(hash string) string {
	return filepath.Join(ThumbnailDirPath, hash+".jpg")
}

//ZIPファイル内のファイル数を取得する
func getZipFileCount(filePath string) (int, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		fmt.Printf("getZipFileCount err=%s\n", err)
		return 0, err
	}
	defer r.Close()

	count := len(r.File)
	return count, nil
}

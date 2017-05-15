package main

import (
	"archive/zip"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nfnt/resize"

	"io"

	"github.com/mryp/squidgirl-go/db"
)

const (
	ThumbnailDirPath     = "_temp/thumbnail"
	ThumbnailWidth       = 512
	ThumbnailJpegQuality = 70
	PageDirPath          = "_temp/cache"
	PageJpegQuality      = 70
)

var (
	unzipMutex     *sync.Mutex = nil
	unzipBusyParam             = ""
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

func CreatePageFile(hash string, index int, maxHeight int, maxWidth int) (string, error) {
	bookRecord, err := db.SelectBookFromHash(hash)
	if err != nil {
		fmt.Printf("CreatePageFile err=%s\n", err)
		return "", err
	}

	zipFilePath := bookRecord.FilePath
	imageFilePath, err := createPageFileFromZip(zipFilePath, index, maxHeight, maxWidth)
	if err != nil {
		fmt.Printf("CreatePageFile err=%s\n", err)
		return "", err
	}

	return imageFilePath, nil
}

func createPageFileFromZip(filePath string, index int, maxHeight int, maxWidth int) (string, error) {
	height, width := convertResizeImageSize(maxHeight, maxWidth)

	//キャッシュされているかどうか確認
	outputPath := createPageFilePath(filePath, index, height, width)
	_, err := os.Stat(outputPath)
	if !os.IsNotExist(err) {
		fmt.Printf("ファイルがキャッシュに見つかった path=%s\n", outputPath)
		return outputPath, nil
	}

	//ZIPファイルを開く
	r, err := zip.OpenReader(filePath)
	if err != nil {
		fmt.Printf("ZIPファイルオープンエラー err:%s\n", err)
		return "", err
	}
	defer r.Close()

	//ページのファイルを検索
	if len(r.File) < index {
		return "", fmt.Errorf("対象ページなし")
	}
	zipFile := r.File[index]
	if zipFile.FileInfo().IsDir() {
		return "", fmt.Errorf("対象ページがフォルダ")
	}

	//ページ画像ファイル取得
	rc, err := zipFile.Open()
	if err != nil {
		fmt.Printf("ZIP内ファイルオープンエラー err:%s\n", err)
		return "", err
	}
	defer rc.Close()

	//指定したサイズに縮小
	saveResizeImage(rc, width, height, PageJpegQuality, outputPath)
	return outputPath, nil
}

func convertResizeImageSize(maxHeight int, maxWidth int) (uint, uint) {
	width := uint(maxWidth)
	height := uint(maxHeight)
	if maxWidth > maxHeight {
		width = 0 //横幅は自動
	} else if maxHeight > maxWidth {
		height = 0 //高さは自動
	}

	return height, width
}

func createPageFilePath(filePath string, index int, height uint, width uint) string {
	hash := db.CreateBookHash(filePath)
	return createPageFilePathFromHash(hash, index, height, width)
}

func createPageFilePathFromHash(hash string, index int, height uint, width uint) string {
	dirPath := filepath.Join(PageDirPath, hash)

	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		os.Mkdir(dirPath, 0777)
	}
	return filepath.Join(dirPath, fmt.Sprintf("%d_%d_%d.jpg", index, height, width))
}

func IsExistPageFile(hash string, index int, maxHeight int, maxWidth int) (bool, string) {
	bookRecord, err := db.SelectBookFromHash(hash)
	if err != nil {
		return false, "" //取得失敗
	}

	height, width := convertResizeImageSize(maxHeight, maxWidth)
	pageFilePath := createPageFilePath(bookRecord.FilePath, index, height, width)
	_, err = os.Stat(pageFilePath)
	if !os.IsNotExist(err) {
		return true, pageFilePath //ファイルあり
	}

	return false, pageFilePath //ファイルなし
}

func UnzipPageFileMutex(hash string, index int, limit int, maxHeight int, maxWidth int) {
	busyParam := fmt.Sprintf("%s_%d_%d_%d_%d", hash, index, limit, maxHeight, maxWidth)
	if unzipBusyParam == busyParam {
		fmt.Print("UnzipPageFileMutex 指定データは現在処理中...")
		return
	}

	//展開処理は1つずつ順番に行うようロックする
	if unzipMutex == nil {
		unzipMutex = new(sync.Mutex)
	}
	unzipMutex.Lock()
	defer unzipMutex.Unlock()

	unzipBusyParam = busyParam
	defer func() {
		unzipBusyParam = ""
	}()
	unzipPageFile(hash, index, limit, maxHeight, maxWidth)
}

func unzipPageFile(hash string, index int, limit int, maxHeight int, maxWidth int) (int, error) {
	fmt.Printf("unzipPageFile start hash=%s, index=%d\n", hash, index)
	start := time.Now()
	bookRecord, err := db.SelectBookFromHash(hash)
	if err != nil {
		fmt.Printf("unzipPageFile ハッシュエラー\n")
		return 0, err
	}

	//ZIPファイルを開く
	r, err := zip.OpenReader(bookRecord.FilePath)
	if err != nil {
		fmt.Printf("unzipPageFile ZIPファイルオープンエラー err:%s\n", err)
		return 0, err
	}
	defer r.Close()

	count := 0
	if limit == 0 {
		limit = len(r.File) //最大数が未指定の時はすべてチェックする
	}
	height, width := convertResizeImageSize(maxHeight, maxWidth)
	for i, imageFile := range r.File {
		if i < index || i >= (index+limit) {
			continue
		}
		if imageFile.FileInfo().IsDir() {
			continue
		}

		//既にファイルがあるかどうか確認
		outputFilePath := createPageFilePathFromHash(hash, i, height, width)
		_, err := os.Stat(outputFilePath)
		if !os.IsNotExist(err) {
			continue
		}

		//ページ画像ファイル取得
		rc, err := imageFile.Open()
		if err != nil {
			fmt.Printf("unzipPageFile ZIP内ファイルオープンエラー err:%s\n", err)
			continue
		}
		defer rc.Close()

		//指定したサイズに縮小
		err = saveResizeImage(rc, width, height, PageJpegQuality, outputFilePath)
		if err != nil {
			fmt.Printf("unzipPageFile リサイズ失敗 err:%s\n", err)
			continue
		}
		count++
		fmt.Printf("unzipPageFile saveResizeImage=%s\n", outputFilePath)
	}

	end := time.Now()
	fmt.Printf("unzipPageFile finish count=%d time=%f\n", count, (end.Sub(start)).Seconds())
	return count, nil
}

package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mryp/squidgirl-go/db"
)

const (
	PageDirPath     = "_temp/cache"
	PageJpegQuality = 70
)

var (
	unzipMutex     *sync.Mutex = nil
	unzipBusyParam             = ""
)

func GetArchivePageCount(filePath string) (int, error) {
	fmt.Printf("GetArchivePageCount filePath=%s\n", filePath)
	return getZipFileCount(filePath)
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

func CreatePageFile(hash string, index int, maxHeight uint, maxWidth uint) (string, error) {
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

func createPageFileFromZip(filePath string, index int, maxHeight uint, maxWidth uint) (string, error) {
	//キャッシュされているかどうか確認
	outputPath := createPageFilePath(filePath, index, maxHeight, maxWidth)
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
	resize := NewResize(uint(maxHeight), uint(maxWidth), PageJpegQuality)
	resize.ResizeFile(rc, outputPath)
	return outputPath, nil
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

func IsExistPageFile(hash string, index int, maxHeight uint, maxWidth uint) (bool, string) {
	bookRecord, err := db.SelectBookFromHash(hash)
	if err != nil {
		return false, "" //取得失敗
	}

	resize := NewResize(maxHeight, maxWidth, PageJpegQuality)
	pageFilePath := createPageFilePath(bookRecord.FilePath, index, resize.height, resize.width)
	_, err = os.Stat(pageFilePath)
	if !os.IsNotExist(err) {
		return true, pageFilePath //ファイルあり
	}

	return false, pageFilePath //ファイルなし
}

func UnzipPageFileMutex(hash string, index int, limit int, maxHeight uint, maxWidth uint) {
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
	UnzipPageFile(hash, index, limit, maxHeight, maxWidth)
}

func UnzipPageFile(hash string, index int, limit int, maxHeight uint, maxWidth uint) (int, error) {
	fmt.Printf("UnzipPageFile start hash=%s, index=%d\n", hash, index)
	//time.Sleep(3 * time.Second)
	start := time.Now()
	bookRecord, err := db.SelectBookFromHash(hash)
	if err != nil {
		fmt.Printf("UnzipPageFile ハッシュエラー\n")
		return 0, err
	}

	//ZIPファイルを開く
	r, err := zip.OpenReader(bookRecord.FilePath)
	if err != nil {
		fmt.Printf("UnzipPageFile ZIPファイルオープンエラー err:%s\n", err)
		return 0, err
	}
	defer r.Close()

	count := 0
	for i, imageFile := range r.File {
		if i < index || i >= (index+limit) {
			continue
		}
		if imageFile.FileInfo().IsDir() {
			continue
		}
		resize := NewResize(maxHeight, maxWidth, PageJpegQuality)

		//既にファイルがあるかどうか確認
		outputFilePath := createPageFilePathFromHash(hash, i, resize.height, resize.width)
		_, err := os.Stat(outputFilePath)
		if !os.IsNotExist(err) {
			continue
		}

		//ページ画像ファイル取得
		rc, err := imageFile.Open()
		if err != nil {
			fmt.Printf("UnzipPageFile ZIP内ファイルオープンエラー err:%s\n", err)
			continue
		}
		defer rc.Close()

		//指定したサイズに縮小
		err = resize.ResizeFile(rc, outputFilePath)
		if err != nil {
			fmt.Printf("UnzipPageFile リサイズ失敗 err:%s\n", err)
			continue
		}
		count++
		fmt.Printf("unzipPageFile saveResizeImage=%s\n", outputFilePath)
	}

	end := time.Now()
	fmt.Printf("UnzipPageFile finish count=%d time=%f\n", count, (end.Sub(start)).Seconds())
	return count, nil
}

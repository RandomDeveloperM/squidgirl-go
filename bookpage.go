package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

//BookPage は書庫ページ情報を保持する構造体
type BookPage struct {
	Hash     string
	FilePath string
}

var (
	unzipMutex     *sync.Mutex
	unzipBusyParam string
)

//NewBookPage は書庫ページ情報を生成する
//引数のhash, filePathはいずれかが空文字でも可能となりその場合は内部で取得して設定する
func NewBookPage(hash string, filePath string) *BookPage {
	bookPage := new(BookPage)
	if hash == "" {
		hash = db.CreateBookHash(filePath)
	}
	if filePath == "" {
		bookRecord, err := db.SelectBookFromHash(hash)
		if err != nil {
			return nil
		}
		filePath = bookRecord.FilePath
	}
	bookPage.Hash = hash
	bookPage.FilePath = filePath
	return bookPage
}

//GetPageCount は書庫ページ数を取得する
func (bookPage *BookPage) GetPageCount() (int, error) {
	r, err := zip.OpenReader(bookPage.FilePath)
	if err != nil {
		fmt.Printf("getZipFileCount err=%s\n", err)
		return 0, err
	}
	defer r.Close()

	count := len(r.File)
	return count, nil
}

//CreatePageFile はページ画像ファイルのパスを取得する
func (bookPage *BookPage) CreatePageFile(index int, maxHeight uint, maxWidth uint) (string, error) {
	//キャッシュされているかどうか確認
	exist, outputPath := bookPage.IsExistPageFile(index, maxHeight, maxWidth)
	if exist {
		return outputPath, nil
	}

	//ZIPファイルを開く
	r, err := zip.OpenReader(bookPage.FilePath)
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
	resize := NewResize(uint(maxHeight), uint(maxWidth), config.GetConfig().File.PageJpegQuality)
	resize.ResizeFile(rc, outputPath)
	return outputPath, nil
}

//IsExistPageFileFromResize は指定したページ位置の画像が存在するかどうかを返す（サイズ指定がResize版）
func (bookPage *BookPage) IsExistPageFileFromResize(index int, resize *Resize) (bool, string) {
	pageFilePath := bookPage.createPageFilePath(index, resize.height, resize.width)
	_, err := os.Stat(pageFilePath)
	if !os.IsNotExist(err) {
		return true, pageFilePath //ファイルあり
	}

	return false, pageFilePath //ファイルなし
}

//IsExistPageFile は指定したページ位置の画像が存在するかどうかを返す
func (bookPage *BookPage) IsExistPageFile(index int, maxHeight uint, maxWidth uint) (bool, string) {
	resize := NewResize(maxHeight, maxWidth, config.GetConfig().File.PageJpegQuality)
	return bookPage.IsExistPageFileFromResize(index, resize)
}

//createPageFilePath は書庫ページのファイルパスを生成して返す
func (bookPage *BookPage) createPageFilePath(index int, height uint, width uint) string {
	dirPath := filepath.Join(config.GetConfig().File.PageDirPath, bookPage.Hash)

	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		os.Mkdir(dirPath, 0777)
	}
	return filepath.Join(dirPath, fmt.Sprintf("%d_%d_%d.jpg", index, height, width))
}

//UnzipPageFileMutex 書庫ファイルから画像ファイルを作成する（同時処理制御版）
func (bookPage *BookPage) UnzipPageFileMutex(index int, limit int, maxHeight uint, maxWidth uint) {
	busyParam := fmt.Sprintf("%s_%d_%d_%d_%d", bookPage.Hash, index, limit, maxHeight, maxWidth)
	if unzipBusyParam == busyParam {
		//処理中の時は待つのではなくあきらめる
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
	bookPage.UnzipPageFile(index, limit, maxHeight, maxWidth)
}

//UnzipPageFile 書庫ファイルから画像ファイルを作成する
func (bookPage *BookPage) UnzipPageFile(index int, limit int, maxHeight uint, maxWidth uint) (int, error) {
	fmt.Printf("UnzipPageFile start index=%d, limit=%d\n", index, limit)
	//time.Sleep(3 * time.Second)
	start := time.Now()

	//ZIPファイルを開く
	r, err := zip.OpenReader(bookPage.FilePath)
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
		resize := NewResize(maxHeight, maxWidth, config.GetConfig().File.PageJpegQuality)

		//既にファイルがあるかどうか確認
		exist, outputFilePath := bookPage.IsExistPageFileFromResize(i, resize)
		if exist {
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

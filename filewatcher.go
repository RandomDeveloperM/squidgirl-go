package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"time"

	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

var (
	fileWatcher          *FileWatcher = nil
	fileWatcherTargetExt []string     = []string{".zip"}
)

type FileWatcher struct {
	mutex         *sync.Mutex
	maxCacheCount int
}

type FileWatcherFunc interface {
	RegistFile()
	ClearFile()
	ClearCache()
}

func NewFileWatcher() *FileWatcher {
	if fileWatcher != nil {
		return fileWatcher
	}

	watcher := new(FileWatcher)
	watcher.mutex = new(sync.Mutex)
	watcher.maxCacheCount = config.GetConfig().File.CacheMaxCount
	return watcher
}

//AddFile 新規に存在するフォルダ・ファイルを追加する
func (watcher *FileWatcher) RegistFile() {
	//ロックをかける
	watcher.mutex.Lock()
	defer watcher.mutex.Unlock()

	//ファイル探索開始
	baseDir := config.GetConfig().File.WatchDir
	err := filepath.Walk(baseDir, registFileWalk)
	if err != nil {
		return
	}
}

//ClearFile 登録されているファイル・フォルダが存在しなかった時は削除する
func (watcher *FileWatcher) ClearFile() {
	watcher.mutex.Lock()
	defer watcher.mutex.Unlock()

	clearFolderAll()
	clearBookAll()
}

//ClearCache 使用頻度が低いキャッシュファイルを削除する
func (watcher *FileWatcher) ClearCache() {
	watcher.mutex.Lock()
	defer watcher.mutex.Unlock()

	clearOldCacheAll(watcher.maxCacheCount)
}

func registFileWalk(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		registDirInfo(path, info)
	} else {
		registFileInfo(path, info)
	}
	return nil
}

func registDirInfo(path string, info os.FileInfo) {
	folder, _ := db.SelectFolder(path)
	if folder.Hash == "" {
		parentDir := filepath.Dir(path)
		parentFolder, err := db.SelectFolder(parentDir)
		if err != nil {
			fmt.Printf("registDirInfo 親フォルダ未登録 err%s\n", err)
			return
		}
		db.InsertFolder(path, parentFolder.Hash, info.ModTime())
	} else {
		fmt.Printf("registDirInfo dir exists hash=%s\n", folder.Hash)
	}
}

func registFileInfo(path string, info os.FileInfo) {
	ext := filepath.Ext(path)
	for _, v := range fileWatcherTargetExt {
		if v == ext {
			registFileZipInfo(path, info)
			break
		}
	}
}

func registFileZipInfo(path string, info os.FileInfo) {
	dir := filepath.Dir(path)
	folder, err := db.SelectFolder(dir)
	if err != nil {
		fmt.Printf("registFileZipInfo フォルダ未登録 err%s\n", err)
		return
	}
	dirHash := folder.Hash

	book, _ := db.SelectBook(path)
	if book.Hash == "" {
		//新規登録
		page, _ := GetArchivePageCount(path)
		CreateThumbnailFile(path)
		db.InsertBook(dirHash, path, int(info.Size()), page, info.ModTime())
	} else if !isEquleDateTime(book.ModTime, info.ModTime()) {
		//更新あり
		page, _ := GetArchivePageCount(path)
		CreateThumbnailFile(path)
		db.UpdateBook(dirHash, path, int(info.Size()), page, info.ModTime())
	} else {
		fmt.Printf("registFileZipInfo file exists hash=%s\n", book.Hash)
	}
}

//指定したフィル時刻が同一かどうか（分単位まででチェックする）
func isEquleDateTime(t1 time.Time, t2 time.Time) bool {
	t1Text := t1.UTC().Format("2006-01-02 15:04")
	t2Text := t2.UTC().Format("2006-01-02 15:04")
	if t1Text == t2Text {
		return true
	}
	return false
}

func clearFolderAll() {
	folderList, err := db.SelectFolderAll()
	if err != nil {
		return
	}

	for _, folder := range folderList {
		_, err := os.Stat(folder.FilePath)
		if !os.IsNotExist(err) {
			continue //フォルダあり
		}

		db.DeleteFolder(folder.ID)
	}
}

func clearBookAll() {
	bookList, err := db.SelectBookAll()
	if err != nil {
		return
	}

	for _, book := range bookList {
		_, err := os.Stat(book.FilePath)
		if !os.IsNotExist(err) {
			continue //フォルダあり
		}

		db.DeleteBook(book.ID)
	}
}

func clearOldCacheAll(maxCount int) {
	fileInfoList, err := ioutil.ReadDir(PageDirPath)
	if err != nil {
		return
	}
	sort.Slice(fileInfoList, func(i, j int) bool {
		return fileInfoList[i].ModTime().Unix() > fileInfoList[j].ModTime().Unix()
	})

	for i, dir := range fileInfoList {
		fmt.Printf("clearOldCacheAll [%v] %s\n", dir.ModTime(), dir.Name())
		if !dir.IsDir() {
			continue
		}
		if i < maxCount {
			continue
		}

		dirPath := filepath.Join(PageDirPath, dir.Name())
		err := os.RemoveAll(dirPath)
		if err != nil {
			fmt.Printf("clearOldCacheAll RemoveAll err=%s", err)
			continue
		}
		fmt.Printf("clearOldCacheAll Remove OK")
	}
}

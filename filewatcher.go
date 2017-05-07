package main

import (
	"fmt"
	"os"
	"path/filepath"

	"time"

	"github.com/mryp/squidgirl-go/config"
	"github.com/mryp/squidgirl-go/db"
)

var fileWatcherBusy = false
var fileWatcherTargetExt = []string{".zip"}

func RegisterFileWatchar() {
	if fileWatcherBusy {
		fmt.Printf("RegisterFileWatchar is bussy...\n")
		return
	}
	fileWatcherBusy = true
	fmt.Printf("RegisterFileWatchar start\n")

	baseDir := config.GetConfig().File.WatchDir
	err := filepath.Walk(baseDir, registFileWalk)
	if err != nil {
		fmt.Printf("RegisterFileWatchar err=%v\n", err)
		return
	}

	fmt.Printf("RegisterFileWatchar finish\n")
	fileWatcherBusy = false
}

func registFileWalk(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		fmt.Printf("registFileWalk dir=%s time=%s\n", path, info.ModTime())
		registDirInfo(path, info)
	} else {
		fmt.Printf("registFileWalk file=%s time=%s\n", path, info.ModTime())
		registFileInfo(path, info)
	}
	return nil
}

func registDirInfo(path string, info os.FileInfo) {
	folder, _ := db.SelectFolder(path)
	if folder.Hash == "" {
		db.InsertFolder(path, info.ModTime())
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
		page := 0
		db.InsertBook(dirHash, path, int(info.Size()), page, info.ModTime())
	} else if !isEquleDateTime(book.ModTime, info.ModTime()) {
		//更新あり
		page := 0
		db.UpdateBook(dirHash, path, int(info.Size()), page, info.ModTime())
	} else {
		fmt.Printf("registFileZipInfo file exists hash=%s\n", book.Hash)
	}
}

func isEquleDateTime(t1 time.Time, t2 time.Time) bool {
	t1Text := t1.UTC().Format("2006-01-02 15:04")
	t2Text := t2.UTC().Format("2006-01-02 15:04")
	fmt.Printf("isEquleDateTime t1=%s t2=%s\n", t1Text, t2Text)
	if t1Text == t2Text {
		return true
	}
	return false
}

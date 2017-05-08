package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"time"

	"github.com/labstack/echo"
	"github.com/mryp/squidgirl-go/db"
)

//未設定時の時刻
var unknownTime = time.Unix(0, 0).UTC()

//FileListRequest ファイルリストリクエストデータ
type FileListRequest struct {
	Hash string `json:"hash" xml:"hash" form:"hash" query:"hash"`
}

//FileListResponce ファイルリストレスポンスデータ
type FileListResponce struct {
	Hash     string    `json:"hash" xml:"hash"`
	Name     string    `json:"name" xml:"name"`
	Size     int       `json:"size" xml:"size"`
	Page     int       `json:"page" xml:"page"`
	IsDir    bool      `json:"isdir" xml:"isdir"`
	ModTime  time.Time `json:"modtime" xml:"modtime"`
	ReadTime time.Time `json:"readtime" xml:"readtime"`
	ReadPos  int       `json:"readpos" xml:"readpos"`
	Reaction int       `json:"reaction" xml:"reaction"`
}

//FileListHandler ユーザーログインハンドラ
func FileListHandler(c echo.Context) error {
	req := new(FileListRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	folderHash := req.Hash
	if folderHash == "" {
		//フォルダハッシュをルートとして取得
		folder, err := db.SelectFolderRoot()
		if err != nil {
			return err
		}
		folderHash = folder.Hash
	}

	//フォルダ一覧を取得
	folderList, err := db.SelectFolderListFromParent(folderHash)
	if err != nil {
		return err
	}

	//ファイル一覧を取得
	bookList, err := db.SelectBookListFromFolder(folderHash)
	if err != nil {
		return err
	}

	//レスポンスを作成
	responce := make([]FileListResponce, 0)
	for _, v := range folderList {
		responce = append(responce, createFileListResponceFromFolder(v))
	}
	for _, v := range bookList {
		responce = append(responce, createFileListResponceFromBook(v))
	}

	return c.JSON(http.StatusOK, responce)
}

func createFileListResponceFromFolder(folder db.FolderTable) FileListResponce {
	name := filepath.Base(folder.FilePath)
	return FileListResponce{
		Hash:     folder.Hash,
		Name:     name,
		Size:     0,
		Page:     0,
		IsDir:    true,
		ModTime:  folder.ModTime.UTC(),
		ReadTime: unknownTime,
		ReadPos:  0,
		Reaction: 0,
	}
}

func createFileListResponceFromBook(book db.BookTable) FileListResponce {
	name := filepath.Base(book.FilePath)
	readTime := unknownTime
	readPos := 0
	reaction := 0

	return FileListResponce{
		Hash:     book.Hash,
		Name:     name,
		Size:     book.FileSize,
		Page:     book.Page,
		IsDir:    false,
		ModTime:  book.ModTime.UTC(),
		ReadTime: readTime,
		ReadPos:  readPos,
		Reaction: reaction,
	}
}

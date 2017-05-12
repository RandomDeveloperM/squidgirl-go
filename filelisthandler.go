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
	Hash   string `json:"hash" xml:"hash" form:"hash" query:"hash"`
	Offset int    `json:"offset" xml:"offset" form:"offset" query:"offset"`
	Limit  int    `json:"limit" xml:"limit" form:"limit" query:"limit"`
}

//FileListResponce ファイルリストレスポンスデータ

type FileListResponce struct {
	Name     string                  `json:"name" xml:"name"`
	AllCount int                     `json:"allcount" xml:"allcount"`
	Count    int                     `json:"count" xml:"count"`
	Files    []FileListFilesResponce `json:"files" xml:"files"`
}

type FileListFilesResponce struct {
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

	//ルートを取得
	rootFolder, err := db.SelectFolderRoot()
	if err != nil {
		return err
	}

	folderHash := req.Hash
	if folderHash == "" {
		//フォルダハッシュをルートとして取得
		folderHash = rootFolder.Hash
	}

	//指定したフォルダの親フォルダを取得する
	selectFolder, err := db.SelectFolderFromHash(folderHash)
	if err != nil {
		return err
	}
	var parentFolder db.FolderTable
	if selectFolder.ParentHash != "" {
		parentFolder, err = db.SelectFolderFromHash(selectFolder.ParentHash)
		if err != nil {
			return err
		}
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

	//ファイル情報レスポンスを作成
	files := make([]FileListFilesResponce, 0)
	if parentFolder.Hash != "" {
		//親フォルダは必ず追加する
		files = append(files, createFileListResponceFromUpperFolder(parentFolder))
	}
	index := 0
	for _, v := range folderList {
		if index >= req.Offset && index < req.Offset+req.Limit {
			files = append(files, createFileListResponceFromFolder(v))
		}
		index++
	}
	for _, v := range bookList {
		if index >= req.Offset && index < req.Offset+req.Limit {
			files = append(files, createFileListResponceFromBook(v))
		}
		index++
	}

	//取得フォルダ情報レスポンスを作成
	responce := new(FileListResponce)
	if rootFolder.Hash == selectFolder.Hash {
		responce.Name = "ルートフォルダ"
	} else {
		responce.Name = filepath.Base(selectFolder.FilePath)
	}
	responce.AllCount = index
	responce.Count = len(files)
	responce.Files = files
	return c.JSON(http.StatusOK, responce)
}

func createFileListResponceFromUpperFolder(folder db.FolderTable) FileListFilesResponce {
	return FileListFilesResponce{
		Hash:     folder.Hash,
		Name:     "..",
		Size:     0,
		Page:     0,
		IsDir:    true,
		ModTime:  folder.ModTime.UTC(),
		ReadTime: unknownTime,
		ReadPos:  0,
		Reaction: 0,
	}
}

func createFileListResponceFromFolder(folder db.FolderTable) FileListFilesResponce {
	name := filepath.Base(folder.FilePath)
	return FileListFilesResponce{
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

func createFileListResponceFromBook(book db.BookTable) FileListFilesResponce {
	name := filepath.Base(book.FilePath)
	readTime := unknownTime
	readPos := 0
	reaction := 0

	return FileListFilesResponce{
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

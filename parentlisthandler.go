package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo"
	"github.com/mryp/squidgirl-go/db"
)

//ParentListRequest はファイル一覧情報取得のリクエストデータを保持する
type ParentListRequest struct {
	Hash string `json:"hash" xml:"hash" form:"hash" query:"hash"`
}

//ParentListResponce はファイル一覧情報取得のレスポンスデータを保持する
type ParentListResponce struct {
	Count   int                        `json:"count" xml:"count"`
	Folders []ParentListFolderResponce `json:"folders" xml:"folders"`
}

//ParentListFolderResponce はファイル一覧取得レスポンスのファイル情報をを保持する
type ParentListFolderResponce struct {
	Hash string `json:"hash" xml:"hash"`
	Name string `json:"name" xml:"name"`
}

//ParentListHandler は親フォルダ一覧取得し返す
func ParentListHandler(c echo.Context) error {
	req := new(ParentListRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	fmt.Printf("request=%v\n", *req)

	//フォルダ情報レスポンスを作成
	folders := make([]ParentListFolderResponce, 0)

	//ルートを取得
	rootFolder, err := db.SelectFolderRoot()
	if err != nil {
		return err
	}

	selectHash := req.Hash
	if selectHash == "" {
		//フォルダハッシュをルートとして取得
		selectHash = rootFolder.Hash
	}

	//現在のフォルダーを取得
	selectFolder, err := db.SelectFolderFromHash(selectHash)
	if err != nil {
		return err
	}
	folders = append(folders, createFolderItemFromFolder(selectFolder, rootFolder))

	//親フォルダをさかのぼって追加
	parentHash := selectFolder.ParentHash
	for {
		if parentHash == "" {
			break
		}

		parentFolder, err := db.SelectFolderFromHash(parentHash)
		if err != nil {
			break
		}
		folders = append(folders, createFolderItemFromFolder(parentFolder, rootFolder))
		parentHash = parentFolder.ParentHash
	}

	//レスポンスを作成
	responce := new(ParentListResponce)
	responce.Count = len(folders)
	responce.Folders = folders
	return c.JSON(http.StatusOK, responce)
}

func createFolderItemFromFolder(folder db.FolderTable, rootFolder db.FolderTable) ParentListFolderResponce {
	name := filepath.Base(folder.FilePath)
	if rootFolder.Hash == folder.Hash {
		name = "ルートフォルダ"
	}
	return ParentListFolderResponce{
		Hash: folder.Hash,
		Name: name,
	}
}

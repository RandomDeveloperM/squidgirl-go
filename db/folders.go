package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocraft/dbr"
	"github.com/mryp/squidgirl-go/config"
)

//テーブル名
const folderTableName = "folders"

//FolderTable アーカイブ情報テーブル
type FolderTable struct {
	ID         int64     `db:"id"`
	Hash       string    `db:"hash"`
	ParentHash string    `db:"parent_hash"`
	FilePath   string    `db:"file_path"`
	ModTime    time.Time `db:"mod_time"`
}

func InsertFolder(filePath string, parentHash string, modTime time.Time) error {
	fmt.Printf("InsertFolder filePath=%s, modTime=%s\n", filePath, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := CreateFolderHash(filePath)
	record := FolderTable{Hash: hash, ParentHash: parentHash, FilePath: filePath, ModTime: modTime}
	err := insertFolder(nil, record)
	if err != nil {
		fmt.Printf("InsertFolder err=%s\n", err)
		return err
	}
	return nil
}

func UpdateFolder(filePath string, parentHash string, modTime time.Time) error {
	fmt.Printf("UpdateFolder filePath=%s, modTime=%s\n", filePath, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := CreateFolderHash(filePath)
	record := FolderTable{Hash: hash, ParentHash: parentHash, FilePath: filePath, ModTime: modTime}
	err := updateFolder(nil, record)
	if err != nil {
		fmt.Printf("UpdateFolder err=%s\n", err)
		return err
	}
	return nil
}

func DeleteFolder(id int64) error {
	fmt.Printf("DeleteFolder id=%d\n", id)

	err := deleteFolder(nil, id)
	if err != nil {
		fmt.Printf("DeleteFolder err=%s\n", err)
		return err
	}
	return nil
}

func SelectFolder(filePath string) (FolderTable, error) {
	fmt.Printf("SelectFolder filePath=%s\n", filePath)
	var result FolderTable
	hash := CreateFolderHash(filePath)
	recordList, err := selectFolderList(nil, hash)
	if err != nil {
		fmt.Printf("SelectFolder err=%s\n", err)
		return result, err
	}

	if len(recordList) == 0 {
		fmt.Printf("SelectFolder len==0\n")
		return result, nil
	}
	return recordList[0], nil
}

func SelectFolderFromHash(hash string) (FolderTable, error) {
	fmt.Printf("SelectFolderFromHash hash=%s\n", hash)
	var result FolderTable
	recordList, err := selectFolderList(nil, hash)
	if err != nil {
		fmt.Printf("SelectFolderFromHash err=%s\n", err)
		return result, err
	}

	if len(recordList) == 0 {
		fmt.Printf("SelectFolderFromHash len==0\n")
		return result, nil
	}
	return recordList[0], nil
}

func SelectFolderListFromParent(parentHash string) ([]FolderTable, error) {
	fmt.Printf("SelectFolderListFromParent parentHash=%s\n", parentHash)
	recordList, err := selectFolderListFromParent(nil, parentHash)
	if err != nil {
		fmt.Printf("SelectFolderListFromParent err=%s\n", err)
		return nil, err
	}

	return recordList, nil
}

func SelectFolderRoot() (FolderTable, error) {
	fmt.Printf("SelectFolderRoot\n")
	return SelectFolder(config.GetConfig().File.WatchDir)
}

func SelectFolderAll() ([]FolderTable, error) {
	fmt.Printf("SelectFolderAll\n")
	recordList, err := selectFolderListAll(nil)
	if err != nil {
		fmt.Printf("SelectFolderAll err=%s\n", err)
		return nil, err
	}

	if len(recordList) == 0 {
		fmt.Printf("SelectFolderAll len==0\n")
		return recordList, nil
	}
	return recordList, nil
}

func insertFolder(session *dbr.Session, record FolderTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil
	}
	defer session.Close()

	_, err = session.InsertInto(folderTableName).
		Columns("hash", "parent_hash", "file_path", "mod_time").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateFolder(session *dbr.Session, record FolderTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil
	}
	defer session.Close()

	_, err = session.Update(folderTableName).
		Set("mod_time", record.ModTime).
		Where("hash = ?", record.Hash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func deleteFolder(session *dbr.Session, id int64) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil
	}
	defer session.Close()
	_, err = session.DeleteFrom(folderTableName).
		Where("id = ?", id).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectFolderList(session *dbr.Session, hash string) ([]FolderTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []FolderTable
	_, err = session.Select("*").From(folderTableName).Where("hash = ?", hash).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func selectFolderListFromParent(session *dbr.Session, parentHash string) ([]FolderTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []FolderTable
	_, err = session.Select("*").From(folderTableName).Where("parent_hash = ?", parentHash).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func selectFolderListAll(session *dbr.Session) ([]FolderTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []FolderTable
	_, err = session.Select("*").From(folderTableName).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func CreateFolderHash(filePath string) string {
	hashBytes := sha256.Sum256([]byte(filePath))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}

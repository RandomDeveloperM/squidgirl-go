package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocraft/dbr"
)

//テーブル名
const folderTableName = "folders"

//FolderTable アーカイブ情報テーブル
type FolderTable struct {
	ID       int64     `db:"id"`
	Hash     string    `db:"hash"`
	FilePath string    `db:"file_path"`
	ModTime  time.Time `db:"mod_time"`
}

func InsertFolder(filePath string, modTime time.Time) error {
	fmt.Printf("InsertFolder filePath=%s, modTime=%s\n", filePath, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createFolderHash(filePath)
	record := FolderTable{Hash: hash, FilePath: filePath, ModTime: modTime}
	err := insertFolder(nil, record)
	if err != nil {
		fmt.Printf("InsertFolder err=%s\n", err)
		return err
	}
	return nil
}

func UpdateFolder(filePath string, modTime time.Time) error {
	fmt.Printf("UpdateFolder filePath=%s, modTime=%s\n", filePath, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createFolderHash(filePath)
	record := FolderTable{Hash: hash, FilePath: filePath, ModTime: modTime}
	err := updateFolder(nil, record)
	if err != nil {
		fmt.Printf("UpdateFolder err=%s\n", err)
		return err
	}
	return nil
}

func SelectFolder(filePath string) (FolderTable, error) {
	fmt.Printf("SelectFolder filePath=%s\n", filePath)
	var result FolderTable
	hash := createFolderHash(filePath)
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

func insertFolder(session *dbr.Session, record FolderTable) error {
	if session == nil {
		newSession, err := ConnectDB()
		if err != nil {
			return err
		}
		session = newSession
		defer session.Close()
	}

	_, err := session.InsertInto(folderTableName).
		Columns("hash", "file_path", "mod_time").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateFolder(session *dbr.Session, record FolderTable) error {
	if session == nil {
		newSession, err := ConnectDB()
		if err != nil {
			return err
		}
		session = newSession
		defer session.Close()
	}

	_, err := session.Update(folderTableName).
		Set("mod_time", record.ModTime).
		Where("hash = ?", record.Hash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectFolderList(session *dbr.Session, hash string) ([]FolderTable, error) {
	if session == nil {
		newSession, err := ConnectDB()
		if err != nil {
			return nil, err
		}
		session = newSession
		defer newSession.Close()
	}

	var resultList []FolderTable
	_, err := session.Select("*").From(folderTableName).Where("hash = ?", hash).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func createFolderHash(filePath string) string {
	hashBytes := sha256.Sum256([]byte(filePath))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}

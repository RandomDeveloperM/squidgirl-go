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
	ID        int64     `db:"id"`
	Hash      string    `db:"hash"`
	FilePath  string    `db:"file_path"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func InsertFolder(filePath string, createTime time.Time, updateTime time.Time) error {
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createFolderHash(filePath)
	record := FolderTable{Hash: hash, FilePath: filePath, CreatedAt: createTime, UpdatedAt: updateTime}
	err := insertFolder(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func UpdateFolder(filePath string, createTime time.Time, updateTime time.Time) error {
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createFolderHash(filePath)
	record := FolderTable{Hash: hash, FilePath: filePath, CreatedAt: createTime, UpdatedAt: updateTime}
	err := updateFolder(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func SelectFolder(filePath string) (FolderTable, error) {
	var result FolderTable
	hash := createFolderHash(filePath)
	recordList, err := selectFolderList(nil, hash)
	if err != nil {
		return result, err
	}

	if len(recordList) == 0 {
		return result, nil
	}
	return recordList[0], nil
}

func insertFolder(session *dbr.Session, record FolderTable) error {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return err
		}
		defer session.Close()
	}

	_, err := session.InsertInto(folderTableName).
		Columns("hash", "file_path", "file_count", "created_at", "updated_at").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateFolder(session *dbr.Session, record FolderTable) error {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return err
		}
		defer session.Close()
	}

	_, err := session.Update(folderTableName).
		Set("created_at", record.CreatedAt).
		Set("updated_at", record.UpdatedAt).
		Where("hash = ?", record.Hash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectFolderList(session *dbr.Session, hash string) ([]FolderTable, error) {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return nil, err
		}
		defer session.Close()
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

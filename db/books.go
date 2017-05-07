package db

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gocraft/dbr"
)

//テーブル名
const bookTableName = "books"

//BookTable アーカイブ情報テーブル
type BookTable struct {
	ID         int64     `db:"id"`
	Hash       string    `db:"hash"`
	FolderHash string    `db:"folder_hash"`
	FilePath   string    `db:"file_path"`
	FileSize   int       `db:"file_size"`
	Page       int       `db:"page"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func InsertBook(folderHash string, filePath string, fileSize int, page int, createTime time.Time, updateTime time.Time) error {
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createBookHash(filePath)
	record := BookTable{FolderHash: folderHash, Hash: hash, FilePath: filePath, FileSize: fileSize, Page: page, CreatedAt: createTime, UpdatedAt: updateTime}
	err := insertBook(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func UpdateBook(folderHash string, filePath string, fileSize int, page int, createTime time.Time, updateTime time.Time) error {
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createBookHash(filePath)
	record := BookTable{FolderHash: folderHash, Hash: hash, FilePath: filePath, FileSize: fileSize, Page: page, CreatedAt: createTime, UpdatedAt: updateTime}
	err := updateBook(nil, record)
	if err != nil {
		return err
	}
	return nil
}

func SelectBook(filePath string) (BookTable, error) {
	var result BookTable
	hash := createBookHash(filePath)
	recordList, err := selectBookList(nil, hash)
	if err != nil {
		return result, err
	}

	if len(recordList) == 0 {
		return result, nil
	}
	return recordList[0], nil
}

func insertBook(session *dbr.Session, record BookTable) error {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return err
		}
		defer session.Close()
	}

	_, err := session.InsertInto(bookTableName).
		Columns("hash", "folder_hash", "file_path", "file_size", "page", "created_at", "updated_at").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateBook(session *dbr.Session, record BookTable) error {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return err
		}
		defer session.Close()
	}

	_, err := session.Update(bookTableName).
		Set("file_size", record.FileSize).
		Set("page", record.Page).
		Set("created_at", record.CreatedAt).
		Set("updated_at", record.UpdatedAt).
		Where("hash = ?", record.Hash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectBookList(session *dbr.Session, hash string) ([]BookTable, error) {
	if session == nil {
		session, err := ConnectDB()
		if err != nil {
			return nil, err
		}
		defer session.Close()
	}

	var resultList []BookTable
	_, err := session.Select("*").From(bookTableName).Where("hash = ?", hash).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func createBookHash(filePath string) string {
	hashBytes := sha256.Sum256([]byte(filePath))
	hash := hex.EncodeToString(hashBytes[:])
	return hash
}

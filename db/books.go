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
	ModTime    time.Time `db:"mod_time"`
}

func InsertBook(folderHash string, filePath string, fileSize int, page int, modTime time.Time) error {
	fmt.Printf("InsertBook folderHash=%s, filePath=%s, fileSize=%d, page=%d, modTime=%s\n", folderHash, filePath, fileSize, page, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createBookHash(filePath)
	record := BookTable{FolderHash: folderHash, Hash: hash, FilePath: filePath, FileSize: fileSize, Page: page, ModTime: modTime}
	err := insertBook(nil, record)
	if err != nil {
		fmt.Printf("InsertBook err=%s\n", err)
		return err
	}
	return nil
}

func UpdateBook(folderHash string, filePath string, fileSize int, page int, modTime time.Time) error {
	fmt.Printf("UpdateBook folderHash=%s, filePath=%s, fileSize=%d, page=%d, modTime=%s\n", folderHash, filePath, fileSize, page, modTime)
	if filePath == "" {
		return fmt.Errorf("パラメーターエラー")
	}

	hash := createBookHash(filePath)
	record := BookTable{FolderHash: folderHash, Hash: hash, FilePath: filePath, FileSize: fileSize, Page: page, ModTime: modTime}
	err := updateBook(nil, record)
	if err != nil {
		fmt.Printf("UpdateBook err=%s\n", err)
		return err
	}
	return nil
}

func SelectBook(filePath string) (BookTable, error) {
	fmt.Printf("SelectBook filePath=%s\n", filePath)
	var result BookTable
	hash := createBookHash(filePath)
	recordList, err := selectBookList(nil, hash)
	if err != nil {
		fmt.Printf("SelectBook err=%s\n", err)
		return result, err
	}

	if len(recordList) == 0 {
		fmt.Printf("SelectBook len==0\n")
		return result, nil
	}
	return recordList[0], nil
}

func SelectBookListFromFolder(folderHash string) ([]BookTable, error) {
	fmt.Printf("SelectBookListFromFolder folderHash=%s\n", folderHash)
	recordList, err := selectBookListFromFolder(nil, folderHash)
	if err != nil {
		fmt.Printf("SelectBookListFromFolder err=%s\n", err)
		return nil, err
	}

	return recordList, nil
}

func insertBook(session *dbr.Session, record BookTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.InsertInto(bookTableName).
		Columns("hash", "folder_hash", "file_path", "file_size", "page", "mod_time").
		Record(record).
		Exec()
	if err != nil {
		return err
	}

	return nil
}

func updateBook(session *dbr.Session, record BookTable) error {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return err
	}
	defer session.Close()

	_, err = session.Update(bookTableName).
		Set("file_size", record.FileSize).
		Set("page", record.Page).
		Set("mod_time", record.ModTime).
		Where("hash = ?", record.Hash).
		Exec()
	if err != nil {
		return err
	}
	return nil
}

func selectBookList(session *dbr.Session, hash string) ([]BookTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []BookTable
	_, err = session.Select("*").From(bookTableName).Where("hash = ?", hash).Load(&resultList)
	if err != nil {
		return nil, err
	}

	return resultList, nil
}

func selectBookListFromFolder(session *dbr.Session, folderHash string) ([]BookTable, error) {
	session, err := ConnectDBRecheck(session)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	var resultList []BookTable
	_, err = session.Select("*").From(bookTableName).Where("folder_hash = ?", folderHash).Load(&resultList)
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

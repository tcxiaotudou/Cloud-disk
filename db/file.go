package db

import (
	"database/sql"
	mydb "flie_store_server/db/mysql"
	"fmt"
)

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// OnFileUploadFinished: 文件上传完成，保存meta到数据库
func OnFileUploadFinished(fileHash string, fileName string, fileSize int64, fileAddr string) bool {
	stmt, err := mydb.DBConn().Prepare(`insert ignore into tbl_file (file_sha1, file_name, file_size, file_addr, status) values (?, ?, ?, ?, 1);`)
	if err != nil {
		fmt.Println("prepare statement failed, err:", err)
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(fileHash, fileName, fileSize, fileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("file with hash: %s has been upload before\n", fileHash)
		}
		return true
	}
	return false
}

// GetFileMeta: 从数据库获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(`select file_sha1, file_name, file_size, file_addr from tbl_file where file_sha1 = ? and status = 1;`)
	if err != nil {
		fmt.Println("prepare statement failed, err:", err)
		return nil, err
	}
	defer stmt.Close()

	tbfile := TableFile{}
	err = stmt.QueryRow(filehash).Scan(&tbfile.FileHash, &tbfile.FileName, &tbfile.FileSize, &tbfile.FileAddr)
	if err != nil {
		fmt.Println("query row failed, err:", err)
		return nil, err
	}
	return &tbfile, nil
}

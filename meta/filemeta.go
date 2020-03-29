package meta

import mydb "flie_store_server/db"

// FileMeta: 文件元信息结构体
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta: 新增/更新文件元信息
func UpdateFileMeta(fileMeta FileMeta) {
	fileMetas[fileMeta.FileSha1] = fileMeta
}

// UpdateFileMetaDB: 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fileMeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize, fileMeta.Location)
}

// GetFileMeta: 通过sha1值获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// GetFileMetaDB: 从数据库中获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tbfile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return nil, nil
	}
	fmeta := FileMeta{
		FileSha1: tbfile.FileHash,
		FileName: tbfile.FileName.String,
		FileSize: tbfile.FileSize.Int64,
		Location: tbfile.FileAddr.String,
	}
	return &fmeta, nil
}

// GetLastFileMetas: 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	fMetaArray := make([]FileMeta, count)
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	//sort.Sort(ByUploadTime(fMetaArray))
	//return fmetaArray[0:count]
	return nil
}

// RemoveFileMeta: 删除文件元信息
func RemoveFileMeta(filesha1 string) {
	delete(fileMetas, filesha1)
}

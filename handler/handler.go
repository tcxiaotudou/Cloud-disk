package handler

import (
	"encoding/json"
	"flie_store_server/db"
	"flie_store_server/meta"
	"flie_store_server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method == "GET" {
		// 返回上传html页面
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			fmt.Println("readFile failed, err:", err)
			io.WriteString(w, "internal server")
			return
		}
		//io.WriteString(w, string(data))
		w.Write(data)
	} else if r.Method == "POST" {
		// 接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Println("get formFile failed, err:", err)
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "D:\\tmp\\" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Println("create temp file failed, err:", err)
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Println("copy file data failed, err:", err)
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		_ = meta.UpdateFileMetaDB(fileMeta)

		// TODO: 更新用户文件表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)
	}
}

// UploadSuccessHandler: 上传已完成
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// GetFileMetaHandler: 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		fmt.Println("parse fmeta to json failed, err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// FileMetaQueryHandler: 查询批量文件元信息接口
func FileMetaQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles, err := db.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// DownloadHandler: 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fmeta := meta.GetFileMeta(fsha1)
	f, err := os.Open(fmeta.Location)
	if err != nil {
		fmt.Println("open fmeta location failed, err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("read file failed, err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=\""+fmeta.FileName+"\"")
	w.Write(data)
}

// FileUpdateMetaHandler: 更新元信息接口(重命名)
func FileUpdateMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	optType := r.Form.Get("op")
	if optType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)
	data, err := json.Marshal(curFileMeta)
	if err != nil {
		fmt.Println("parse curFileMeta failed, err:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// FileDeleteHandler: 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filesha1 := r.Form.Get("filehash")
	fMeta := meta.GetFileMeta(filesha1)
	fmt.Println("要删除的文件为：", fMeta.Location)
	os.Remove(fMeta.Location)
	meta.RemoveFileMeta(filesha1)

	w.WriteHeader(http.StatusOK)
}

func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 1.解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	// 2.从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 3.查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 4.上传过则将文件信息写入用户文件表，返回成功
	suc := db.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}

package main

import (
	"flie_store_server/handler"
	"fmt"
	"net/http"
)

func main() {

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/success", handler.UploadSuccessHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/query", handler.FileMetaQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileUpdateMetaHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)
	http.HandleFunc("/user/signup", handler.SignUpHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HttpInterceptor(handler.UserInfoHandler))
	http.HandleFunc("/file/fastupload", handler.HttpInterceptor(handler.TryFastUploadHandler))
	http.HandleFunc("/file/mpupload/init", handler.HttpInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HttpInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete", handler.HttpInterceptor(handler.CompleteUploadHandler))
	//验证 cat `ls | sort -n` > /tmp/a
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		fmt.Println("start server failed, err:", err)
		return
	}
}

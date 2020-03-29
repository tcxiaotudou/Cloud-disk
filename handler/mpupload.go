package handler

import (
	"flie_store_server/cache/redis"
	"flie_store_server/util"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.RespMsg{
			Code: -1,
			Msg:  "params invalid",
			Data: nil,
		}.JSONBytes())
		return
	}
	// 2.获得redis的一个连接
	conn := redis.RedisPool().Get()
	defer conn.Close()
	// 3.生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	// 4.将初始化信息写入到redis缓存
	conn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	conn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	conn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)
	// 5.将初始化的信息返回给客户端
	w.Write(util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: upInfo,
	}.JSONBytes())
}

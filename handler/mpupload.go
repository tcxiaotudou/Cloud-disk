package handler

import (
	"flie_store_server/cache/redis"
	"flie_store_server/db"
	"flie_store_server/util"
	"fmt"
	redis2 "github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")
	// 2.获取redis连接池的一个连接
	conn := redis.RedisPool().Get()
	defer conn.Close()
	// 3.获得文件句柄，用于存储分块内容
	fpath := "D:\\tmp\\" + uploadID + "\\" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 4.更新redis缓存数据
	conn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	// 5.返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler: 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	// 2.获得redis连接池的一个连接
	conn := redis.RedisPool().Get()
	defer conn.Close()
	// 3.通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis2.Values(conn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	// 4.TODO：合并上传

	// 5.更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	db.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	db.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))
	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

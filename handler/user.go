package handler

import (
	user_db "flie_store_server/db"
	"flie_store_server/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwd_salt = "*#890"
)

// SignUpHandler: 处理用户注册请求handler
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			fmt.Println("read sign page failed, err:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if len(username) < 3 || len(password) < 5 {
			w.Write([]byte("Invalid parameter"))
			return
		}
		enc_passwd := util.Sha1([]byte(password + pwd_salt))
		suc := user_db.UserSignUp(username, enc_passwd)
		if suc {
			w.Write([]byte("SUCCESS"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("FAILED"))
		}
	}
}

// SignInHandler: 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	encPassword := util.Sha1([]byte(password + pwd_salt))
	// 1.校验用户名，密码
	pwdChecked := user_db.UserSignin(username, encPassword)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}
	// 2.生成访问凭证（token）
	token := GenToken(username)
	upRes := user_db.UpdateToken(username, token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	// 3.生成登录成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// UserInfoHandler: 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1.解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	//token := r.Form.Get("token")
	// 2.验证token是否有效
	//isValidToken := isTokenValid(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	// 3.查询用户信息
	user, err := user_db.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 4.组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

func GenToken(username string) string {
	// md5 (username + timestamp + token_salt) + timestamp[:8] = 40 位
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

func isTokenValid(token string) bool {
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对于的token信息
	// TODO: 对比2个token是否一致
	if len(token) != 40 {
		return false
	}

	return true
}

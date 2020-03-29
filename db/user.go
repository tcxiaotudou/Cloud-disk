package db

import (
	mydb "flie_store_server/db/mysql"
	"fmt"
)

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignUp: 用户注册
func UserSignUp(username, password string) bool {
	stmt, err := mydb.DBConn().Prepare(`insert ignore into tbl_user (user_name, user_pwd) values (?, ?);`)
	if err != nil {
		fmt.Println("insert user failed, err:", err)
		return false
	}
	defer stmt.Close()
	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("exec insert user failed, err:", err)
		return false
	}
	if rowAffected, err := ret.RowsAffected(); rowAffected > 0 && err == nil {
		return true
	}
	return false
}

// UserSignin: 判断密码是否正确
func UserSignin(username, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare(`select * from tbl_user where user_name = ? limit 1;`)
	if err != nil {
		fmt.Println("select user info failed, err:", err)
		return false
	}
	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println("exec select user failed, err:", err)
		return false
	} else if rows == nil {
		fmt.Println("username not found: ", username)
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

// UpdateToken: 刷新用户登录的token
func UpdateToken(username, token string) bool {
	stmt, err := mydb.DBConn().Prepare(`replace into tbl_user_token (user_name, user_token) values (?, ?);`)
	if err != nil {
		fmt.Println("replace tbl_user_token failed, err:", err)
		return false
	}
	defer stmt.Close()
	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println("exec replace user_token failed, err:", err)
		return false
	}
	return true
}

// GetUserInfo: 用户信息查询
func GetUserInfo(username string) (User, error) {
	user := User{}
	stmt, err := mydb.DBConn().Prepare(`select user_name, signup_at from tbl_user where user_name = ? limit 1;`)
	if err != nil {
		fmt.Println("select userInfo failed, err:", err)
		return user, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		fmt.Println("query userInfo failed, err:", err)
		return user, err
	}
	return user, nil
}

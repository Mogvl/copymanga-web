package client

import "encoding/base64"

// encodePassword 对密码进行编码
// 拷贝漫画使用 base64(password-salt) 的方式
func encodePassword(password string) string {
	salt := "1729"
	data := password + "-" + salt
	return base64.StdEncoding.EncodeToString([]byte(data))
}

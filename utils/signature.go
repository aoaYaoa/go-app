package utils

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strconv"
	"strings"
	"time"
)

// GenerateSignature 生成API请求签名
func GenerateSignature(params map[string]string, appSecret string) string {
	// 按参数名排序
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
		signStr.WriteString("&")
	}
	signStr.WriteString("app_secret=")
	signStr.WriteString(appSecret)

	// 计算MD5签名
	hash := md5.New()
	hash.Write([]byte(signStr.String()))
	return hex.EncodeToString(hash.Sum(nil))
}

// GenerateAPIParams 生成API请求参数
func GenerateAPIParams(appKey string, appSecret string, params map[string]string) map[string]string {
	// 添加公共参数
	params["app_key"] = appKey
	params["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	params["nonce"] = GenerateNonce()

	// 生成签名
	params["sign"] = GenerateSignature(params, appSecret)

	return params
}

// GenerateNonce 生成随机字符串
func GenerateNonce() string {
	return time.Now().Format("20060102150405.000")
}

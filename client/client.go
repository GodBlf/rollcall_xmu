package client

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"log"
	"math/big"
	"net/http/cookiejar"
	"rollcall_xmu/logs"
	"strconv"
	"strings"
	"time"
)

type XMULogin struct {
	client *resty.Client
}

func NewXMULogin(userAgent string) *XMULogin {
	jar, _ := cookiejar.New(nil)
	c := resty.New().
		SetCookieJar(jar).
		// 缺省超时和重试策略可按需设置
		SetTimeout(30*time.Second).
		SetHeader("User-Agent", userAgent)

	return &XMULogin{client: c}
}

// 加密部分
func randomString(length int) (string, error) {
	chars := "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
	var b strings.Builder
	for i := 0; i < length; i++ {
		nBig, err := crand.Int(crand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		b.WriteByte(chars[nBig.Int64()])
	}
	return b.String(), nil
}

func pkcs7Pad(src []byte, blockSize int) []byte {
	padLen := blockSize - (len(src) % blockSize)
	return append(src, bytes.Repeat([]byte{byte(padLen)}, padLen)...)
}

func aesEncryptCBCBase64(plaintext, key, iv string) (string, error) {
	keyBytes := []byte(key)
	ivBytes := []byte(iv)
	plainBytes := []byte(plaintext)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}
	if len(ivBytes) != block.BlockSize() {
		return "", fmt.Errorf("invalid IV size: %d", len(ivBytes))
	}

	padded := pkcs7Pad(plainBytes, block.BlockSize())
	encrypted := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, ivBytes)
	mode.CryptBlocks(encrypted, padded)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func encryptPassword(password, salt string) string {
	salt = strings.TrimSpace(salt)
	if salt == "" {
		return password
	}

	randomPrefix, err := randomString(64)
	if err != nil {
		log.Printf("生成随机前缀失败: %v，回退为明文密码", err)
		return password
	}
	iv, err := randomString(16)
	if err != nil {
		log.Printf("生成IV失败: %v，回退为明文密码", err)
		return password
	}

	combined := randomPrefix + password
	enc, err := aesEncryptCBCBase64(combined, salt, iv)
	if err != nil {
		log.Printf("AES加密失败: %v，回退为明文密码", err)
		return password
	}
	return enc
}

// 登录部分
func (x *XMULogin) getLoginPage() (salt, execution, lt string, err error) {
	url := "https://ids.xmu.edu.cn/authserver/login"
	resp, err := x.client.R().Get(url)
	if err != nil {
		return "", "", "", fmt.Errorf("请求登录页失败: %w", err)
	}
	if resp.StatusCode() >= 400 {
		return "", "", "", fmt.Errorf("请求登录页返回错误状态码: %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(resp.String()))
	if err != nil {
		return "", "", "", fmt.Errorf("解析登录页失败: %w", err)
	}

	salt, boolean := doc.Find("#pwdEncryptSalt").Attr("value")
	if boolean == false {
		logs.Logger.Error(
			"加密盐抓取失败",
			zap.Error(err),
		)
	} else {
		logs.Logger.Info(
			"抓取加密盐成功",
			zap.String("encryp_salt", salt),
		)
	}
	execution, boolean = doc.Find("input[name='execution']").Attr("value")
	if boolean == false {
		logs.Logger.Error(
			"execution抓取失败",
			zap.Error(err),
		)
	} else {
		logs.Logger.Info(
			"execution抓取成功",
			zap.String("execution", execution),
		)
	}
	lt, _ = doc.Find("input[name='lt']").Attr("value")

	if salt == "" || execution == "" {
		return "", "", "", fmt.Errorf("未从登录页提取到必要字段 salt/execution")
	}
	return salt, execution, lt, nil
}

func (x *XMULogin) Login(username, password string) (bool, error) {
	salt, execution, lt, err := x.getLoginPage()
	if err != nil {
		return false, err
	}
	encPwd := encryptPassword(password, salt)

	form := map[string]string{
		"username":  username,
		"password":  encPwd,
		"captcha":   "",
		"_eventId":  "submit",
		"lt":        lt,
		"cllt":      "userNameLogin",
		"dllt":      "generalLogin",
		"execution": execution,
	}

	// 不跟随重定向以便判断 302
	x.client.
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Referer", "https://ids.xmu.edu.cn/authserver/login").
		SetFormData(form)
	resp, err := x.client.R().Post("https://ids.xmu.edu.cn/authserver/login")

	if err != nil {
		return false, fmt.Errorf("登录请求失败: %w", err)
	}

	if resp.StatusCode() == 200 {
		//location := resp.Header().Get("Location")
		location := resp.Request.URL
		log.Printf("登录成功，重定向到: %s", location)
		return true, nil
	}

	body := resp.String()
	if strings.Contains(body, "用户名或密码错误") || strings.Contains(body, "errorMessage") {
		return false, fmt.Errorf("登录失败：用户名或密码错误")
	}

	log.Printf("登录状态未知，响应码: %d，部分内容: %.200s", resp.StatusCode(), body)
	return false, fmt.Errorf("登录状态未知")
}

type RadarRollcall struct {
	RollcallStatus string `json:"rollcall_status"`
	Status         string `json:"status"`
	IsExpired      bool   `json:"is_expired"`
	CourseTitle    string `json:"course_title"`
	RollcallID     int    `json:"rollcall_id"`
}

type RadarResp struct {
	Rollcalls []RadarRollcall `json:"rollcalls"`
}

func (x *XMULogin) RollCallStatus() (map[string]int, error) {
	url := "https://lnt.xmu.edu.cn/api/radar/rollcalls?api_version=1.1.0"
	resp, err := x.client.R().Get(url)
	if err != nil {
		return nil, fmt.Errorf("查询签到状态失败: %w", err)
	}
	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("查询签到状态失败，状态码: %d", resp.StatusCode())
	}

	var r RadarResp
	if err := json.Unmarshal(resp.Body(), &r); err != nil {
		return nil, fmt.Errorf("解析签到状态JSON失败: %w", err)
	}

	pending := make(map[string]int)
	for _, rc := range r.Rollcalls {
		if rc.RollcallID != 0 {
			pending[rc.CourseTitle] = rc.RollcallID
		}

	}

	return pending, nil
}

type StudentRollcallResp struct {
	NumberCode string `json:"number_code"`
}

func (x *XMULogin) RollCallAnswer(rollcall map[string]int) (map[string]*string, error) {
	results := make(map[string]*string)
	for title, id := range rollcall {
		url := fmt.Sprintf("https://lnt.xmu.edu.cn/api/rollcall/%d/student_rollcalls", id)
		resp, err := x.client.R().Get(url)
		if err != nil {
			log.Printf("查询课程 '%s' 签到码请求失败: %v", title, err)
			results[title] = nil
			continue
		}
		if resp.StatusCode() >= 400 {
			log.Printf("查询课程 '%s' 签到码失败，状态码: %d", title, resp.StatusCode())
			results[title] = nil
			continue
		}
		var s StudentRollcallResp
		if err := json.Unmarshal(resp.Body(), &s); err != nil {
			log.Printf("解析课程 '%s' 签到码响应失败: %v", title, err)
			results[title] = nil
			continue
		}
		if s.NumberCode != "" {
			code := s.NumberCode
			results[title] = &code
			log.Printf("课程 '%s' 的签到码: %s", title, code)
		} else {
			log.Printf("课程 '%s' 未找到签到码", title)
			results[title] = nil
		}
	}
	return results, nil
}

func (x *XMULogin) RollCallAnswerTest(id int) error {
	url := fmt.Sprintf("https://lnt.xmu.edu.cn/api/rollcall/%d/student_rollcalls", id)
	resp, err := x.client.R().Get(url)
	if err != nil {
		return err
	}
	log.Printf("签到码查询响应状态码: %d", resp.StatusCode())

	var s StudentRollcallResp
	if err := json.Unmarshal(resp.Body(), &s); err != nil {
		return err
	}
	log.Printf("签到码: %s", s.NumberCode)
	return nil
}

func (x *XMULogin) AutoAnswerRollCall(course map[string]int, rollcall map[string]int, deviceId string) error {
	for courseName, courseRollCallId := range rollcall {
		m := make(map[string]string)
		m[strconv.Itoa(courseRollCallId)] = deviceId
		url := fmt.Sprintf("https://lnt.xmu.edu.cn/api/rollcall/%d/student_rollcalls", course[courseName])
		x.client.SetFormData(m)
		x.client.R().Post(url)
	}
	return errors.New("error")
}

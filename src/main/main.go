package main

import (
	"fmt"
	"github.com/spf13/viper"
	"go_projects1/client"
	"log"
	"time"
)

func loadConfig() (username, password, ua string, err error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err = viper.ReadInConfig(); err != nil {
		return "", "", "", fmt.Errorf("读取配置失败: %w", err)
	}
	username = viper.GetString("username")
	password = viper.GetString("password")
	ua = viper.GetString("user_agent")
	if ua == "" {
		ua = "Mozilla/5.0 (Linux; Android 6.0; Nexus 5 Build/MRA58N) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Mobile Safari/537.36"
	}
	if username == "" || password == "" {
		return "", "", "", fmt.Errorf("配置文件缺少 username/password")
	}
	return
}

func main() {
	username, password, ua, err := loadConfig()
	if err != nil {
		log.Fatalf("配置错误: %v", err)
	}

	x := client.NewXMULogin(ua)
	log.Println("开始模拟登录厦门大学统一认证系统...")

	ok, err := x.Login(username, password)
	if err != nil || !ok {
		log.Fatalf("登录失败: %v", err)
	}
	log.Println("登录成功！")

	//err = x.RollCallAnswerTest(141798)
	//time.Sleep(time.Second * 100)

	for {
		log.Println("=== 查询签到状态 ===")
		pending, err := x.RollCallStatus()
		if err != nil {
			log.Printf("查询签到状态失败: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if len(pending) == 0 {
			log.Println("当前没有需要签到的课程，2秒后重试")
			time.Sleep(2 * time.Second)
			continue
		}

		log.Println("=== 签到码查询操作 ===")
		results, err := x.RollCallAnswer(pending)
		if err != nil {
			log.Printf("签到码查询过程发生错误: %v", err)
		}

		log.Println("=== 签到结果总结 ===")
		for title, code := range results {
			if code != nil && *code != "" {
				log.Printf("✅ %s: 签到码 %s", title, *code)
			} else {
				log.Printf("❌ %s: 获取签到码失败", title)
			}
		}
		time.Sleep(200 * time.Second)
		break
	}
}

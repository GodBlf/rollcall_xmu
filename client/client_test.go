package client

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"log"
	"net/http/cookiejar"
	"rollcall_xmu/logs"
	"testing"
)

func Test_RollCallStatus(t *testing.T) {
	radarrest := &RadarResp{}
	radarrest.Rollcalls = append(radarrest.Rollcalls, RadarRollcall{})
	radarrest.Rollcalls[0].CourseTitle = "大学物理"
	radarrest.Rollcalls[0].RollcallID = 1234
	pending := make(map[string]int, 2)
	for _, rc := range radarrest.Rollcalls {
		if rc.RollcallID != 0 {
			pending[rc.CourseTitle] = rc.RollcallID
		}

	}
	for i, j := range pending {
		fmt.Printf("course: %s  id: %d",
			i,
			j)
	}

}

func TestYinyong(t *testing.T) {
	arr := make([]int, 10)
	arr = append(arr, 1, 2)
	for i, i2 := range arr {
		fmt.Printf("%d %d\n", i, i2)
	}
}
func TestRollCallAnswer(t *testing.T) {
	jar, _ := cookiejar.New(nil)
	client := resty.New()
	client.SetCookieJar(jar)
	url := fmt.Sprintf("https://lnt.xmu.edu.cn/api/rollcall/%d/student_rollcalls", 141798)
	resp, err := client.R().Get(url)
	if err != nil {
		logs.Logger.Error("签到码查询请求失败", zap.Error(err))
		return
	}
	log.Printf("签到码查询响应状态码: %d", resp.StatusCode())
	get := gjson.Get(resp.String(), "number_code")
	logs.Logger.Info(
		"签到码查询结果",
		zap.Int64("number_code", get.Int()),
	)

}

//todo: autorollcallanswer unittest

//todo: 雷达签到测试

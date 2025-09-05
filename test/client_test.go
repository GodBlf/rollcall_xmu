package test

import (
	"fmt"
	"rollcall_xmu/src/client"
	"testing"
)

func Test_RollCallStatus(t *testing.T) {
	radarrest := &client.RadarResp{}
	radarrest.Rollcalls = append(radarrest.Rollcalls, client.RadarRollcall{})
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

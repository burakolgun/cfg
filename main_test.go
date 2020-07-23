package cfg_test

import (
	"fmt"
	. "github.com/burakolgun/cfg"
	"testing"
	"time"
)

func TestInit(t *testing.T) {

	Init(Settings{
		Host:                 "https://stageconfigurationapi.trendyol.com",
		ProjectName:          "refund_consumer",
		IntervalTimeInSecond: 0,
		FirstTimeLoadRetryCount: 5,
	})

	fmt.Println(GetEnvironmentVariable("CONFIGURATION_API_URL"))

	<-Complete
	for {
		time.Sleep(time.Second * 1)
		fmt.Println(Get("CFG_TEST2").String())
	}
}
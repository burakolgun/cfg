package cfg_test

import (
	"fmt"
	. "github.com/burakolgun/cfg"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	expected := "settings must be valid"

	actual := Init(Settings{
		Host:  "",
		ProjectName:          "",
		IntervalTimeInSecond: 3,
	})

	if true == <-Complete {
		fmt.Println("done")
	}

	if actual == nil || actual.Error() != expected {
		time.Sleep(time.Minute * 1)
		t.Errorf("Init() = %q, want %q", actual, expected)
	}
}

func TestFirstLoadInformation(t *testing.T) {
}

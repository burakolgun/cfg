package cfg_test

import (
	. "github.com/burakolgun/cfg"
	"testing"
)

func TestInit(t *testing.T) {
	expected := "settings must be valid"

	actual := Init(Settings{
		Host:  "",
		ProjectName:          "",
		IntervalTimeInSecond: 3,
	})

	if actual == nil || actual.Error() != expected {
		t.Errorf("Init() = %q, want %q", actual, expected)
	}
}

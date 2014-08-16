package axmlParser

import (
	"fmt"
	"testing"
)

func TestParser(t *testing.T) {
	var filename = "/Users/lunny/Downloads/RyingMap.apk"

	listener := new(AppNameListener)
	_, err := ParseApk(filename, listener)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Init package is", listener.Name)
}

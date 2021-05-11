package golin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//
// PrintGoVersionList is download list printing
//
// インストール可能なバージョンリストを元に並び替えを行い表示します
// 存在するバージョンには「*」を表示します
//
func PrintGoVersionList() error {

	verList, err := createVersionList()
	if err != nil {
		return err
	}

	parent := filepath.Dir(os.Getenv("GOROOT"))
	gb := filepath.Join(parent, "*")

	matches, err := filepath.Glob(gb)
	if err != nil {
		return err
	}

	exists := make([]string, len(matches))
	for idx, ex := range matches {
		wk := strings.Replace(ex, parent, "", 1)
		exists[idx] = wk[1:]
	}

	//op := getOption()
	for _, ver := range verList {
		v := ver.String()

		for _, ex := range exists {
			if ex == v {
				v = v + strings.Repeat(" ", 20-len(v)) + "*"
				break
			}
		}
		fmt.Println(v)
	}

	return nil
}

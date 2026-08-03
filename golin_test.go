package golin_test

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/shizuokago/golin/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const testDir = "test_GOROOT"

var workROOT string

func TestMain(m *testing.M) {

	work := filepath.Join(getHome(), testDir)
	err := os.MkdirAll(work, 0777)
	if err == nil {
		workROOT = filepath.Join(work, "fake")

		ret := m.Run()

		err = os.RemoveAll(work)
		if err != nil {
			fmt.Printf("remove directory error[%v]\n", testDir)
		}
		os.Exit(ret)
	} else {
		fmt.Printf("make directory error[%v]\n", testDir)
		os.Exit(1)
	}
}

func TestGoEnv(t *testing.T) {

	goexe := golin.GetGoEnv("GOEXE")
	val := ""
	if runtime.GOOS == "windows" {
		val = ".exe"
	}

	if goexe != val {
		t.Errorf("GOEXE[%s] OS[%s]", goexe, runtime.GOOS)
	}

	//GOPATHはTestGoPathでテスト
}

func TestGoPath(t *testing.T) {

	org := os.Getenv("GOPATH")
	defer func(path string) {
		os.Setenv("GOPATH", path)
	}(org)

	if org != "" {
		gopath := golin.GetGoPath()
		if org != gopath {
			t.Errorf("GetGoPath error[%s] != [%s]", org, gopath)
		}
		err := os.Setenv("GOPATH", "")
		if err != nil {
			t.Logf("GOPATH set error")
		}
	}

	path := golin.GetGoPath()

	home := getHome()

	defPath := filepath.Join(home, "go")
	if path != defPath {
		t.Errorf("GetGoPath error[%s] != [%s]", path, defPath)
	}
}

func TestDownload(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping Download() test in short mode.")
	}

	sdk, err := golin.Download("1.12")
	if err != nil {
		t.Errorf("Download error[%v]", err)
	}

	defer func(path string) {
		os.RemoveAll(path)
	}(sdk)

	sdk, err = golin.Download("1.12")
	if err != nil {
		t.Errorf("Downloaded version error[%v]", err)
	}

	//work -> GOROOT

	sdk, err = golin.Download("1.7")
	if err == nil {
		t.Errorf("not download version [1.7](%s)", sdk)
	}
}

// TODO(secondarykey): Create()のテストを復活させる
//
// 元のテストはgolin.DefaultOption()/SetOption()で標準入力を差し替えていたが、
// これらのAPIは削除されており、現在getRoot()はos.Stdinを直接読む為、
// GOROOTの確認プロンプトを抜けられない。
// またGOROOTに実ディレクトリを設定していた為、readyLink()のos.Remove()が
// 空でないディレクトリに対して失敗し、API以前に前提が壊れている。
// 標準入力を注入できる形に戻すか、リンク名をテスト用に変更する必要がある。
func TestCreate(t *testing.T) {
	t.Skip("skipping Create() test: stdin injection API was removed.")
}

func BenchmarkParseVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		golin.NewVersion("1.12.1")
		golin.NewVersion("1.12beta1")
		golin.NewVersion("1.12rc1")
		golin.NewVersion("1.12")
		golin.NewVersion("2.0")
	}
}

func randomVersion() string {

	v := rand.Intn(3)
	r := rand.Intn(20)
	m := rand.Intn(10)

	mean := rand.Intn(3)

	buf := ""
	switch mean {
	case 0:
		buf = "."
	case 1:
		buf = "rc"
	case 2:
		buf = "beta"
	}

	rtn := fmt.Sprintf("%d.%d%s%d", v, r, buf, m)
	return rtn
}

func BenchmarkSortVersion(b *testing.B) {

	versionList := make([]*golin.Version, 10000)
	for i := 0; i < len(versionList); i++ {
		v := golin.NewVersion(randomVersion())
		versionList[i] = v
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sort.Slice(versionList, func(i, j int) bool {
			return versionList[i].Less(versionList[j])
		})
	}
}

// ExamplePrintは削除した
//
// golin.Print()は既に存在せず(現在はPrintGoVersionList(all bool))、
// 期待値もGo1.12時点のバージョン一覧で固定されていた為、
// go.dev/dlの内容に追従できずExampleとして成立しない。

// Test getHome()
// ユーザディレクトリの取得のメソッド
func getHome() string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

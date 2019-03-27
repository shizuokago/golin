package golin_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/shizuokago/golin"
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

func TestCreate(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping Create() test in short mode.")
	}

	sdk, err := golin.Download("1.12")
	if err != nil {
		t.Errorf("Downloaded version error[%v]", err)
	}

	//fakeの位置に移動
	err = os.Rename(sdk+string(filepath.Separator), workROOT+string(filepath.Separator))
	if err != nil {
		t.Logf("Rename Error[%v]", err)
	}

	org := os.Getenv("GOROOT")
	defer func(path string) {
		err := os.Setenv("GOROOT", path)
		if err != nil {
			t.Logf("setenv error[%v]", err)
		}
	}(org)

	err = os.Setenv("GOROOT", "")
	if err != nil {
		t.Logf("GOROOT set error")
	}

	op := golin.DefaultOption()
	op.StdIn = bytes.NewBufferString("Y\n")
	err = golin.Create("1.12")
	if err == nil {
		t.Errorf("GOROOT setting not error")
	}

	t.Logf("Work GOROOT[%s]", workROOT)
	err = os.Setenv("GOROOT", workROOT)
	if err != nil {
		t.Logf("GOROOT set error")
	}

	op.StdIn = bytes.NewBufferString("Y\n")
	golin.SetOption(op)
	err = golin.Create("1.12")
	if err != nil {
		t.Errorf("Create 1.12[%v]", err)
	}

	//switch
	op.StdIn = bytes.NewBufferString("Y\n")
	golin.SetOption(op)
	err = golin.Create("1.11")
	if err != nil {
		t.Errorf("Create 1.11[%v]", err)
	}

	//reswitch
	op.StdIn = bytes.NewBufferString("Y\n")
	golin.SetOption(op)
	err = golin.Create("1.12")
	if err != nil {
		t.Errorf("Create(exist) 1.12[%v]", err)
	}

	//Set Option Operation Y
}

func TestVersion(t *testing.T) {

	v18 := golin.NewVersion("1.8")
	v1110 := golin.NewVersion("1.11.0")
	v1116 := golin.NewVersion("1.11.6")
	v112 := golin.NewVersion("1.12.1")
	v2 := golin.NewVersion("2.0beta1")

	if !v18.Less(v1110) {
		t.Errorf("Version less error. 1.8 < 1.11.0")
	}
	if !v18.Less(v2) {
		t.Errorf("Version less error. 1.8 < 2.0")
	}

	if v1116.Less(v1110) {
		t.Errorf("Version less error. 1.11.6 > 1.11.0")
	}

	if !v1116.Less(v112) {
		t.Errorf("Version less error. 1.11.6 < 1.12")
	}

	if !v112.Less(v2) {
		t.Errorf("Version less error. 1.12 < 2.0")
	}

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

func ExamplePrint() {

	err := golin.Print()
	if err != nil {
	}

	// Output:
	// 1.8beta1
	// 1.8beta2
	// 1.8rc1
	// 1.8rc2
	// 1.8rc3
	// 1.8
	// 1.8.1
	// 1.8.2
	// 1.8.3
	// 1.8.4
	// 1.8.5
	// 1.8.6
	// 1.8.7
	// 1.9beta1
	// 1.9beta2
	// 1.9rc1
	// 1.9rc2
	// 1.9
	// 1.9.1
	// 1.9.2
	// 1.9.3
	// 1.9.4
	// 1.9.5
	// 1.9.6
	// 1.9.7
	// 1.10beta1
	// 1.10beta2
	// 1.10rc1
	// 1.10rc2
	// 1.10
	// 1.10.1
	// 1.10.2
	// 1.10.3
	// 1.10.4
	// 1.10.5
	// 1.10.6
	// 1.10.7
	// 1.10.8
	// 1.11beta1
	// 1.11beta2
	// 1.11beta3
	// 1.11rc1
	// 1.11rc2
	// 1.11
	// 1.11.1
	// 1.11.2
	// 1.11.3
	// 1.11.4
	// 1.11.5
	// 1.11.6
	// 1.12beta1
	// 1.12beta2
	// 1.12rc1
	// 1.12
	// 1.12.1
	//
}

// Test getHome()
// ユーザディレクトリの取得のメソッド
func getHome() string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

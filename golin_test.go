package golin_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shizuokago/golin"
)

const testDir = "test_GOROOT"

var workROOT string

func TestMain(m *testing.M) {

	work := filepath.Join(getHome(), testDir)
	err := os.MkdirAll(work, 0777)
	if err == nil {
		workROOT = filepath.Join(work, "fake")

		ret := m.Run()
		err = os.RemoveAll(workROOT)
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

	sdk, err := golin.Download("1.12")
	if err != nil {
		t.Errorf("Downloaded version error[%v]", err)
	}

	err = os.Rename(sdk+string(filepath.Separator), workROOT+string(filepath.Separator))
	if err != nil {
		t.Logf("Rename Error[%v]", err)
	}

	org := os.Getenv("GOROOT")
	defer func(path string) {
		err := os.Setenv("GOROOT", org)
		if err != nil {
			t.Logf("setenv error[%v]", err)
		}
	}(org)

	err = os.Setenv("GOROOT", "")
	if err != nil {
		t.Logf("GOROOT set error")
	}

	err = golin.Create("1.12")
	if err == nil {
		t.Errorf("GOROOT setting not error")
	}

	err = os.Setenv("GOROOT", workROOT)
	if err != nil {
		t.Logf("GOROOT set error")
	}

	//op := golin.DefaultOption()
	//op.StdIn = bytes.NewBuffer("N\n")
	//golin.SetOption(op)

	//Set Option Operation N
	//Set Option Operation Y
	err = golin.Create("1.12")
	if err != nil {
		t.Errorf("Create 1.12")
	}

	//switch
	err = golin.Create("1.11")
	if err != nil {
		t.Errorf("Create 1.11")
	}

	//current change

	//reswitch
	err = golin.Create("1.12")
	if err != nil {
		t.Errorf("Create 1.12")
	}

}

func getHome() string {
	home := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

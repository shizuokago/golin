package golin

import (
	"os"
	"os/exec"
	"testing"
)

func TestRun(t *testing.T) {

	pkgLinkName = "current"
	args := []string{"1.12rc1"}
	err := Run(args)
	if err != nil {
		t.Errorf("getRoot() Error")
	}
}

func TestCheckVersion(t *testing.T) {
	if !checkVersion() {
		t.Errorf("version error[%s]", pkgVersion)
	}
}

func TestGetRoot(t *testing.T) {

	root, err := getRoot()
	if err != nil {
		t.Errorf("getRoot() Error")
	}
	if root != "/usr/local/go" {
		t.Errorf("getRoot() get error[%s]", root)
	}

	//not found goroot

}

func TestGetPath(t *testing.T) {
	path := getPath()
	if path != "/home/secondarykey/go" {
		t.Errorf("getPath() error[%s]", path)
	}

	//exist gopath
}

func TestGoGet(t *testing.T) {

	pkgVersion = "1.9"
	cmd, err := goget()
	if err != nil {
		t.Errorf("goget() error[%v]", err)
	}

	_, err = os.Stat(cmd)
	if err != nil {
		t.Errorf("Stats Error[%v]", err)
	}

	os.Remove(cmd)
}

func TestDownload(t *testing.T) {

	pkgVersion = "1.9"
	cmd, err := goget()
	if err != nil {
		t.Errorf("goget() error[%v]", err)
	}

	sdk, err := download(cmd)
	if err != nil {
		t.Errorf("download() error[%v]", err)
	}

	_, err = os.Stat(sdk)
	if err != nil {
		t.Errorf("not downloaded sdk[%s]", sdk)
	}

	os.Remove(cmd)
	os.RemoveAll(sdk)

}

func TestCreateLink(t *testing.T) {
	dir := "/usr/local/go"

	pkgVersion = "1.8.1"
	pkgLinkName = "current"

	err := createLink(dir)
	if err != nil {
		t.Errorf("Error createLink()")
	}
}

func TestCmdRun(t *testing.T) {

	cmd := exec.Command("echo", "Hello")
	err := cmdRun(cmd)
	if err != nil {
		t.Errorf("not runing echo command")
	}

	cmd = exec.Command("unknown", "Hello")
	err = cmdRun(cmd)
	if err == nil {
		t.Errorf("runing unknown command")
	}

}

func TestGoEnv(t *testing.T) {
}

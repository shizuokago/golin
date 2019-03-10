package main

import (
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

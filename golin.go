package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	GoPrefix        = "go"
	DefaultLinkName = "current"
)

var (
	pkgVersion  string
	pkgLinkName string
)

func main() {

	flag.Parse()

	pkgLinkName = DefaultLinkName

	args := flag.Args()
	err := Run(args)
	if err != nil {
		fmt.Printf("Error:\n  %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Change current GOROOT")
	os.Exit(0)
}

func Run(args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("golin arguments required version")
	}

	pkgVersion = args[0]

	if !checkVersion() {
		return fmt.Errorf("this version not semantic version[%s]", pkgVersion)
	}

	root, err := getRoot()
	if err != nil {
		return err
	}

	return createLink(root)
}

func checkVersion() bool {
	return true
}

func getRoot() (string, error) {

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return "", fmt.Errorf("golin command required GOROOT environment variable.")
	}
	root := filepath.Dir(goroot)
	return root, nil
}

func createLink(dir string) error {

	path := filepath.Join(dir, pkgVersion)
	link := filepath.Join(dir, pkgLinkName)

	fmt.Println(link)
	if _, err := os.Lstat(link); err == nil {
		cmd := exec.Command("rm", link)
		if err := cmdRun(cmd); err != nil {
			return err
		}
	} else {
		fmt.Println("lstat error." + err.Error())
	}

	cmd := exec.Command("ln", "-ds", path, link)
	if err := cmdRun(cmd); err != nil {
		return err
	}

	return nil
}

func cmdRun(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

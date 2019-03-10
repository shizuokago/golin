package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	GoPrefix        = "go"
	DefaultLinkName = "current"
	DownloadLink    = "golang.org/dl"
)

var (
	pkgVersion  string
	pkgLinkName string
	stdErr      io.Writer
	stdOut      io.Writer
)

func main() {

	flag.Parse()

	pkgLinkName = DefaultLinkName
	stdOut = os.Stdout
	stdErr = os.Stderr

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

func getPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath != "" {
		return gopath
	}

	//Go Default gopath
	home := os.Getenv("HOME")
	return filepath.Join(home, "go")
}

func getSDK() string {
	home := os.Getenv("HOME")
	if home == "" {
		return ""
	}
	return filepath.Join(home, "sdk", "go"+pkgVersion)
}

func goget() (string, error) {
	link := fmt.Sprintf("%s/go%s", DownloadLink, pkgVersion)

	cmd := exec.Command("go", "get", link)
	err := cmdRun(cmd)
	if err != nil {
		return "", err
	}

	genCmd := fmt.Sprintf("go%s", pkgVersion)
	genPath := filepath.Join(getPath(), "bin", genCmd)
	return genPath, nil
}

func download(bin string) (string, error) {
	cmd := exec.Command(bin, "download")
	err := cmdRun(cmd)
	if err != nil {
		return "", err
	}
	return getSDK(), nil
}

func createLink(dir string) error {

	path, err := readyPath(dir)
	if err != nil {
		return err
	}

	link, err := readyLink(dir)
	if err != nil {
		return err
	}

	cmd := exec.Command("ln", "-ds", path, link)
	if err := cmdRun(cmd); err != nil {
		return err
	}

	return nil
}

func readyLink(dir string) (string, error) {
	link := filepath.Join(dir, pkgLinkName)
	if _, err := os.Lstat(link); err == nil {
		cmd := exec.Command("rm", link)
		if err := cmdRun(cmd); err != nil {
			return "", err
		}
	} else {
		//first run?
		return "", err
	}
	return link, nil
}

func readyPath(dir string) (string, error) {

	path := filepath.Join(dir, pkgVersion)
	_, err := os.Stat(path)
	if err == nil {
		return path, nil
	}

	bin, err := goget()
	if err != nil {
		return "", err
	}
	defer os.Remove(bin)

	sdk, err := download(bin)
	if err != nil {
		return "", err
	}

	fmt.Println("SDK=" + sdk)
	fmt.Println("PATH=" + path)
	//move
	cmd := exec.Command("mv", sdk, path)
	err = cmdRun(cmd)
	if err != nil {
		return "", err
	}
	return path, nil
}

func cmdRun(cmd *exec.Cmd) error {

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

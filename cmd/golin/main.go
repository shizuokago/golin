//Golin Command
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shizuokago/golin"
)

// golin設定用のオプション
var op *golin.Option

// Initialize golin command
//
// オプションに-dでリンク名を変更できるようにし、Usageを設定する
func init() {
	op = golin.DefaultOption()
	flag.StringVar(&op.LinkName, "d", op.LinkName, "symbolic link name")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage of golin:

    It is possible to switch by setting GOROOT to a symbolic link.
    A list of downloads is available at the link below.

        https://github.com/golang/dl

    Please remove the "go" for the specification of the version.
    (ex: go1.12.1 -> 1.12.1

        golin 1.12.1

    There are special arguments "list" and "development".
    "list" indicates the specifiable version
    “development” compiles the latest version(gotip).
`)
		flag.PrintDefaults()
	}
}

// 特殊引数
//
// list でダウンロードできるバージョンのリストを表示
// development で最新の開発バージョンを取得
const (
	DownloadList = "list"
	Development  = "development"
)

//
// This golin command main
//
func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Printf("golin arguments required version")
		os.Exit(1)
	}

	var err error
	arg := args[0]
	golin.SetOption(op)

	switch arg {
	case DownloadList:
		err = golin.Print()
	case Development:
		err = golin.CompileLatestSDK()
	default:
		err = golin.Create(arg)
	}

	if err != nil {
		fmt.Printf("Error:\n  %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

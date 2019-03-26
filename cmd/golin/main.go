//Golin Command
//
//golin install command
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shizuokago/golin"
)

var op *golin.Option

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

//
// This command main
//
// コマンド実行時の初期処理です
// 標準出力、標準エラーはそのままosの値を設定しています
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
	case golin.DownloadList:
		err = golin.Print()
	case golin.Development:
		err = golin.Create("tip")
	default:
		err = golin.Create(arg)
	}

	if err != nil {
		fmt.Printf("Error:\n  %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

//Thid method is empty
func Empty() {
}

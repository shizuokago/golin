package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shizuokago/golin"
)

//
// This command main
//
// コマンド実行時の初期処理です
// 標準出力、標準エラーはそのままosの値を設定しています
//
//
func main() {

	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		fmt.Printf("golin arguments required version")
		os.Exit(1)
	}

	err := golin.Create(args[0])
	if err != nil {
		fmt.Printf("Error:\n  %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

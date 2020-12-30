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

var (
	version  string
	revision string
	date     string
	build    string
)

// Initialize golin command
//
// オプションに-dでリンク名を変更できるようにし、Usageを設定する
func init() {
	op = golin.DefaultOption()
	flag.StringVar(&op.LinkName, "d", op.LinkName, "symbolic link name")
	flag.Usage = Usage
}

// 特殊引数
//
// list でダウンロードできるバージョンのリストを表示
// development で最新の開発バージョンを取得
const (
	Version         = "version"
	Install         = "install"
	DownloadList    = "list"
	Development     = "dev"
	ReleaseCompress = "compress"
)

//
// This golin command main
//
func main() {

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("golin arguments required version")
		os.Exit(1)
	}

	var err error
	arg := args[0]
	golin.SetOption(op)

	switch arg {
	case Version:
		err = printVersion()
	case DownloadList:
		err = golin.PrintGoVersionList()
	case Development:
		err = golin.CompileLatestSDK()
	case Install:
		path := args[1]
		err = golin.Install(path)
	case ReleaseCompress:
		path := args[1]
		exe := args[2]
		err = golin.CompressReleaseZip(path, exe)
	default:
		err = golin.Create(arg)
	}

	if err != nil {
		fmt.Printf("Error:\n  %+v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func printVersion() error {
	if version == "" || revision == "" || date == "" || build == "" {
		return fmt.Errorf("version is empty.")
	}
	fmt.Printf("golin version %s %s\nBuild Information:%s (%s)\n",
		version, build, date, revision)
	return nil
}

func Usage() {
	help := `Usage of golin:

  It is possible to switch by setting GOROOT to a symbolic link.
  A list of downloads is available at the link below.

      https://github.com/golang/dl

  まだGoが存在しない場合、

      golin install {path}

  これにより最新のGoがインストールされます。
  {path}はGOROOTの元になる位置を指定します。

     e.g) golin install /usr/local/go

  これを行うことで{path}/{version}にGoが設定され、
  そのGoに対して{path}/current にシンボリックリンクを作成します。

  あなたはその後、{path}/currentに対して、GOROOTの環境変数を設定する必要があります。

  Goのバージョンの切り替えは
  Please remove the "go" for the specification of the version.
    (ex: go1.12.1 -> 1.12.1

      golin 1.12.1

  現在インストール可能なGoのバージョンと、インストールされているバージョンは

      golin list
    
  を実行することで一覧で表示されます。

  現在開発中の最新バージョン(gotip)を手に入れる場合

      golin dev

  を行うとビルドして更新されます（少し時間がかかります。
  devはビルドしたバージョンの有無ではなく、常に最新のビルドを行いますが

     golin tip

  はdevでビルドしたバージョンが存在する場合、切り替えるのみで終了します

  また-d を指定することでcurrentを変更することができます

     e.g.) golin -d root 1.16

  これにより切り替え先のシンボリックリンクが{path}/rootになりますので、
  そこをGOROOTに指定してください。
`
	fmt.Fprintf(os.Stderr, help)
	flag.PrintDefaults()
}

//Golin Command
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shizuokago/golin"
	"github.com/shizuokago/golin/config"
)

var (
	version  string
	revision string
	date     string
	build    string
)

var (
	link string
)

// Initialize golin command
//
// オプションに-dでリンク名を変更できるようにし、Usageを設定する
func init() {
	flag.StringVar(&link, "d", config.DefaultLinkName, "symbolic link name")
	flag.Usage = Usage
}

type Cmd string

//
// golinコマンド
//
// version  バージョン表示
// install  Goのインストール
// list     ダウンロードできるバージョンのリストを表示
// dev      最新の開発バージョンを取得
// compress コマンド等の圧縮(リリース用)
//
const (
	Version         Cmd = "version"
	Install         Cmd = "install"
	DownloadList    Cmd = "list"
	Development     Cmd = "dev"
	ReleaseCompress Cmd = "compress"
	//バージョン指定を行っている場合の文字列
	ChangeVersion Cmd = ""
)

//
// This golin command main
//
func main() {

	err := run()

	if err != nil {
		fmt.Fprintf(os.Stderr, "golin Error: %+v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		return fmt.Errorf("golin arguments required command(install,list,dev,version) or version(e.g. 1.15.6,1.16beta1).")
	}

	cmd := Cmd(args[0])
	//cmd = ChangeVersion

	opts := make([]config.Option, 1)
	opts[0] = config.SetLinkName(link)

	err := config.Set(opts...)
	if err != nil {
		return fmt.Errorf("config.Set() error: %w", err)
	}

	switch cmd {
	case Version:
		//コマンドのバージョン表示
		err = printVersion()
		//バージョン表示のみで終了(Successを表示しない)
		return nil
	case DownloadList:
		//ダウンロードのリスト表示
		err = golin.PrintGoVersionList()
	case Development:
		//開発バージョンのコンパイル
		err = golin.CompileLatestSDK()
	case Install:
		if len(args) < 2 {
			return fmt.Errorf("golin install arguments required path")
		}
		path := args[1]
		v := ""
		if len(args) >= 3 {
			v = args[2]
		}
		//インストールを行う
		err = golin.Install(path, v)
	case ReleaseCompress:
		if len(args) < 3 {
			return fmt.Errorf("golin compress arguments required filename and command name.")
		}
		dst := args[1]
		src := args[2]
		//リリース用のZip作成
		err = golin.CompressReleaseZip(dst, src)
	default:
		if len(args) < 1 {
			return fmt.Errorf("golin arguments required version(e.g. 1.15.6, 1.16beta1).")
		}
		v := args[0]
		//バージョンの変更
		err = golin.Create(v)
	}

	if err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	fmt.Println("Success.")
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
  install時は-vを指定することでバージョンを最新以外で選択できます。

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

func printVersion() error {
	if version == "" || revision == "" || date == "" || build == "" {
		return fmt.Errorf("version is empty.")
	}
	fmt.Printf("golin version %s %s\nBuild Information:%s (%s)\n", version, build, date, build)
	return nil
}

//
//Commad golin is Switching the symbolic link of GOROOT
//
//んなもんDockerでやりゃいい！という思いを跳ね除け、
//Shizuoka.goの為に作りましたが、多分secondarykeyはそのままつかいます
//https://github.com/shizuokago/golin で管理しています
//
//Reference
//
//versionが対象ディレクトリに存在しない場合、自動的にダウンロードを行い、
//バージョンの切り替えを行ってくれます
//
//Windowsも対応予定です
//
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

//定数群
const (
	GoPrefix        = "go"            //コマンドなどのPrefix
	DefaultLinkName = "current"       //作成するリンク名
	DownloadLink    = "golang.org/dl" //ダウンロード時のリンク先
)

// パッケージ内で使用する変数
// 引数で渡すのも考えたけど。だるいので変数化した
// Run()の挙動が変わるのでOptionとして構造体化するべきかも
var (
	pkgVersion  string    //切替対象のバージョン
	pkgLinkName string    //リンク名
	stdErr      io.Writer //エラー時の出力場所
	stdOut      io.Writer //出力場所
)

//
// This command main()
//
// コマンド実行時の初期処理です
// flag指定がない為、現状はDefaultLinkName でリンクを作成
// 標準出力、標準エラーはそのままosの値を設定しています
//
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

//
// Function Run is Command Controller
//
// args[0]に指定バージョンがあり、大域変数である
// pkgLinkName,stdOut,stdErrに設定を行って呼び出します
// pkgLinkNameは後日flag指定の予定です
//
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

//
// Function checkVersion is Version Check
//
// 現状はバージョン指定のチェックを行っていませんが、
// X.xx.xx形式をチェック仕様かと思っています
// ただし、BetaやReleaseCandidateがあるので
// Semantic versioningだけのチェックは適用できない
//
// TODO(secondarykey) : Not yet implemented
//
func checkVersion() bool {
	return true
}

//
// Function getRoot() is return Work Directory Path
//
// この関数は処理対象のディレクトリを返します。
// 具体的には現在のGOROOTの上の階層を返します。
// GOROOTが存在しない場合はエラーとなります
//
func getRoot() (string, error) {

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return "", fmt.Errorf("golin command required GOROOT environment variable.")
	}
	root := filepath.Dir(goroot)
	return root, nil
}

// Function getPath() is return GOPATH
//
// GOPATHの値を返しますが、
// 設定がない場合もあるのでos.Getenv()ではなく
// go env からの値を取得
//
func getPath() string {
	return getGoEnv("GOPATH")
}

//
// Function getSDK() is return Download path
//
// golang.org/dl/goX.x.x のコマンドにdownloadした場合の
// パスを作成して返します
// golang.org/dl/internal/version.go の仕様が変更になった場合、
// 変更する必要があります
//
func getSDK() string {
	home := getHome()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "sdk", "go"+pkgVersion)
}

//
// Function goget() is downlod command download(go get)
//
// Goをダウンロードするコマンドである
// golang.org/dl/goX.x.x をgo getして取得してきます
// そのままGOPATHの位置でinstallされ、コマンドが作成されますので
// そのコマンド名も返します
//
func goget() (string, error) {
	link := fmt.Sprintf("%s/go%s", DownloadLink, pkgVersion)

	cmd := exec.Command("go", "get", link)
	err := runCmd(cmd)
	if err != nil {
		return "", err
	}

	genCmd := fmt.Sprintf("go%s%s", pkgVersion, getGoEnv("GOEXE"))
	genPath := filepath.Join(getPath(), "bin", genCmd)
	return genPath, nil
}

//
// Function download() is Golang download
//
// 渡された引数を元にGo言語をダウンロードしてきます
// 戻り値はダウンロードして来たディレクトリを返します
//
func download(bin string) (string, error) {
	cmd := exec.Command(bin, "download")
	err := runCmd(cmd)
	if err != nil {
		return "", err
	}
	return getSDK(), nil
}

//
// Function createLink() is create symboliclink
//
// 指定されたディレクトリにシンボリックリンクを作成します
// readyPath() で指定したバージョンのGo言語の準備
// readyLink() で指定したリンクの準備(削除) を行います
// 最後にlnでシンボリックリンクを作成します
//
func createLink(dir string) error {

	path, err := readyPath(dir)
	if err != nil {
		return err
	}

	link, err := readyLink(dir)
	if err != nil {
		return err
	}

	cmd := createLinkCmd(path, link)
	if err := runCmd(cmd); err != nil {
		return err
	}

	return nil
}

//
// Function readyLink() is remove symbolic link
//
// シンボリックリンクは存在する場合の
// コマンドの動作が違うので削除を行っておきます
// 現状初回起動時にシンボリックリンクがない場合に
// 問い合わせしたりする処理がありません
//
// BUG(secondarykey): ロールバックがない
//
func readyLink(dir string) (string, error) {
	link := filepath.Join(dir, pkgLinkName)
	if _, err := os.Lstat(link); err == nil {
		cmd := createRemoveCmd(link)
		if err := runCmd(cmd); err != nil {
			return "", err
		}
	} else {
		//first run?
		return "", err
	}
	return link, nil
}

//
// Function readyPath() is golang version path
//
// 対象バージョンのパスを確認し、
// 存在しない場合はダウンロードを行って準備する
// 存在するバージョンの場合はそのままパスを返す
//
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

	//move
	cmd := createMoveCmd(sdk, path)
	err = runCmd(cmd)
	if err != nil {
		return "", err
	}
	return path, nil
}

//
// Function runCmd() is command running
//
// 実際コマンドを実行する処理
// 標準出力等を一括管理する為に関数化を行った
//
func runCmd(cmd *exec.Cmd) error {

	cmd.Stdout = stdOut
	cmd.Stderr = stdErr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

//
// Function getGoEnv() is go env {key} command
//
func getGoEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		return ""
	}
	return string(out)
}

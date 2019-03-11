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

package golin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//定数群
const (
	goPrefix        = "go"            //コマンドなどのPrefix
	defaultLinkName = "current"       //作成するリンク名
	downloadLink    = "golang.org/dl" //ダウンロード時のリンク先
)

//特殊引数
//
// list でダウンロードできるバージョンのリストを表示
// development で最新の開発バージョンを取得D
//
// TODO(secondarykey) : Not yet Implemented
//
const (
	DownloadList = "list"
	Development  = "development" //gotip
)

//実行オプション
type Option struct {
	LinkName string    //リンク名
	StdIn    io.Reader //エラー時の出力場所
	StdErr   io.Writer //エラー時の出力場所
	StdOut   io.Writer //出力場所
}

var option *Option

func SetOption(op *Option) {
	option = op
}

func getOption() *Option {
	if option == nil {
		option = &Option{
			LinkName: defaultLinkName,
			StdIn:    os.Stdin,
			StdOut:   os.Stdout,
			StdErr:   os.Stderr,
		}
	}
	return option
}

//
// Function Run is Command Controller
//
// args[0]に指定バージョンがあり、大域変数である
// pkgLinkName,stdOut,stdErrに設定を行って呼び出します
// pkgLinkNameは後日flag指定の予定です
//
func Run(v string) error {

	if !checkVersion(v) {
		return fmt.Errorf("this version not semantic version[%s]", v)
	}

	root, err := getRoot()
	if err != nil {
		return err
	}

	path, err := readyPath(root, v)
	if err != nil {
		return err
	}

	link, err := readyLink(root)
	if err != nil {
		return err
	}

	err = os.Symlink(path, link)
	if err != nil {
		return err
	}

	return nil
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
func checkVersion(v string) bool {
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
func getSDKPath(v string) string {
	home := getHome()
	if home == "" {
		return ""
	}
	return filepath.Join(home, "sdk", "go"+v)
}

//
// Function goget() is downlod command download(go get)
//
// Goをダウンロードするコマンドである
// golang.org/dl/goX.x.x をgo getして取得してきます
// そのままGOPATHの位置でinstallされ、コマンドが作成されますので
// そのコマンド名も返します
//
func goget(v string) (string, error) {

	link := fmt.Sprintf("%s/go%s", downloadLink, v)

	// go get golang.org/dl/go{version}
	cmd := exec.Command("go", "get", link)
	err := runCmd(cmd)
	if err != nil {
		return "", err
	}

	//GOEXE windows excutable file extention
	// go{version}{.exe}
	genCmd := fmt.Sprintf("go%s%s", v, getGoEnv("GOEXE"))
	genPath := filepath.Join(getPath(), "bin", genCmd)
	return genPath, nil
}

//
// Function download() is Golang download
//
// 渡された引数を元にGo言語をダウンロードしてきます
// 戻り値はダウンロードして来たディレクトリを返します
//
func download(bin, v string) (string, error) {

	// $GOPATH/bin/go{version}{.exe} download
	cmd := exec.Command(bin, "download")
	err := runCmd(cmd)
	if err != nil {
		return "", err
	}
	return getSDKPath(v), nil
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

	link := filepath.Join(dir, getOption().LinkName)
	//symbliclink
	if _, err := os.Lstat(link); err == nil {
		err = os.Remove(link)
		if err != nil {
			return "", err
		}

	} else {
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
func readyPath(dir, v string) (string, error) {

	path := filepath.Join(dir, v)
	_, err := os.Stat(path)
	//Exist
	if err == nil {
		return path, nil
	}

	bin, err := goget(v)
	if err != nil {
		return "", err
	}
	//delete exe file
	defer os.Remove(bin)

	//go download
	sdk, err := download(bin, v)
	if err != nil {
		return "", err
	}

	err = os.Rename(sdk+string(filepath.Separator), path+string(filepath.Separator))
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

	op := getOption()
	cmd.Stdout = op.StdOut
	cmd.Stderr = op.StdErr

	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

//
// Function getGoEnv() is go env {key} command
// TODO(secondarykey) : change replaceall
//
func getGoEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		return ""
	}
	return strings.Replace(string(out), "\n", "", -1)
}

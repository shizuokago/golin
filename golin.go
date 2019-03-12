//
// Commad golin is Switching the symbolic link of GOROOT
//
// んなもんDockerでやりゃいい！という思いを跳ね除け、
// Shizuoka.goの為に作りましたが、多分secondarykeyはそのままつかいます
// https://github.com/shizuokago/golin で管理しています
//
// versionが対象ディレクトリに存在しない場合、自動的にダウンロードを行い、
// バージョンの切り替えを行ってくれます
//

package golin

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//定数
const (
	defaultLinkName = "current"       //作成するリンク名
	downloadLink    = "golang.org/dl" //ダウンロード時のリンク先
	workDirectory   = "golin_work"
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

func DefaultOption() *Option {
	return &Option{
		LinkName: defaultLinkName,
		StdIn:    os.Stdin,
		StdOut:   os.Stdout,
		StdErr:   os.Stderr,
	}
}

func SetOption(op *Option) {
	option = op
}

func getOption() *Option {
	if option == nil {
		option = DefaultOption()
	}
	return option
}

//
// Create is create symblic link
//
// 引数でバージョンを指定します
// GOROOTの確認、権限の確認、パスの準備、リンクの準備(削除)
// リンクの張り直しを行います
//
func Create(v string) error {

	printGoVersion("Before:")

	root, err := getRoot()
	if err != nil {
		return err
	}

	err = checkAuthorization(root)
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

	printGoVersion("After :")

	return nil
}

//
// checkAuthorization is authorization check
//
// 引数のパスにリンクが貼れるかをワークでチェック
//
func checkAuthorization(path string) error {

	work := filepath.Join(path, "."+workDirectory)
	err := os.Mkdir(work, 0777)
	if err != nil {
		return err
	}
	defer os.Remove(work)

	link := filepath.Join(path, "_"+workDirectory+"_")
	err = os.Symlink(work, link)
	if err != nil {
		return err
	}
	defer os.Remove(link)

	return nil
}

//
// checkVersion is Version Check
//
// バージョンのリストを作成し、
// 指定バージョンがダウンロード可能かを確認
//
func checkVersion(v string) bool {

	versionList, err := getVersionList()
	if err != nil {
		fmt.Println(err)
		return false
	}

	for _, ver := range versionList {
		if ver == v {
			return true
		}
	}

	return false
}

//
// Function getRoot is return Work Directory Path
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

//
// GetGoPath is return GOPATH
//
// GOPATHの値を返しますが、
// 設定がない場合もあるのでos.Getenv()ではなくgo env からの値を取得
//
func GetGoPath() string {
	return GetGoEnv("GOPATH")
}

//
// getSDKPath is return Downloaded path
//
// golang.org/dl/goX.x.x のコマンドにdownloadした場合のパスを作成して返します
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
// getVersionList is return Download list
//
// ダウンロードのリポジトリをgo getし、
// ディレクトリ名からダウンロード可能なバージョンのリストを作成
//
func getVersionList() ([]string, error) {

	// go get golang.org/dl/
	cmd := exec.Command("go", "get", "-u", downloadLink)
	// no go files error
	runCmd(cmd)

	dir := filepath.Join(GetGoPath(), "src", filepath.Clean(downloadLink))

	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	versionList := make([]string, 0, len(infos))
	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		name := info.Name()
		if name != "internal" && name != ".git" {
			ver := strings.Replace(name, "go", "", -1)
			versionList = append(versionList, ver)
		}
	}

	return versionList, nil
}

//
// createDownloadCmd is downlod command download(go get)
//
// Goをダウンロードするコマンドである
// golang.org/dl/goX.x.x をgo getして取得してきます
// そのままGOPATHの位置でinstallされ、コマンドが作成されますので
// そのコマンド名も返します
//
func createDownloadCmd(v string) (string, error) {

	link := fmt.Sprintf("%s/go%s", downloadLink, v)

	// go get golang.org/dl/go{version}
	cmd := exec.Command("go", "get", link)
	err := runCmd(cmd)
	if err != nil {
		return "", err
	}

	//GOEXE windows excutable file extention
	// go{version}{.exe}
	genCmd := fmt.Sprintf("go%s%s", v, GetGoEnv("GOEXE"))
	genPath := filepath.Join(GetGoPath(), "bin", genCmd)
	return genPath, nil
}

//
// Download is Go download
//
// Go言語をダウンロードします
// 戻り値はダウンロードして来たディレクトリを返します
//
func Download(v string) (string, error) {

	//$GOPATH/bin/go{version}{.exe}
	bin, err := createDownloadCmd(v)
	if err != nil {
		return "", err
	}
	//delete exe file
	defer os.Remove(bin)

	// $GOPATH/bin/go{version}{.exe} download
	cmd := exec.Command(bin, "download")
	err = runCmd(cmd)
	if err != nil {
		return "", err
	}
	return getSDKPath(v), nil
}

//
// readyLink is remove symbolic link
//
// シンボリックリンクは存在する場合の
// コマンドの動作が違うので削除を行っておきます
// 現状初回起動時にシンボリックリンクがない場合に
// 問い合わせしたりする処理がありません
//
// BUG(secondarykey): 作成に失敗した場合のロールバックがない
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
		// Confirm
		//return "", err
	}
	return link, nil
}

//
// readyPath is golang version path
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

	if !checkVersion(v) {
		return "", fmt.Errorf("this version not exist download version[%s]", v)
	}

	//go download
	sdk, err := Download(v)
	if err != nil {
		return "", err
	}

	//Download SDK Rename
	err = os.Rename(sdk+string(filepath.Separator), path+string(filepath.Separator))
	if err != nil {
		return "", err
	}
	return path, nil
}

//
// runCmd is command running
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
// GetGoEnv is go env {key} command
//
// go envを引数で実行します
//
// TODO(secondarykey) : change replaceall(1.12 after,,,)
//
func GetGoEnv(key string) string {
	out, err := exec.Command("go", "env", key).Output()
	if err != nil {
		return ""
	}
	return strings.Replace(string(out), "\n", "", -1)
}

//
// printGoVersion is current go command version
//
// go versionを実行します
//
func printGoVersion(prefix string) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return
	}

	ver := strings.Replace(string(out), "\n", "", -1)

	op := getOption()
	fmt.Fprintln(op.StdOut, prefix, ver)
}

package golin

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/xerrors"
)

//定数
const (
	defaultLinkName    = "current"                      //作成するリンク名
	downloadLink       = "golang.org/dl"                //ダウンロード時のリンク先
	workDirectory      = "golin_work"                   //権限確認用のディレクトリ名
	GitHubDownloadPage = "https://github.com/golang/dl" //GitHub上のバージョンリスト
	GolangDownloadPage = "https://golang.org/dl"        //
)

func Install(path string) error {

	//権限の確認
	err := checkAuthorization(path)
	if err != nil {
		return xerrors.Errorf("Authorization error: %w", err)
	}

	// 最終バージョンを取得
	v, err := getLatestVersion()
	if err != nil {
		return xerrors.Errorf("getLatestVersion() error: %w", err)
	}

	// そのバージョンをダウンロードし展開
	url := fmt.Sprintf("%s/go%s.%s-%s.%s", GolangDownloadPage, v.String(), runtime.GOOS, runtime.GOARCH, getDownloadExt())

	fmt.Println("Download Latest Version:", url)

	dp := filepath.Join(path, v.String())
	//作成
	err = DecompressURL(url, dp)
	if err != nil {
		return xerrors.Errorf("DecompressURL() error: %w", err)
	}

	//currentを作成
	link, err := readyLink(path)
	if err != nil {
		return xerrors.Errorf("readyLink() error: %w", err)
	}

	//シンボリックリンクを作成
	err = os.Symlink(dp, link)
	if err != nil {
		return xerrors.Errorf("symlink: %w", err)
	}

	// 各OSに合わせた設定手順を表示
	printSetting(link, v.String())

	return nil
}

//
// Create is create symblic link
//
// 引数でバージョンを指定します
// GOROOTの確認、権限の確認、パスの準備、リンクの準備(削除)
// リンクの張り直しを行います
//
// BUG(secondarykey): テストがGo1.12にしてないと通らない
//
func Create(v string) error {

	var err error
	root := v

	//ルートを取得
	root, err = getRoot(v)
	if err != nil {
		return xerrors.Errorf("getRoot() error: %w", err)
	}
	//権限チェック
	err = checkAuthorization(root)
	if err != nil {
		return xerrors.Errorf("authorization error: %w", err)
	}

	//設定前のGoのバージョン表示
	printGoVersion("Before:")

	//指定バージョンでパスを作成
	path, err := readyPath(root, v)
	if err != nil {
		return xerrors.Errorf("ready path: %w", err)
	}

	//シンボリックを準備
	link, err := readyLink(root)
	if err != nil {
		return xerrors.Errorf("ready link: %w", err)
	}

	//シンボリックリンクを作成
	err = os.Symlink(path, link)
	if err != nil {
		return xerrors.Errorf("symlink: %w", err)
	}

	//終了したバージョンを作成
	printGoVersion("After :")

	return nil
}

//
// CompileGoSDK is Compile from the latest repository to Create GoSDK
//
// Createに"tip"を渡すことで開発用のgotipの実行を行います
//
func CompileLatestSDK() error {
	return Create("tip")
}

//
// PrintGoVersionList is download list printing
//
// インストール可能なバージョンリストを元に並び替えを行い表示します
// 存在するバージョンには「*」を表示します
//
func PrintGoVersionList() error {

	verList, err := createVersionList()
	if err != nil {
		return err
	}

	parent := filepath.Dir(os.Getenv("GOROOT"))
	gb := filepath.Join(parent, "*")

	matches, err := filepath.Glob(gb)
	if err != nil {
		return err
	}

	exists := make([]string, len(matches))
	for idx, ex := range matches {
		wk := strings.Replace(ex, parent, "", 1)
		exists[idx] = wk[1:]
	}

	//op := getOption()
	for _, ver := range verList {
		v := ver.String()

		for _, ex := range exists {
			if ex == v {
				v = v + strings.Repeat(" ", 20-len(v)) + "*"
				break
			}
		}

		fmt.Println(v)
	}

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
		return xerrors.Errorf("そのディレクトリに権限がありません。: %w", err)
	}
	defer os.Remove(work)

	link := filepath.Join(path, "_"+workDirectory+"_")
	err = os.Symlink(work, link)
	if err != nil {
		return xerrors.Errorf("シンボリックリンクの作成に失敗しました。", err)
	}
	defer os.Remove(link)

	return nil
}

//
// Function getRoot is return Work Directory Path
//
// この関数は処理対象のディレクトリを返します。
// 具体的には現在のGOROOTの上の階層を返します。
// GOROOTが存在しない場合はエラーとなります
//
func getRoot(ver string) (string, error) {

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		return "", fmt.Errorf("golin command required GOROOT environment variable.")
	}

	op := getOption()
	root := filepath.Dir(goroot)
	now := filepath.Base(goroot)

	idx := strings.Index(goroot, op.LinkName)

	//最後がリンク名と同一かを見る
	if idx != len(goroot)-len(op.LinkName) {
		fmt.Fprintf(op.StdOut, `
This command creates the Go SDK within the current GOROOT parent directory. 
It is recommended to specify a dedicated directory.

%s -
   |- %s
   |- %s [Download specified Go SDK]
   |- %s <- symbolic link that the creates.(Eval:%s)

By changing the environment variable GOROOT to [%s], you can easily switch GOROOT.

Is it OK?[Y/n] 
`, root, now, ver, op.LinkName, ver, filepath.Join(root, op.LinkName))

		//入力受付
		stdin := bufio.NewScanner(op.StdIn)
		stdin.Scan()
		text := stdin.Text()
		if text != "Y" {
			return "", fmt.Errorf("Cancel.")
		}
	}

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
		return "", xerrors.Errorf("create download command: %w", err)
	}
	//delete exe file
	defer os.Remove(bin)

	err = runDownloadCmd(bin)
	if err != nil {
		return "", xerrors.Errorf("run download command: %w", err)
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

	op := getOption()
	link := filepath.Join(dir, op.LinkName)
	//symbliclink
	if _, err := os.Lstat(link); err == nil {
		err = os.Remove(link)
		if err != nil {
			return "", err
		}
	} else {
		//???
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
		if v == "tip" {
			err := os.RemoveAll(path)
			if err != nil {
				//開発中実行がこのパスだった場合goを削除できないので無視
				fmt.Fprintln(getOption().StdErr, err)
			}
		} else {
			return path, nil
		}
	}

	//go download
	sdk, err := Download(v)
	if err != nil {
		return "", xerrors.Errorf("download error: %w", err)
	}

	//Download SDK Rename
	err = os.Rename(sdk+string(filepath.Separator), path+string(filepath.Separator))
	if err != nil {
		return "", xerrors.Errorf("rename error: %w", err)
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

func runDownloadCmd(bin string) error {

	fmt.Println("download cmd start")
	// $GOPATH/bin/go{version}{.exe} download
	cmd := exec.Command(bin, "download")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	err = cmd.Start()
	if err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Fprintln(os.Stdout, line)
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	fmt.Println("download cmd end")
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

//
// existsGo is go command exists
//
// goコマンドが存在するかを見ます
//
func existGo() bool {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return false
	}
	if out != nil && len(out) != 0 {
		return true
	}
	return false
}

func printSetting(root, version string) {
	fmt.Printf(`
%s にGoの最新バージョン(%s)をインストールしました。
環境変数GOROOTに%sを設定し、PATHをGOROOT/binに設定してください。

今後は

  $ golin 1.16

などでバージョンの切り替えが可能になります。
    `, root, version, root)
}

// リリース用のZIPを作成
func CompressReleaseZip(dst string, cmd string) error {

	w, err := os.Create(dst)
	if err != nil {
		return xerrors.Errorf("os.Create(): %w", err)
	}
	defer w.Close()

	err = Compress(w, cmd, "README.md")
	if err != nil {
		return xerrors.Errorf("CompressZip(): %w", err)
	}

	return nil
}

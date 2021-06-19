package golin

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shizuokago/golin/config"
	"golang.org/x/xerrors"
)

//定数
const (
	workDirectory = "golin_work" //権限確認用のディレクトリ名
)

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

	conf := config.Get()
	link := conf.LinkName

	root := filepath.Dir(goroot)
	now := filepath.Base(goroot)
	idx := strings.Index(goroot, link)

	//最後がリンク名と同一かを見る
	if idx != len(goroot)-len(link) {
		fmt.Fprintf(os.Stdout, `
This command creates the Go SDK within the current GOROOT parent directory. 
It is recommended to specify a dedicated directory.

%s -
   |- %s
   |- %s [Download specified Go SDK]
   |- %s <- symbolic link that the creates.(Eval:%s)

By changing the environment variable GOROOT to [%s], you can easily switch GOROOT.

Is it OK?[Y/n] 
`, root, now, ver, link, ver, filepath.Join(root, link))

		//入力受付
		stdin := bufio.NewScanner(os.Stdin)
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
// golang.org/dl/goX.x.x をinstallして取得してきます
// そのままGOPATHの位置でinstallされ、コマンドが作成されますので
// そのコマンド名も返します
//
func createDownloadCmd(v string) (string, error) {

	link := fmt.Sprintf("%s/go%s", config.GoGetLink, v)

	// go get golang.org/dl/go{version}
	cmd := exec.Command("go", "install", link)
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

	conf := config.Get()

	link := filepath.Join(dir, conf.LinkName)
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
		if v == CompileSDK {
			err := os.RemoveAll(path)
			if err != nil {
				//開発中実行がこのパスだった場合goを削除できないので無視
				fmt.Fprintln(os.Stderr, err)
			}
			//再度ダウンロード用に設定する
			v = "tip"
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

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return xerrors.Errorf("runCmd() error: %w", err)
	}
	return nil
}

func runDownloadCmd(bin string) error {

	//fmt.Println("download cmd start")

	// $GOPATH/bin/go{version}{.exe} download
	cmd := exec.Command(bin, "download")
	w := &downloadWriter{}
	cmd.Stdout = w
	cmd.Stderr = w
	cmd.Stdin = os.Stdin

	err := cmd.Start()
	if err != nil {
		return xerrors.Errorf("command start error: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return xerrors.Errorf("command wait error: %w", err)
	}

	//fmt.Println("download cmd end")
	return nil
}

type downloadWriter struct {
}

func (w *downloadWriter) Write(b []byte) (int, error) {
	line := string(b)

	if strings.Index(line, "Downloaded") != -1 {
		if strings.Index(line, "\n") != -1 {
			line = line[0 : len(line)-1]
		}
		fmt.Fprint(os.Stdout, "\r"+line)
	} else if strings.Index(line, "Unpacking") != -1 {
		fmt.Fprint(os.Stdout, "\n"+line)
	} else {
		fmt.Fprint(os.Stderr, line)
	}
	return len(b), nil
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

	fmt.Fprintln(os.Stdout, prefix, ver)
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

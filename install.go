package golin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/shizuokago/golin/v2/config"
	"golang.org/x/xerrors"
)

func Install(path string, ver string) error {

	//権限の確認
	err := checkAuthorization(path)
	if err != nil {
		return xerrors.Errorf("Authorization error: %w", err)
	}

	var v *Version
	// 指定がない場合、バージョンを取得
	if ver == "" {
		v, err = getLatestVersion()
		if err != nil {
			return xerrors.Errorf("getLatestVersion() error: %w", err)
		}
	} else {
		v = NewVersion(ver)
	}

	// そのバージョンをダウンロードし展開
	url := fmt.Sprintf("%s/go%s.%s-%s.%s", config.GolangDownloadPage, v.String(), runtime.GOOS, runtime.GOARCH, getDownloadExt())

	fmt.Println("Download Latest Version:", url)

	dp := filepath.Join(path, v.String())
	//作成
	err = DecompressURL(url, dp)
	if err != nil {
		return xerrors.Errorf("DecompressURL() error: %w", err)
	}

	//currentを作成
	link, restore, err := readyLink(path)
	if err != nil {
		return xerrors.Errorf("readyLink() error: %w", err)
	}

	//シンボリックリンクを作成
	err = os.Symlink(dp, link)
	if err != nil {
		//失敗するとGOROOTが存在しない状態になる為、元のリンクに戻す
		if rerr := restore(); rerr != nil {
			fmt.Fprintf(os.Stderr, "リンクの復元に失敗しました。: %v\n", rerr)
		}
		return xerrors.Errorf("symlink: %w", err)
	}

	// 各OSに合わせた設定手順を表示
	printSetting(link, v.String())

	return nil
}

package golin

import (
	"os"

	"golang.org/x/xerrors"
)

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

	//ルートを取得
	root, err := getRoot(v)
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

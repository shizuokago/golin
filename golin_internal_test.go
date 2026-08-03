package golin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shizuokago/golin/v2/config"
)

// copyDir()がコピー元のパーミッションを引き継ぐかを確認する
//
// os.Create()/os.Mkdir(dst,0666)で作成していた際、
// Windows以外でbin/goの実行権限が落ちて切り替え後のGoが動かなくなった
//
// Windowsは読み取り専用ビットしか扱わない為、
// 固定値ではなくコピー元のモードと比較する
func TestCopyDirPerm(t *testing.T) {

	base := t.TempDir()
	src := filepath.Join(base, "src")
	dst := filepath.Join(base, "dst")

	err := os.MkdirAll(filepath.Join(src, "bin"), 0755)
	if err != nil {
		t.Fatalf("os.MkdirAll() error: %v", err)
	}

	files := map[string][]byte{
		filepath.Join("bin", "go"): []byte("binary"),
		"README":                   []byte("text"),
	}
	for name, data := range files {
		perm := os.FileMode(0644)
		if filepath.Dir(name) == "bin" {
			perm = 0755
		}
		err = os.WriteFile(filepath.Join(src, name), data, perm)
		if err != nil {
			t.Fatalf("os.WriteFile() error: %v", err)
		}
	}

	err = copyDir(src, dst)
	if err != nil {
		t.Fatalf("copyDir() error: %v", err)
	}

	for name, want := range files {
		data, err := os.ReadFile(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("os.ReadFile() error: %v", err)
		}
		if string(data) != string(want) {
			t.Errorf("%s content = %q, want %q", name, data, want)
		}
	}

	//コピー元とコピー先のモードが一致するか
	for _, name := range []string{"bin", filepath.Join("bin", "go"), "README"} {
		si, err := os.Stat(filepath.Join(src, name))
		if err != nil {
			t.Fatalf("os.Stat() error: %v", err)
		}
		di, err := os.Stat(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("os.Stat() error: %v", err)
		}
		if si.Mode().Perm() != di.Mode().Perm() {
			t.Errorf("%s mode = %v, want %v", name, di.Mode().Perm(), si.Mode().Perm())
		}
	}

	//既存の位置に対して再実行しても失敗しない事
	err = copyDir(src, dst)
	if err != nil {
		t.Errorf("copyDir() rerun error: %v", err)
	}
}

// golin tip がdevでビルドした結果を再利用するかを確認する
//
// 以前はdevが{root}/compile_sdkに配置していた為、
// tipが参照する{root}/tipが作られずビルドし直しになっていた
//
// 存在する場合はダウンロードを行わない事の確認となる
// (存在しない場合はネットワークアクセスが発生する為テストしない)
func TestReadyPathTipExists(t *testing.T) {

	dir := t.TempDir()

	tip := filepath.Join(dir, TipVersion)
	err := os.Mkdir(tip, 0755)
	if err != nil {
		t.Fatalf("os.Mkdir() error: %v", err)
	}

	path, err := readyPath(dir, TipVersion)
	if err != nil {
		t.Fatalf("readyPath() error: %v", err)
	}
	if path != tip {
		t.Errorf("readyPath() = %s, want %s", path, tip)
	}
}

// リンクの作成に失敗した場合に元のリンクへ戻るかを確認する
func TestReadyLinkRestore(t *testing.T) {

	dir := t.TempDir()

	org := filepath.Join(dir, "1.24.0")
	err := os.Mkdir(org, 0755)
	if err != nil {
		t.Fatalf("os.Mkdir() error: %v", err)
	}

	link := filepath.Join(dir, config.Get().LinkName)
	err = os.Symlink(org, link)
	if err != nil {
		//Windowsでは管理者権限が必要な為、作成できない場合はスキップする
		t.Skipf("skipping: symlink not available: %v", err)
	}

	got, restore, err := readyLink(dir)
	if err != nil {
		t.Fatalf("readyLink() error: %v", err)
	}
	if got != link {
		t.Errorf("readyLink() = %s, want %s", got, link)
	}

	//削除されている事
	if _, err := os.Lstat(link); err == nil {
		t.Fatalf("link is not removed")
	}

	//復元でリンク先が元に戻る事
	err = restore()
	if err != nil {
		t.Fatalf("restore() error: %v", err)
	}
	dst, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("os.Readlink() error: %v", err)
	}
	if dst != org {
		t.Errorf("restored link = %s, want %s", dst, org)
	}
}

// printGoVersion()が解析する go version の出力形式
func TestParseGoVersionOutput(t *testing.T) {

	data := []struct {
		out  string
		want string
	}{
		{"go version go1.17beta1 windows/amd64", "1.17beta1"},
		{"go version go1.24.0 linux/amd64", "1.24.0"},
		{"go version devel go1.25-0123456789 windows/amd64", "1.25"},
		{"go version devel +abc123 Mon Jan 1 00:00:00 2024 +0000 linux/amd64", ""},
		{"go version", ""},
	}

	for _, d := range data {
		v := parseGoVersion(d.out)
		if d.want == "" {
			if v != nil {
				t.Errorf("parseGoVersion(%q) = %v, want nil", d.out, v)
			}
			continue
		}
		if v == nil {
			t.Errorf("parseGoVersion(%q) = nil, want %s", d.out, d.want)
			continue
		}
		if v.Source() != d.want {
			t.Errorf("parseGoVersion(%q) = %s, want %s", d.out, v.Source(), d.want)
		}
	}
}

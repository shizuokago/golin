//go:build !windows
// +build !windows

package golin

import (
	"os"
)

// createLink is create directory link
//
// Windows以外はシンボリックリンクをそのまま作成する
func createLink(target, link string) error {
	return os.Symlink(target, link)
}

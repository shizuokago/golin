// +build !windows

package golin

import (
	"os"
)

//
// Function getHome is HOME directory
//
// 環境変数HOMEを返す
//
func getHome() string {
	home := os.Getenv("HOME")
	return home
}

func getDownloadExt() string {
	return "tar.gz"
}

//go:build windows
// +build windows

package golin

import (
	"os"
)

//
// getHome is USERPROFILE directory
//
// WindowsのユーザHOMEであるUSERPROFILEのパスを返す
//
func getHome() string {
	home := os.Getenv("USERPROFILE")
	return home
}

func getDownloadExt() string {
	return "zip"
}

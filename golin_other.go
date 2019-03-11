// +build !windows

package golin

import (
	"os"
)

//
// Function getHome() is HOME directory
//
func getHome() string {
	home := os.Getenv("HOME")
	return home
}

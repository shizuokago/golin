// +build windows

package main

import (
	"os"
	"os/exec"
)

//
// Function getHome() is HOME directory
//
func getHome() string {
	home := os.Getenv("USERPROFILE")
	return home
}

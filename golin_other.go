package main

// +build !windows

import (
	"os"
	"os/exec"
)

//
// Function getHome() is HOME directory
//
func getHome() string {
	home := os.Getenv("HOME")
	return home
}

func createRemoveCmd(link string) *exec.Cmd {
	cmd := exec.Command("rm", link)
	return cmd
}

func createMoveCmd(sdk, path string) *exec.Cmd {
	cmd := exec.Command("mv", sdk, path)
	return cmd
}

func createLinkCmd(path, link string) *exec.Cmd {
	cmd := exec.Command("ln", "-ds", path, link)
	return cmd
}

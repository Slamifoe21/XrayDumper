//go:build !windows

package main

func ownsConsole() bool {
	return false
}

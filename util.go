package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrScreenshotCmdNotFound = errors.New("Screenshot executable not found")
	ErrNoScreenshotUtilFound = errors.New("No suitable screenshot utility found")
)

func BinExists(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func TakeScreenshot(path, cmdName string) (string, error) {
	path += ".png"
	var cmd *exec.Cmd

	if cmdName != "" {
		args := strings.Split(cmdName, " ")
		args = append(args, path)

		if !BinExists(args[0]) {
			return "", ErrScreenshotCmdNotFound
		}
		cmd = exec.Command(args[0], args[1:]...)
	} else if BinExists("maim") && BinExists("slop") {
		cmd = exec.Command("maim", "-ns", "-t0", "-b2", "-c0.5,0.4,0.9,0.75", path)
	} else if BinExists("scrot") {
		cmd = exec.Command("scrot", "-s", path)
	} else if BinExists("boxcutter") {
		cmd = exec.Command("boxcutter", path)
	} else if BinExists("gm") {
		cmd = exec.Command("gm", "import", path)
	} else if BinExists("import") {
		cmd = exec.Command("import", path)
	} else if BinExists("screencapture") {
		cmd = exec.Command("screencapture", "-sxr", path)
	} else {
		return "", ErrNoScreenshotUtilFound
	}

	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New("Screenshot file not found\n" + string(out))
	}

	return path, nil
}

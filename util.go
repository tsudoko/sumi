package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
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
			return "", errors.New(strScreenshotExecNotFound)
		}
		cmd = exec.Command(args[0], args[1:]...)
	} else if BinExists("maim") && BinExists("slop") {
		cmd = exec.Command("maim", "-ns", "-t0", "-b2", "-c0.5,0.4,0.9,1", path)
	} else if BinExists("scrot") {
		cmd = exec.Command("scrot", "-s", path)
	} else if BinExists("boxcutter") {
		cmd = exec.Command("boxcutter", path)
	} else {
		return "", errors.New(strNoScreenshotUtilFound)
	}

	out, err := cmd.CombinedOutput()

	if err != nil {
		return "", errors.New(err.Error())
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", errors.New(strScreenshotFileNotFound + "\n" + string(out))
	}

	return path, nil
}

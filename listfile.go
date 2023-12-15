package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var gHebDataDir string
var gHebTxtPath string
var gHebFd *os.File
var gHebFileIndex int = 0

func listFile(argStartIndex int) int {
	{
		ret := getDataDir()
		if 0 != ret {
			return ret
		}
		getFilelistPath()
		if filepath.Separator != gHebDataDir[len(gHebDataDir)-1] {
			gHebDataDir += string(filepath.Separator)
		}

		//create list file
		var err error
		gHebFd, err = os.Create(gHebTxtPath)
		if nil != err {
			printf("failed to create listfile.txt err=%s", err)
			return -4
		}
	}

	err2 := filepath.WalkDir(gHebExeDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return saveFilePath(path)

	})
	if err2 != nil {
		printf("failed to WalkDir err=%s", err2)
		return -3
	}

	printf("***************************")
	printf("********** Done ***********")
	printf("***************************")
	return 0
}

func saveFilePath(path string) (err error) {
	if path == gHebExePath {
		return nil
	}

	if strings.HasPrefix(path, gHebDataDir) {
		return nil
	}

	_, err = gHebFd.WriteString(fmt.Sprintf("%s\n", path))
	if nil != err {
		return
	}
	gHebFileIndex++

	return nil
}

func getDataDir() int {
	gHebDataDir = filepath.Join(gHebExeDir, "hebEraseData")
	err := os.MkdirAll(gHebDataDir, os.ModePerm)
	if nil != err {
		printf("failed to MkdirAll err=%s", err)
		return -2
	}
	return 0
}

func getFilelistPath() {
	gHebTxtPath = filepath.Join(gHebDataDir, "listfile.txt")
}

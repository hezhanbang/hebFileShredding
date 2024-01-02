package main

import (
	"os"
	"path/filepath"
)

var gHebCfg hebCfg

type hebCfg struct {
	workDir           string
	dataDir           string
	fileAboutListFile string

	exePath string
	exeDir  string
	exeName string
}

func (this *hebCfg) init() int {
	//about exe file
	this.exePath, _ = os.Executable()
	this.exePath, _ = filepath.Abs(this.exePath)

	exeDir := filepath.Dir(this.exePath)
	if len(exeDir) <= 1 {
		printf("failed to get dir from exe path")
		return -1
	}
	this.exeDir = exeDir

	exeName := filepath.Base(this.exePath)
	if len(exeName) <= 1 {
		printf("failed to get exe file name")
		return -2
	}
	this.exeName = exeName

	this.workDir, _ = os.Getwd()

	//data dir
	dataDir := filepath.Join(this.workDir, ".hebFileShredding")
	err := os.MkdirAll(dataDir, os.ModePerm)
	if nil != err {
		printf("failed to MkdirAll for data dir, err=%s", err)
		return -3
	}
	//dataDir must end with '/' or '\'
	if filepath.Separator != dataDir[len(dataDir)-1] {
		dataDir += string(filepath.Separator)
	}
	this.dataDir = dataDir

	//a file which contains all file path from `listfile` command
	this.fileAboutListFile = filepath.Join(this.dataDir, "listfile.txt")

	return 0
}

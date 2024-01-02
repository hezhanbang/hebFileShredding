package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func main() {
	if ret := gHebCfg.init(); 0 != ret {
		os.Exit(ret)
		return
	}

	//获得顶级指令
	cmd, index := getTopCommand(gHebCfg.exeName)
	ret := 100

	//myTest()

	//仅仅用于调试
	if false {
		cmd = "erase"
	}

	//调用子命令模块
	if "listfile" == cmd {
		ret = gHebList.do(index + 1)
	} else if "erase" == cmd {
		ret = eraseFiles(index+1, false)
	} else if "deepErase" == cmd {
		ret = eraseFiles(index+1, true)
	} else {
		ret = cmdUsage()
	}

	if ret < 0 {
		ret = 0 - ret
	}
	os.Exit(ret)
}

func cmdUsage() int {
	/*
		fmt.Printf("please use the comand argument: listfile / erase / deepErase\n\n")
		fmt.Printf("erase: it will erase data at the begin of file and the end of file\n")
		fmt.Printf("deepErase: it will only erase some of data at the middle of file randomly\n")
	*/
	fmt.Printf("本软件用于粉碎文件夹里的全部文件内容。\n\n")
	fmt.Printf("请使用如下的任意一个参数: listfile / erase / deepErase\n\n")
	fmt.Printf(" listfile: 列出当前文件夹下的全部文件，并记录到[hebEraseData/listfile.txt]\n")
	fmt.Printf("    erase: 根据[listfile.txt]，擦写文件开头1MB内容和文件结尾1MB内容\n")
	fmt.Printf("deepErase: 根据[listfile.txt]，擦写文件中间段的内容，以随机的方式修改少量数据\n")

	return 101
}

func getTopCommand(exeNotExt string) (cmd string, location int) {
	/*
		./xxxx.out listfile
		bash ./xxxx.out listfile
	*/

	if strings.HasSuffix(strings.ToLower(exeNotExt), ".exe") || strings.HasSuffix(strings.ToLower(exeNotExt), ".out") {
		exeNotExt = exeNotExt[0 : len(exeNotExt)-4]
	}

	if len(os.Args) < 2 {
		return "", -1
	}
	if !strings.Contains(os.Args[1], exeNotExt) {
		return os.Args[1], 1
	}

	if len(os.Args) < 3 {
		return "", -2
	}
	return os.Args[2], 2
}

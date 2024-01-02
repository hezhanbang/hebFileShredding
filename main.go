package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if ret := gHebCfg.init(); 0 != ret {
		os.Exit(ret)
		return
	}

	//获得顶级指令
	cmd, otherArgs := getTopCommand(gHebCfg.exeName)
	ret := 100

	//myTest()

	//仅仅用于调试
	if false {
		cmd = "erase"
	}

	//调用子命令模块
	if "listfile" == cmd {
		ret = gHebList.do(cmd, otherArgs)
	} else if "erase" == cmd {
		ret = gHebErase.do(cmd, otherArgs)
	} else {
		ret = cmdUsage()
	}

	if ret < 0 {
		ret = 0 - ret
	}
	os.Exit(ret)
}

// 打印参数用法
func cmdUsage() int {
	/*
		fmt.Printf("please use the comand argument: listfile / erase / deepErase\n\n")
		fmt.Printf("erase: it will erase data at the begin of file and the end of file\n")
		fmt.Printf("deepErase: it will only erase some of data at the middle of file randomly\n")
	*/
	fmt.Printf("本软件用于粉碎文件夹里的全部文件内容。\n\n")
	fmt.Printf("请使用如下的任意一个参数: listfile / erase\n")
	fmt.Printf(" listfile: 列出当前文件夹下的全部文件，并记录到[hebEraseData/listfile.txt]\n")
	fmt.Printf("    erase: 根据[listfile.txt]，擦写文件内容\n")

	return 101
}

// 从命令行里获取顶层命令
func getTopCommand(exeNotExt string) (retCmd string, retOtherArgs []string) {
	/*
		./xxxx.out listfile
		bash ./xxxx.out listfile
	*/

	retCmd = ""
	retOtherArgs = nil

	if strings.HasSuffix(strings.ToLower(exeNotExt), ".exe") || strings.HasSuffix(strings.ToLower(exeNotExt), ".out") {
		exeNotExt = exeNotExt[0 : len(exeNotExt)-4]
	}
	allArgLen := len(os.Args)

	//第一个参数一般是可执行文件，顶层命令位于第二个参数。
	if allArgLen < 2 {
		return
	}
	if !strings.Contains(os.Args[1], exeNotExt) {
		retCmd = os.Args[1]
		if allArgLen >= 3 {
			retOtherArgs = os.Args[2:]
		}
	}

	//第二个参数是可执行文件，顶层命令位于第三个参数。
	if len(os.Args) < 3 {
		return "", nil
	}
	retCmd = os.Args[2]
	if allArgLen >= 4 {
		retOtherArgs = os.Args[3:]
	}
	return
}

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var gHebTxtPath string
var gHebFd *os.File
var gHebFileIndex int = 0

func listFile(argStartIndex int) int {
	{
		//create list file
		var err error
		gHebFd, err = os.Create(gHebTxtPath)
		if nil != err {
			printf("failed to create listfile.txt err=%s", err)
			return -4
		}
	}

	err2 := filepath.WalkDir(gHebCfg.workDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return filePath2File(path)

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

func filePath2File(pathNeedErase string) (err error) {
	//数据文件夹和本可执行文件，不需要擦除。
	if pathNeedErase == gHebCfg.exePath {
		return nil
	}
	if strings.HasPrefix(pathNeedErase, gHebCfg.dataDir) {
		return nil
	}

	_, err = gHebFd.WriteString(fmt.Sprintf("%s\n", pathNeedErase))
	if nil != err {
		return
	}
	gHebFileIndex++

	return nil
}

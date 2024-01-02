package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var gHebList hebListContext

type hebListContext struct {
	cmd       string
	fd        *os.File
	fileIndex int
}

func (this *hebListContext) do(cmd string, args []string) int {
	//create list file
	{
		var err error
		this.fd, err = os.Create(gHebCfg.fileAboutListFile)
		if nil != err {
			printf("failed to create listfile.txt. %s err=%s", gHebCfg.fileAboutListFile, err)
			return -1
		}
	}

	err2 := filepath.WalkDir(gHebCfg.workDir, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		return this.filePath2File(path)

	})
	if err2 != nil {
		printf("failed to WalkDir err=%s", err2)
		return -3
	}

	printf("a total of %d files has been recorded to listfile.txt", this.fileIndex)

	printf("***************************")
	printf("********** Done ***********")
	printf("***************************")
	return 0
}

func (this *hebListContext) filePath2File(pathNeedErase string) (err error) {
	//数据文件夹和本可执行文件，不需要擦除。
	if pathNeedErase == gHebCfg.exePath {
		return nil
	}
	if strings.HasPrefix(pathNeedErase, gHebCfg.dataDir) {
		return nil
	}

	_, err = this.fd.WriteString(fmt.Sprintf("%s\n", pathNeedErase))
	if nil != err {
		return
	}
	this.fileIndex++

	return nil
}

package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"strings"
)

var gHebErase hebEraseContext

type hebEraseContext struct {
	dataSizeNeedToErase     int64
	zeroDataForFileStarting []byte
	zeroDataForFileEnding   []byte
	deepEraseMode           bool

	fd        *os.File
	fileIndex int
	total     int
}

/*
用全零填充文件开始1MB的数据，
用全零填充文件末尾1MB的数据，
文件中间的数据不处理。
*/
func (this *hebEraseContext) init(deepErase bool) int {
	this.deepEraseMode = deepErase

	this.dataSizeNeedToErase = (1024 * 1024)

	//生成用于填充文件开始1MB的零数据。
	this.zeroDataForFileStarting = make([]byte, this.dataSizeNeedToErase)
	for i := int64(0); i < this.dataSizeNeedToErase; i++ {
		this.zeroDataForFileStarting[i] = 0
	}
	//在最开始处，添加一个标识。
	mark := []byte("File Shredding start \n\n")
	for i := 0; i < len(mark); i++ {
		this.zeroDataForFileStarting[i] = mark[i]
	}

	//生成用于填充文件末尾1MB的零数据。
	this.zeroDataForFileEnding = make([]byte, this.dataSizeNeedToErase)
	for i := int64(0); i < this.dataSizeNeedToErase; i++ {
		this.zeroDataForFileEnding[i] = 0
	}
	//在最末尾处，添加一个标识。
	mark = []byte("\n\nFile Shredding end\n")
	offset := int(this.dataSizeNeedToErase) - len(mark)
	for i := 0; i < len(mark); i++ {
		this.zeroDataForFileEnding[offset+i] = mark[i]
	}

	//获取需要被擦除的文件总数。
	{
		this.openListFile()

		//逐行读取文本文件
		fileScanner := bufio.NewScanner(this.fd)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			fileNeedToErase := fileScanner.Text()
			_, ok := this.needErase(fileNeedToErase)
			if ok {
				this.total++
			}
		}

		this.closeListFile()

		if nil != fileScanner.Err() {
			printf("failed to Scan txt file, err=%s", fileScanner.Err())
			return -3
		}
	}

	return 0
}

func (this *hebEraseContext) do(argStartIndex int, deepErase bool) int {
	if ret := this.init(deepErase); 0 != ret {
		return -1
	}

	if ret := this.openListFile(); 0 != ret {
		return -2
	}
	defer this.fd.Close()

	//逐行读取文本文件
	fileScanner := bufio.NewScanner(this.fd)
	fileScanner.Split(bufio.ScanLines)
	ok := false

	for fileScanner.Scan() {
		fileNeedToErase := fileScanner.Text()
		printf("erasing (%d/%d): %s", this.fileIndex+1, this.total, fileNeedToErase)

		fileNeedToErase, ok = this.needErase(fileNeedToErase)
		if false == ok {
			printf("skipped the file No.=%d %s", this.fileIndex+1, fileNeedToErase)
			continue
		}

		ret := this.eraseOneFile(fileNeedToErase)
		if 0 != ret {
			printf("failed to erase one file, No.=%d, %s, ret=%d", this.fileIndex+1, fileNeedToErase, ret)
			return -3
		}
		this.fileIndex++
	}

	if nil != fileScanner.Err() {
		printf("failed to Scan txt file, err=%s", fileScanner.Err())
		return -4
	}

	printf("***************************")
	printf("********** Done ***********")
	printf("***************************")
	return 0
}

func (this *hebEraseContext) eraseOneFile(path string) int {
	//打开需要被擦除的文件
	fd, err := os.OpenFile(path, os.O_RDWR, 0644)
	if nil != err {
		if errors.Is(err, os.ErrNotExist) {
			printf("skipped the file, because it does not exist, No.=%d %s", this.fileIndex+1, path)
			return 0
		}

		//没有权限，尝试更改文件权限，然后再次打开文件。
		if errors.Is(err, os.ErrPermission) {
			err = os.Chmod(path, 0644)
			if nil != err {
				printf("failed to Chmod file to erase, err=%s", err)
				return -1
			}

			fd, err = os.OpenFile(path, os.O_RDWR, 0644)
			if nil != err {
				printf("failed to open file again to erase, err=%s", err)
				return -2
			}
		} else {
			printf("failed to open file to erase, err=%s", err)
			return -3
		}
	}

	defer func() {
		if nil != fd {
			fd.Close()
			fd = nil
		}
	}()

	//检查文件信息
	sta, err2 := fd.Stat()
	if nil != err2 {
		printf("failed to get file info to erase, err=%s", err)
		return -4
	}

	if sta.IsDir() {
		printf("skipped the file, because it is dir, No.=%d %s", this.fileIndex+1, path)
		return 0
	}

	filesize := sta.Size()
	if filesize <= 0 {
		printf("skipped the file, because its file size is 0, No.=%d %s", this.fileIndex+1, path)
		return 0
	}

	//开始擦除
	if filesize <= this.dataSizeNeedToErase { //文件很小，擦除文件的全部内容
		ret := this.erasePart(path, fd, 0, filesize, this.zeroDataForFileStarting)
		if 0 != ret {
			return -15
		}
	} else if filesize >= this.dataSizeNeedToErase*2 { //文件很大
		if false == this.deepEraseMode {
			//填充文件开始1MB的数据
			ret := this.erasePart(path, fd, 0, this.dataSizeNeedToErase, this.zeroDataForFileStarting)
			if 0 != ret {
				return -16
			}
			//填充文件末尾1MB的数据
			ret = this.erasePart(path, fd, filesize-this.dataSizeNeedToErase, this.dataSizeNeedToErase, this.zeroDataForFileEnding)
			if 0 != ret {
				return -17
			}
		} else {
			//深度擦写，随机擦写文件中间的个别内容

			ranMax := new(big.Int)
			ranMax.SetUint64(1000)

			zero := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

			step := this.dataSizeNeedToErase + 123

			total := filesize - this.dataSizeNeedToErase*2
			total = total - int64(len(zero)) - 2

			for cur := step; cur < total; cur += step {
				fileOffset := this.dataSizeNeedToErase + cur
				count, err := fd.WriteAt(zero, fileOffset)
				if err != nil || count != len(zero) {
					printf("failed to write file deeply [%s] err=%s", path, err)
					return -1
				}
				number, err := rand.Int(rand.Reader, ranMax)
				if nil == err && nil != number {
					ranKey := number.Int64()
					//fmt.Printf("new rank key=%d\n", ranKey)
					cur += ranKey
				}
			}
		}
	}

	return 0
}

func (this *hebEraseContext) erasePart(path string, fd *os.File, startIndex, len int64, zeroArray []byte) int {
	array := zeroArray
	if len < this.dataSizeNeedToErase {
		array = zeroArray[:len]
	}
	count, err := fd.WriteAt(array, startIndex)
	if err != nil || int64(count) != len {
		printf("failed to write file [%s] err=%s", path, err)
		return -1
	}

	return 0
}

func (this *hebEraseContext) openListFile() int {
	this.closeListFile()

	fd, err := os.Open(gHebCfg.fileAboutListFile)
	if nil != err {
		printf("failed to open listfile.txt, err=%s", err)
		return -1
	}
	this.fd = fd
	return 0
}

func (this *hebEraseContext) closeListFile() {
	if nil != this.fd {
		this.fd.Close()
		this.fd = nil
	}
}

func (this *hebEraseContext) needErase(pathLine string) (retPath string, retNeedErase bool) {
	pathLine = strings.TrimSpace(pathLine)
	retPath = strings.Trim(pathLine, "\t")

	if strings.HasPrefix(pathLine, "#") || strings.HasPrefix(pathLine, "//") {
		retNeedErase = false
		return
	}
	retNeedErase = true
	return
}

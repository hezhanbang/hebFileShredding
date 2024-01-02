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
文件中间的数据随机处理。
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

	//打开文本文件
	if ret := this.openListFile(); 0 != ret {
		return -2
	}
	defer this.closeListFile()

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
	fdToErase, gotErr := this.openFileNeedToErase(path)
	if gotErr {
		return -1
	}

	//文件不存在。
	if nil == fdToErase {
		return 0
	}

	assert(nil != fdToErase)
	defer fdToErase.Close()

	//检查文件信息
	sta, err := fdToErase.Stat()
	if nil != err {
		printf("failed to get file info to erase, err=%s", err)
		return -2
	}

	if sta.IsDir() {
		printf("skipped the file, because it is dir, No.=%d %s", this.fileIndex+1, path)
		return 0
	}

	filesize := sta.Size()
	if filesize <= 0 {
		printf("skipped the file, because the file size is 0, No.=%d %s", this.fileIndex+1, path)
		return 0
	}

	//开始擦除
	ret := this.eraseNow(path, fdToErase, filesize)
	if 0 != ret {
		return -3
	}

	return 0
}

// 打开需要被擦除的文件
func (this *hebEraseContext) openFileNeedToErase(pathNeedErease string) (retDd *os.File, retErr bool) {
	fd, err := os.OpenFile(pathNeedErease, os.O_RDWR, 0644)
	if nil == err {
		return fd, true
	}

	if errors.Is(err, os.ErrNotExist) {
		printf("skipped the file, because it does not exist, No.=%d %s", this.fileIndex+1, pathNeedErease)
		return nil, true
	}

	//没有权限，尝试更改文件权限，然后再次打开文件。
	if errors.Is(err, os.ErrPermission) {
		err = os.Chmod(pathNeedErease, 0644)
		if nil != err {
			printf("failed to Chmod file to erase, err=%s", err)
			return nil, false
		}

		fd, err = os.OpenFile(pathNeedErease, os.O_RDWR, 0644)
		if nil != err {
			printf("failed to open file again to erase, err=%s", err)
			return nil, false
		}
	} else {
		printf("failed to open file to erase, err=%s", err)
		return nil, false
	}

	return fd, true
}

// 开始擦除单个文件
func (this *hebEraseContext) eraseNow(pathNeedErease string, fd *os.File, filesize int64) int {
	if filesize <= this.dataSizeNeedToErase { //文件很小，擦除文件的全部内容
		ret := this.writeZeroToFile(pathNeedErease, fd, 0, filesize, this.zeroDataForFileStarting)
		if 0 != ret {
			return -1
		}
	} else if filesize >= this.dataSizeNeedToErase*2 { //文件很大
		ret := this.eraseBigFile(pathNeedErease, fd, filesize)
		if 0 != ret {
			return -2
		}
	}
	return 0
}

// 擦除大文件
func (this *hebEraseContext) eraseBigFile(pathNeedErease string, fd *os.File, filesize int64) int {
	if false == this.deepEraseMode {
		//填充文件开始1MB的数据
		ret := this.writeZeroToFile(pathNeedErease, fd, 0, this.dataSizeNeedToErase, this.zeroDataForFileStarting)
		if 0 != ret {
			return -1
		}
		//填充文件末尾1MB的数据
		ret = this.writeZeroToFile(pathNeedErease, fd, filesize-this.dataSizeNeedToErase, this.dataSizeNeedToErase, this.zeroDataForFileEnding)
		if 0 != ret {
			return -2
		}
	} else {
		//深度擦写，随机擦写文件中间的个别内容
		ret := this.randomErase(pathNeedErease, fd, filesize)
		if 0 != ret {
			return -3
		}
	}

	return 0
}

// 深度擦写，随机擦写文件中间的个别内容
func (this *hebEraseContext) randomErase(pathNeedErease string, fd *os.File, filesize int64) int {
	ranMax := new(big.Int)
	ranMax.SetUint64(1000)

	zero := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	step := this.dataSizeNeedToErase + 123

	//跳过文件开头和末尾
	total := filesize - this.dataSizeNeedToErase*2
	total = total - int64(len(zero)) - 2

	for cur := step; cur < total; cur += step {
		fileOffset := this.dataSizeNeedToErase + cur //跳过文件开头

		//擦除文件局部
		ret := this.writeZeroToFile(pathNeedErease, fd, fileOffset, int64(len(zero)), zero)
		if 0 != ret {
			return -1
		}

		//生成随机数，随机调到文件的下一个位置
		number, err := rand.Int(rand.Reader, ranMax)
		assert(nil == err && nil != number)
		ranKey := number.Int64()
		//fmt.Printf("new rank key=%d\n", ranKey)
		cur += ranKey
	}

	return 0
}

// 把全零数据写入文件，以实现文件数据擦除的目的。
func (this *hebEraseContext) writeZeroToFile(path string, fd *os.File, startIndex, len int64, zeroArray []byte) int {
	array := zeroArray
	if len < this.dataSizeNeedToErase {
		array = zeroArray[:len]
	}
	count, err := fd.WriteAt(array, startIndex)
	if err != nil || int64(count) != len {
		printf("failed to write zero to file [%s] err=%s", path, err)
		return -1
	}

	return 0
}

// 打开`文件列表`文件。
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

// 关闭`文件列表`文件。
func (this *hebEraseContext) closeListFile() {
	if nil != this.fd {
		this.fd.Close()
		this.fd = nil
	}
}

// 检查`文件列表`文件里的一行数据，如果行首有`//`或者`#`，表示不需要执行文件擦除。
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

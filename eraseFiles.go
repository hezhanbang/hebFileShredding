package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"math/big"
	"os"
)

var gHebMB int64 = (1024 * 1024)
var gHebZeroArray_start []byte
var gHebZeroArray_end []byte
var gHebDeepErase = false

/*
用全零填充文件开始1MB的数据，
用全零填充文件末尾1MB的数据，
文件中间的数据不处理。
*/
func eraseFiles(argStartIndex int, deepErase bool) int {
	gHebDeepErase = deepErase

	//zero 1
	gHebZeroArray_start = make([]byte, gHebMB)
	for i := int64(0); i < gHebMB; i++ {
		gHebZeroArray_start[i] = 0
	}
	mark := []byte("File Shredding start \n\n")
	for i := 0; i < len(mark); i++ {
		gHebZeroArray_start[i] = mark[i]
	}

	//zero 2
	gHebZeroArray_end = make([]byte, gHebMB)
	for i := int64(0); i < gHebMB; i++ {
		gHebZeroArray_end[i] = 0
	}
	mark = []byte("\n\nFile Shredding end\n")
	offset := int(gHebMB) - len(mark)
	for i := 0; i < len(mark); i++ {
		gHebZeroArray_end[offset+i] = mark[i]
	}

	readFile, err := os.Open(gHebTxtPath)
	if nil != err {
		printf("failed to open txt file, err=%s", err)
		return -1
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		//fmt.Println(fileScanner.Text())

		path := fileScanner.Text()
		ret := eraseOneFile(path)
		if 0 != ret {
			printf("failed to eraseOneFile %s, ret=%d", path, ret)
			return -2
		}
	}

	if nil != fileScanner.Err() {
		printf("failed to Scan txt file, err=%s", fileScanner.Err())
		return -3
	}

	printf("***************************")
	printf("********** Done ***********")
	printf("***************************")
	return 0
}

func eraseOneFile(path string) int {
	fd, err := os.OpenFile(path, os.O_RDWR, 0644)
	if nil != err {
		if errors.Is(err, os.ErrNotExist) {
			return 0
		}

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
		}
		return -1
	}

	defer func() {
		if nil != fd {
			fd.Close()
			fd = nil
		}
	}()

	sta, err2 := fd.Stat()
	if nil != err2 {
		printf("failed to get file info to erase, err=%s", err)
		return -3
	}

	if sta.IsDir() {
		return 0
	}

	filesize := sta.Size()
	if filesize <= 0 {
		return 0
	}

	if filesize <= gHebMB {
		ret := erasePart(path, fd, 0, filesize, gHebZeroArray_start)
		if 0 != ret {
			return -15
		}
	} else if filesize >= gHebMB*2 {
		if false == gHebDeepErase {
			//填充文件开始1MB的数据
			ret := erasePart(path, fd, 0, gHebMB, gHebZeroArray_start)
			if 0 != ret {
				return -16
			}
			//填充文件末尾1MB的数据
			ret = erasePart(path, fd, filesize-gHebMB, gHebMB, gHebZeroArray_end)
			if 0 != ret {
				return -17
			}
		} else {
			//深度擦写，随机擦写文件中间的个别内容

			ranMax := new(big.Int)
			ranMax.SetUint64(1000)

			zero := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

			step := gHebMB + 123

			total := filesize - gHebMB*2
			total = total - int64(len(zero)) - 2

			for cur := step; cur < total; cur += step {
				fileOffset := gHebMB + cur
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

func erasePart(path string, fd *os.File, startIndex, len int64, zeroArray []byte) int {
	array := zeroArray
	if len < gHebMB {
		array = zeroArray[:len]
	}
	count, err := fd.WriteAt(array, startIndex)
	if err != nil || int64(count) != len {
		printf("failed to write file [%s] err=%s", path, err)
		return -1
	}

	return 0
}

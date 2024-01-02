package main

import (
	"errors"
	"os"
)

func myTest() {
	defer os.Exit(0)

	file2 := "D:\\hebStreamMedia\\test\\clean\\.git\\objects\\0e\\960ae110c992d92e84528b474c352cda982019"
	_, err := os.OpenFile(file2, os.O_RDWR, 0644)
	if nil != err {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		if errors.Is(err, os.ErrPermission) {
			err = os.Chmod(file2, 0644)
			return
		}
		printf("failed to open file to erase, err=%s", err)
		return
	}
}

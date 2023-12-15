package main

import (
	"fmt"
	"time"
)

func printf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	all := "[" + time.Now().String() + "] " + msg
	if all[len(all)-1] != '\n' {
		all += "\n"
	}
	fmt.Print(all)
}

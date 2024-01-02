package main

import (
	"fmt"
	"strings"
	"time"
)

func printf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	timeStr := time.Now().String()
	if endIndex := strings.Index(timeStr, " m="); endIndex > 0 {
		timeStr = timeStr[0:endIndex]
	}
	all := "[" + time.Now().String() + "] " + msg
	if all[len(all)-1] != '\n' {
		all += "\n"
	}
	fmt.Print(all)
}

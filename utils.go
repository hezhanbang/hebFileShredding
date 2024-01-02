package main

import (
	"fmt"
	"strings"
	"time"
)

func printf(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)

	//格式化时间字符串，让时间字符串的长度都一样
	timeStr := time.Now().String()
	if endIndex := strings.Index(timeStr, " m="); endIndex > 0 {
		timeStr = timeStr[0:endIndex]
	}
	if endIndex := strings.Index(timeStr, " +"); endIndex > 0 && endIndex < 27 {
		part1 := timeStr[0:endIndex]

		for endIndex < 27 {
			part1 += "0"
			endIndex++
		}

		timeStr = part1 + timeStr[endIndex:]
	}

	all := "[" + timeStr + "] " + msg
	if all[len(all)-1] != '\n' {
		all += "\n"
	}
	fmt.Print(all)
}

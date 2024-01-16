package main

import (
	"flag"
	"fmt"
	"os"
)

//run: ./test.exe --mode=hello

func main() {
	CommandLine := flag.NewFlagSet("abc", flag.ExitOnError)

	mode := CommandLine.String("mode", "default", "description")

	CommandLine.VisitAll(func(a *flag.Flag) {
		fmt.Println(">>>>> ", a.Name, "value=", a.Value)
	})

	//命令行参数，需要剔除第一个命令名参数
	if err := CommandLine.Parse(os.Args[1:]); nil != err {
		fmt.Printf("err\n")
		return
	}

	if nil != mode {
		fmt.Printf("mode=%s\n", *mode)
	}

	fmt.Printf("\ndone\n")
}

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

	if err := CommandLine.Parse(os.Args); nil != err {
		fmt.Printf("err\n")
		return
	}

	if nil != mode {
		fmt.Printf("mode=%s\n", *mode)
	}

	fmt.Printf("\ndone\n")
}

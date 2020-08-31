// logtool manipulates local log files
package main

import (
	"flag"
	"fmt"
	"os"
	"io"

	"substation/pkg/appender"
)

func init() {
	if len(os.Args) < 2 {
		fmt.Println("Expected subcommand 'read', 'write'")
		os.Exit(1)
	}
	flagset := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	switch os.Args[1] {
	case "write":
	}
	flagset.Parse(os.Args[2:])
}

func doWrite() error {
	if a, err := appender.Create("./data"); err != nil {
		return err
	} else {
		defer a.Close()
		if _, err := io.Copy(a, os.Stdin); err != nil {
			return err
		}
		return nil
	}
}

func main() {
	var err error
	switch os.Args[1] {
	case "write":
		err = doWrite()
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

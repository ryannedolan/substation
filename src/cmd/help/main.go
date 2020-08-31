package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello! This image is designed to be run within K8s. Check out the Helm chart.")
	os.Exit(1)
}

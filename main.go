package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	username := os.Args[1]
	port := os.Args[2]
	numberOfMember := os.Args[3]
	n, err := strconv.Atoi(numberOfMember)
	if err != nil {
		fmt.Println("please use Integer to indicate the number of members.")
		return
	}
	StartServer(port, username, n)

}

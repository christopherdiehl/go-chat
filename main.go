package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage:\n connect -- connect to chat server\n start -- spool up chat server")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "connect":
		fmt.Println("Connecting you now")
	case "start":
		fmt.Println("Spooling up now")
	default:
		fmt.Println("Please specify connect or start")
		os.Exit(1)
	}

}

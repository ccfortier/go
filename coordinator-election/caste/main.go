package main

import (
	"fmt"
	"os"
)

func main() {
	var input string
	for {
		fmt.Print("Enter text: ")
		fmt.Scanln(&input)

		switch input {
		case "bye":
			fmt.Println("bye...")
			os.Exit(0)
		default:
			fmt.Println(input)
		}
	}

}

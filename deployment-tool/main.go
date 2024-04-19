package main

import (
	"os"
    "fmt"
    "diploy/cmd"
)

func main() {
    err := cmd.Execute()
    if err != nil {
        fmt.Println("An error occurred:", err)
        os.Exit(1) 
    }

}

package main

import (
	"fmt"
	"os"

	"github.com/AdityaKrSingh26/Glime/internal/terminal"
)

func main() {
	key, _ := terminal.ReadKey(os.Stdin)
	fmt.Printf("Key: %+v\n", key)
}

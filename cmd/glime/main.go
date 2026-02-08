package main

import (
	"fmt"
	"os"

	"github.com/AdityaKrSingh26/Glime/internal/input"
)

func main() {
	key, _ := input.ReadKey(os.Stdin)
	fmt.Printf("Key: %+v\n", key)
}

package main

import (
	"fmt"
	"os"

	"github.com/divinerapier/divineflake"
)

func main() {
	// fmt.Println(divineflake.LocalAddrWithPrefix(192, 168))

	fd, err := os.Create("./output.txt")
	if err != nil {
		panic(err)
	}

	for i := 0; i < 500000; i++ {
		fmt.Fprintf(fd, "%d\n", divineflake.Generate())
	}
}

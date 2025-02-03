/*
Package main implements cli execution.
*/
package main

import (
	"fmt"
	"os"

	"github.com/vigo/tablo/internal/tablo"
)

func main() {
	if err := tablo.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

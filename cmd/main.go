package main

import (
	"fmt"
	"os"

	"github.com/euank/sprigcli/cmd/sprig"
)

func main() {
	if err := sprig.NewSprigCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

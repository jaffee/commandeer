package main

import (
	"fmt"

	"github.com/jaffee/commandeer"
	"github.com/jaffee/commandeer/examples/myapp"
)

func main() {
	err := commandeer.Run(myapp.NewMain())
	if err != nil {
		fmt.Println(err)
	}
}

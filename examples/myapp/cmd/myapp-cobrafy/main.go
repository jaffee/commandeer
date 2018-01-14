package main

import (
	"fmt"

	"github.com/jaffee/commandeer/cobrafy"
	"github.com/jaffee/commandeer/examples/myapp"
)

func main() {
	err := cobrafy.Execute(myapp.NewMain())
	if err != nil {
		fmt.Println(err)
	}
}

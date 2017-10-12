package main

import (
	"fmt"
	"log"

	"github.com/jaffee/commandeer"
	"github.com/spf13/cobra"
)

type Main struct {
	Thing string `flag:"thing" help: "does a thing"`
}

func (m *Main) Run() error {
	fmt.Println(m.Thing)
	return nil
}

func main() {
	m := &Main{Thing: "man, what a thing"}
	rootCmd, err := commandeer.Cobra(m)
	if err != nil {
		log.Fatal(err)
	}
	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		return m.Run()
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

package myapp

import "fmt"

type Main struct {
	Num     int    `help:"How many does it take?"`
	Vehicle string `help:"What did they get?"`
}

func NewMain() *Main { return &Main{Num: 5, Vehicle: "jeep"} }

func (m *Main) Run() error {
	if m.Num < 2 || m.Vehicle == "" {
		return fmt.Errorf("Need more gophers and/or vehicles.")
	}
	fmt.Printf("%d gophers stole my %s!\n", m.Num, m.Vehicle)
	return nil
}

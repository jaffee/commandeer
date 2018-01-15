package myapp

import "fmt"

// Main has a couple basic fields with tags which will be used to set up flags
// with the given help strings.
type Main struct {
	Num     int    `help:"How many does it take?"`
	Vehicle string `help:"What did they get?"`
}

// NewMain makes a Main and sets up the default values (which will be used by
// commandeer as flag defaults).
func NewMain() *Main { return &Main{Num: 5, Vehicle: "jeep"} }

// Run implements the Runner interface so that commandeer.Run can be used with
// Main.
func (m *Main) Run() error {
	if m.Num < 2 || m.Vehicle == "" {
		return fmt.Errorf("Need more gophers and/or vehicles.")
	}
	fmt.Printf("%d gophers stole my %s!\n", m.Num, m.Vehicle)
	return nil
}

package myapp

import (
	"fmt"
	"time"
)

//                Go From This                              //                 To This
/***********************************************************/ /********************************************/
//  $ myapp -h
//  Usage of myapp:
type MyApp struct { //    -duration duration
	Num      int           `help:"How many does it take?"`  //          How long is it gone?
	Vehicle  string        `help:"What did they get?"`      //    -num int
	Running  bool          `help:"Is the vehicle working?"` //          How many does it take? (default 5)
	Duration time.Duration `help:"How long is it gone?"`    //    -running
} //          Is the vehicle working?
//    -vehicle string
//          What did they get? (default "jeep")
/***********************************************************/ /********************************************/
//                                         without manually defining flags.

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

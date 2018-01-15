package test

import (
	"fmt"
	"time"
)

// MyMain defines a variety of different field types and exercises various
// different tags.
type MyMain struct {
	Thing string `flag:"thing" help:"does a thing"`
	wing  string `flag:"wing" help:"does a wing"`
	Bing  string `flag:"-" help:"shouldn't happen"`

	Abool     bool          `flag:"a-bool" help:"boolean flag" short:"b"`
	Aint      int           `flag:"a-int" help:"int flag"`
	Aint64    int64         `flag:"a-int64" help:"int64 flag"`
	Afloat    float64       `flag:"a-float" help:"float flag"`
	Auint     uint          `flag:"a-uint" help:"uint flag"`
	Auint64   uint64        `flag:"a-uint64" help:"uint64 flag"`
	Aduration time.Duration `flag:"a-duration" help:"duration flag"`

	AStringSlice []string `help:"string slice flag"`

	SubThing SubThing `flag:"subthing"`
}

// SubThing exists to test nested structs.
type SubThing struct {
	SubBool bool `flag:"a-bool" help:"nested boolean flag"`
}

// Run implements the Runner interface.
func (m *MyMain) Run() error {
	return fmt.Errorf("mymain error")
}

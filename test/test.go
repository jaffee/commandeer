package test

import (
	"fmt"
	"net"
	"time"
)

// MyMain defines a variety of different field types and exercises various
// different tags.
type MyMain struct {
	Thing string `flag:"thing" help:"does a thing"`
	wing  string `flag:"wing" help:"does a wing"`
	Bing  string `flag:"-" help:"shouldn't happen"`

	AIPMask  net.IPMask `flag:"ipmask"`
	AIPNet   net.IPNet
	AIP      net.IP
	AIPSlice []net.IP

	Abool     bool `flag:"a-bool" help:"boolean flag" short:"b"`
	Aint      int  `flag:"a-int" help:"int flag"`
	Aint8     int8
	Aint16    int16 `json:"anint16"`
	Aint32    int32
	Aint64    int64   `flag:"a-int64" help:"int64 flag"`
	Afloat    float64 `flag:"a-float" help:"float flag"`
	Afloat32  float32
	Auint     uint `flag:"a-uint" help:"uint flag"`
	Auint8    uint8
	Auint16   uint16
	Auint32   uint32
	Auint64   uint64        `flag:"a-uint64" help:"uint64 flag"`
	Aduration time.Duration `flag:"a-duration" help:"duration flag"`

	AStringSlice []string `help:"string slice flag"`
	ABoolSlice   []bool
	AIntSlice    []int
	AUintSlice   []uint

	SubThing SubThing `flag:"subthing"`
}

func NewMyMain() *MyMain {
	return &MyMain{
		Thing:    "Thing",
		AIPMask:  net.IPv4Mask(1, 2, 3, 0),
		AIPNet:   net.IPNet{IP: net.IPv4(1, 2, 3, 4), Mask: net.IPv4Mask(1, 2, 3, 0)},
		AIP:      net.IP{1, 2, 3, 4},
		AIPSlice: []net.IP{},

		Abool:     false,
		Aint:      1000000000,
		Aint8:     -7,
		Aint16:    -30000,
		Aint32:    -2000000000,
		Aint64:    -9999999999,
		Afloat:    12.2,
		Afloat32:  -12.3,
		Auint:     9,
		Auint8:    10,
		Auint16:   11,
		Auint32:   12,
		Auint64:   13,
		Aduration: time.Second,

		AStringSlice: []string{"hey", "there"},
		ABoolSlice:   []bool{true, false},
		AIntSlice:    []int{9, -8, 7},
		AUintSlice:   []uint{7, 8, 9},

		SubThing: SubThing{
			SubBool: true,
			Recursion: Recursion{
				NestBool: true,
			},
		},
	}
}

// SubThing exists to test nested structs.
type SubThing struct {
	SubBool   bool      `flag:"a-bool" help:"nested boolean flag"`
	Recursion Recursion `flag:"recursion"`
}

// Recursion exists to test multiple levels of nested structs.
type Recursion struct {
	NestBool bool `flag:"b-bool" help:"doubly nested boolean flag"`
}

// Run implements the Runner interface.
func (m *MyMain) Run() error {
	return fmt.Errorf("mymain error")
}

type SimpleMain struct {
	One    string
	Two    int
	Three  int64
	Four   bool
	Five   uint
	Six    uint64
	Seven  float64
	Eight  time.Duration
	Nine   []string
	Ten    time.Time
	Eleven MyDuration
}

func NewSimpleMain() *SimpleMain {
	return &SimpleMain{
		One:    "one",
		Two:    2,
		Three:  3,
		Four:   true,
		Five:   5,
		Six:    6,
		Seven:  7.0,
		Eight:  time.Millisecond * 8,
		Nine:   []string{"9", "nine"},
		Ten:    time.Unix(0, 0).UTC(),
		Eleven: MyDuration(time.Second * 11),
	}
}

func (m *SimpleMain) Run() error {
	return fmt.Errorf("SimpleMain error")
}

type MyDuration time.Duration

// String returns the string representation of the duration.
func (d MyDuration) String() string { return time.Duration(d).String() }

func (d *MyDuration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	*d = MyDuration(v)
	return nil
}

// MarshalText writes duration value in text format.
func (d MyDuration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

package commandeer

import (
	"fmt"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

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

type SubThing struct {
	SubBool bool `flag:"a-bool" help:"nested boolean flag"`
}

func (m *MyMain) Run() error {
	return fmt.Errorf("mymain error")
}

func TestZeroStruct(t *testing.T) {
	fs := pflag.NewFlagSet("myset", pflag.ContinueOnError)
	mm := &MyMain{}
	err := SetFlags(fs, mm)
	if err != nil {
		t.Fatalf("setting flags with zero MyMain: %v", err)
	}
	err = fs.Parse([]string{"-h"})
	if err != nil && err != pflag.ErrHelp {
		t.Fatalf("parsing help flag: %v", err)
	}

}

func TestCom(t *testing.T) {

	mm := &MyMain{
		Thing:        "blahhh",
		Abool:        true,
		Aint:         -1,
		Aint64:       -987987987987,
		Afloat:       12.23,
		Auint:        1,
		Auint64:      987987987987,
		Aduration:    time.Second * 3,
		AStringSlice: []string{"hello", "goodbye"},
		SubThing: SubThing{
			SubBool: true,
		},
	}
	com, err := Cobra(mm)
	if err != nil {
		t.Fatal(err)
	}
	f := com.Flags().Lookup("thing")
	if f.Name != "thing" || f.Usage != "does a thing" || f.DefValue != "blahhh" {
		t.Fatalf("flag 'thing' not properly defined")
	}
	f = com.Flags().Lookup("a-bool")
	if f.Name != "a-bool" || f.Usage != "boolean flag" || f.DefValue != "true" {
		t.Fatalf("flag 'a-bool' not properly defined")
	}
	f = com.Flags().ShorthandLookup("b")
	if f.Name != "a-bool" || f.Usage != "boolean flag" || f.DefValue != "true" {
		t.Fatalf("shorthand for 'a-bool' not properly defined")
	}
	f = com.Flags().Lookup("a-int")
	if f.Name != "a-int" || f.Usage != "int flag" || f.DefValue != "-1" {
		t.Fatalf("flag 'a-int' not properly defined")
	}
	f = com.Flags().Lookup("a-int64")
	if f.Name != "a-int64" || f.Usage != "int64 flag" || f.DefValue != "-987987987987" {
		t.Fatalf("flag 'a-int64' not properly defined")
	}
	f = com.Flags().Lookup("a-float")
	if f.Name != "a-float" || f.Usage != "float flag" || f.DefValue != "12.23" {
		t.Fatalf("flag 'a-float' not properly defined")
	}
	f = com.Flags().ShorthandLookup("f")
	if f.Name != "a-float" || f.Usage != "float flag" || f.DefValue != "12.23" {
		t.Fatalf("shorthand for 'a-float' not properly defined")
	}
	f = com.Flags().Lookup("a-uint")
	if f.Name != "a-uint" || f.Usage != "uint flag" || f.DefValue != "1" {
		t.Fatalf("flag 'a-uint' not properly defined")
	}
	f = com.Flags().Lookup("a-uint64")
	if f.Name != "a-uint64" || f.Usage != "uint64 flag" || f.DefValue != "987987987987" {
		t.Fatalf("flag 'a-uint64' not properly defined")
	}
	f = com.Flags().Lookup("a-duration")
	if f.Name != "a-duration" || f.Usage != "duration flag" || f.DefValue != "3s" {
		t.Fatalf("flag 'a-duration' not properly defined")
	}
	f = com.Flags().Lookup("a-string-slice")
	if f.Name != "a-string-slice" || f.Usage != "string slice flag" || f.DefValue != "[hello,goodbye]" {
		t.Fatalf("flag 'a-string-slice' not properly defined")
	}
	f = com.Flags().Lookup("subthing.a-bool")
	if f.Name != "subthing.a-bool" || f.Usage != "nested boolean flag" || f.DefValue != "true" {
		t.Fatalf("flag 'subthing.a-bool' not properly defined")
	}
	com.Flags().VisitAll(func(flag *pflag.Flag) {
		fmt.Println(flag.Name)
		if flag.Name == "bing" || flag.Name == "wing" || flag.Name == "-" || flag.Usage == "shouldn't happen" {
			t.Fatalf("explicitly ignored flag is present: %v", flag)
		}
	})

	err = com.Execute()
	if err.Error() != "mymain error" {
		t.Fatalf("unexpected execution error: %v", err)
	}
}

func TestCobraFail(t *testing.T) {
	mm := MyMain{Thing: "blah"}

	_, err := Cobra(mm)
	if err == nil {
		t.Fatalf("Should have gotten an error passing non-pointer to Cobra")
	}

	m := make(map[string]int)
	_, err = Cobra(m)
	if err == nil {
		t.Fatalf("Should have gotten an error passing map to Cobra")
	}

	_, err = Cobra(&m)
	if err == nil {
		t.Fatalf("Should have gotten an error passing pointer to map to Cobra")
	}
}

func TestDowncaseAndDash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "a",
			expected: "a",
		},
		{
			input:    "A",
			expected: "a",
		},
		{
			input:    "aB",
			expected: "a-b",
		},
		{
			input:    "Ba",
			expected: "ba",
		},
		{
			input:    "HelloFriend",
			expected: "hello-friend",
		},
		{
			input:    "helloFriend",
			expected: "hello-friend",
		},
		{
			input:    "AAA",
			expected: "aaa",
		},
		{
			input:    "myURL",
			expected: "my-url",
		},
		{
			input:    "MyURL",
			expected: "my-url",
		},
		{
			input:    "URLFinder",
			expected: "url-finder",
		},
		{
			input:    "AaURLFinder",
			expected: "aa-url-finder",
		},
		{
			input:    "aURLFinder",
			expected: "a-url-finder",
		},
		{
			input:    "AURLFinder",
			expected: "aurl-finder", // NOTE: no easy way to handle this properly
		},
		{
			input:    "KissAToad",
			expected: "kiss-a-toad",
		},
		{
			input:    "IAmAToad",
			expected: "i-am-a-toad",
		},
	}

	for i, tst := range tests {
		output := downcaseAndDash(tst.input)
		if output != tst.expected {
			t.Errorf("test: %d, '%v' is not '%v'", i, output, tst.expected)
		}
	}
}

package commandeer

import (
	"testing"

	"github.com/spf13/pflag"
)

type MyMain struct {
	Thing string `flag:"thing" help:"does a thing"`
	Bing  string `flag:"-" help:"shouldn't happen"`
}

func TestCom(t *testing.T) {

	mm := &MyMain{Thing: "blahhh"}
	com, err := Cobra(mm)
	if err != nil {
		t.Fatal(err)
	}
	f := com.Flags().Lookup("thing")
	if f.Name != "thing" && f.Usage != "does a thing" {
		t.Fatalf("flag 'thing' not properly defined")
	}
	com.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name == "bing" || flag.Name == "-" || flag.Usage == "shouldn't happen" {
			t.Fatalf("explicitly ignored flag is present: %v", flag)
		}
	})
}

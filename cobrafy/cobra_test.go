package cobrafy

import (
	"fmt"
	"testing"
	"time"

	"github.com/jaffee/commandeer/test"
	"github.com/spf13/pflag"
)

func TestCom(t *testing.T) {

	mm := &test.MyMain{
		Thing:        "blahhh",
		Abool:        true,
		Aint:         -1,
		Aint64:       -987987987987,
		Afloat:       12.23,
		Auint:        1,
		Auint64:      987987987987,
		Aduration:    time.Second * 3,
		AStringSlice: []string{"hello", "goodbye"},
		SubThing: test.SubThing{
			SubBool: true,
		},
	}
	com, err := Command(mm)
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
	mm := test.MyMain{Thing: "blah"}

	_, err := Command(mm)
	if err == nil {
		t.Fatalf("Should have gotten an error passing non-pointer to Cobra")
	}

	m := make(map[string]int)
	_, err = Command(m)
	if err == nil {
		t.Fatalf("Should have gotten an error passing map to Cobra")
	}

	_, err = Command(&m)
	if err == nil {
		t.Fatalf("Should have gotten an error passing pointer to map to Cobra")
	}
}

func TestExecute(t *testing.T) {
	mm := test.MyMain{Thing: "blah"}
	err := Execute(&mm)
	if err.Error() != "mymain error" {
		t.Fatalf("wrong error executing MyMain: %v", err)
	}
}

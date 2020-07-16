package pflag_test

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/jaffee/commandeer"
	compflag "github.com/jaffee/commandeer/pflag"
	"github.com/jaffee/commandeer/test"
	"github.com/spf13/pflag"
)

func TestLoadArgsEnvPflag(t *testing.T) {
	mm := test.NewSimpleMain()

	flags := &compflag.FlagSet{pflag.NewFlagSet("tst", pflag.ContinueOnError)}
	err := commandeer.LoadArgsEnv(flags, mm, []string{"--nine=8,7", "--eleven=11m30s"}, "TZT", nil)
	if err != nil {
		t.Fatalf("loading args env: %v", err)
	}

	if !reflect.DeepEqual(mm.Nine, []string{"8", "7"}) {
		t.Errorf("unexpected string slice: %v", mm.Nine)
	}

	if time.Duration(mm.Eleven) != time.Minute*11+time.Second*30 {
		t.Errorf("unexpected value for field 11 (wrapped Duration type)")
	}

}

func TestNames(t *testing.T) {
	m := test.NewMyMain()
	flags := &compflag.FlagSet{pflag.NewFlagSet("tstsimplemain", pflag.ContinueOnError)}
	err := commandeer.Flags(flags, m)
	if err != nil {
		t.Fatalf("getting flags for MyMain: %v", err)
	}
	flagNames := flags.Flags()
	sort.Strings(flagNames)
	expect := []string{
		"a-bool",
		"a-bool-slice",
		"a-duration",
		"a-float",
		"a-int",
		"a-int-slice",
		"a-int64",
		"a-string-slice",
		"a-uint",
		"a-uint-slice",
		"a-uint64",
		"afloat32",
		"aint32",
		"aint8",
		"aip",
		"aip-net",
		"aip-slice",
		"anint16",
		"auint16",
		"auint32",
		"auint8",
		"ipmask",
		"subthing.a-bool",
		"subthing.recursion.b-bool",
		"thing",
	}
	if !reflect.DeepEqual(expect, flagNames) {
		t.Fatalf("expected %v but got %v", expect, flagNames)
	}
}

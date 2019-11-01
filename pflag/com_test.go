package pflag_test

import (
	"reflect"
	"testing"

	"github.com/jaffee/commandeer"
	compflag "github.com/jaffee/commandeer/pflag"
	"github.com/jaffee/commandeer/test"
	"github.com/spf13/pflag"
)

func TestLoadArgsEnvPflag(t *testing.T) {
	mm := test.NewSimpleMain()

	flags := &compflag.FlagSet{pflag.NewFlagSet("tst", pflag.ContinueOnError)}
	err := commandeer.LoadArgsEnv(flags, mm, []string{"--nine=8,7"}, "TZT", nil)
	if err != nil {
		t.Fatalf("loading args env: %v", err)
	}

	if !reflect.DeepEqual(mm.Nine, []string{"8", "7"}) {
		t.Fatalf("unexpected string slice: %v", mm.Nine)
	}

}

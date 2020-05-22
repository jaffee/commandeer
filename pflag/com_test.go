package pflag_test

import (
	"reflect"
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

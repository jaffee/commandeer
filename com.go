package commandeer

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Cobra takes a struct pointer (optionally with tagged fields), and produces a
// cobra.Command with flags set up to populate the values of the struct.
func Cobra(main interface{}) (*cobra.Command, error) {
	typ := reflect.TypeOf(main)
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("value must be Struct, but is %s", typ.Kind())
	}

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	com := &cobra.Command{
		Use: strings.ToLower(mainTyp.Name()),
		// TODO get short and long desc from docstrings somehow?
	}
	// TODO define RunE?
	flags := com.Flags()
	SetFlags(flags, main)

	return com, nil
}

func ImplementsRunner(t reflect.Type) bool {
	var run Runner
	runType := reflect.TypeOf(run)
	return t.Implements(runType)
}

func SetFlags(flags Flagger, main interface{}) {
	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()

	for i := 0; i < mainTyp.NumField(); i++ {
		ft := mainTyp.Field(i)
		f := mainVal.Field(i)
		switch ft.Type.Kind() {
		case reflect.String:
			if ft.PkgPath != "" {
				continue // this field is unexported
			}
			flagName := flagName(ft)
			if flagName == "-" || flagName == "" {
				continue // explicitly ignored
			}
			p := f.Addr().Interface().(*string)
			flags.StringVar(p, flagName, f.String(), flagHelp(ft))
		}
	}

}

func flagName(field reflect.StructField) (flagname string) {
	if flagname, ok := field.Tag.Lookup("flag"); ok {
		return flagname
	}

	if flagname, ok := field.Tag.Lookup("json"); ok {
		return flagname
	}
	flagname = field.Name
	return flagname
}

func flagHelp(field reflect.StructField) (flaghelp string) {
	if flaghelp, ok := field.Tag.Lookup("help"); ok {
		return flaghelp
	}
	return ""
}

type Flagger interface {
	StringVar(p *string, name string, value string, usage string)
	IntVar(p *int, name string, value int, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	BoolVar(p *bool, name string, value bool, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
}

type PFlagger interface {
	Flagger
	StringSliceVar(p *[]string, name string, value []string, usage string)
	BoolSliceVar(p *[]bool, name string, value []bool, usage string)
	Float32Var(p *float32, name string, value float32, usage string)
	IPMaskVar(p *net.IPMask, name string, value net.IPMask, usage string)
	IPSliceVar(p *[]net.IP, name string, value []net.IP, usage string)
	IPNetVar(p *net.IPNet, name string, value net.IPNet, usage string)
	IPVar(p *net.IP, name string, value net.IP, usage string)
	Int32Var(p *int32, name string, value int32, usage string)
	Uint16Var(p *uint16, name string, value uint16, usage string)
	Uint32Var(p *uint32, name string, value uint32, usage string)
	Uint8Var(p *uint8, name string, value uint8, usage string)
	UintSliceVar(p *[]uint, name string, value []uint, usage string)
	IntSliceVar(p *[]int, name string, value []int, usage string)
	Int8Var(p *int8, name string, value int8, usage string)
}

type Runner interface {
	Run() error
}

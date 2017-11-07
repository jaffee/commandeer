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
		return nil, fmt.Errorf("value must be pointer to struct, but is %s", typ.Kind())
	}

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	if mainTyp.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value must be pointer to struct, but is pointer to %s", typ.Kind())
	}
	com := &cobra.Command{
		Use: strings.ToLower(mainTyp.Name()),
		// TODO get short and long desc from docstrings somehow?
	}
	if ImplementsRunner(typ) {
		com.RunE = func(cmd *cobra.Command, args []string) error {
			return main.(Runner).Run()
		}
	}
	flags := com.Flags()
	err := SetFlags(flags, main)
	if err != nil {
		return nil, err
	}
	return com, nil
}

func ImplementsRunner(t reflect.Type) bool {
	runType := reflect.TypeOf((*Runner)(nil)).Elem()
	return t.Implements(runType)
}

func SetFlags(flags Flagger, main interface{}) error {
	return setFlags(flags, main, "")
}

func setFlags(flags Flagger, main interface{}, prefix string) error {
	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()

	for i := 0; i < mainTyp.NumField(); i++ {
		ft := mainTyp.Field(i)
		f := mainVal.Field(i)
		if ft.PkgPath != "" {
			continue // this field is unexported
		}
		flagName := flagName(ft, prefix)
		if flagName == "-" || flagName == "" {
			continue // explicitly ignored
		}
		switch f.Interface().(type) {
		case time.Duration:
			p := f.Addr().Interface().(*time.Duration)
			flags.DurationVar(p, flagName, time.Duration(f.Int()), flagHelp(ft))
			continue
		}
		switch ft.Type.Kind() {
		case reflect.String:
			p := f.Addr().Interface().(*string)
			flags.StringVar(p, flagName, f.String(), flagHelp(ft))
		case reflect.Bool:
			p := f.Addr().Interface().(*bool)
			flags.BoolVar(p, flagName, f.Bool(), flagHelp(ft))
		case reflect.Int:
			p := f.Addr().Interface().(*int)
			val := int(f.Int())
			flags.IntVar(p, flagName, val, flagHelp(ft))
		case reflect.Int64:
			p := f.Addr().Interface().(*int64)
			flags.Int64Var(p, flagName, f.Int(), flagHelp(ft))
		case reflect.Float64:
			p := f.Addr().Interface().(*float64)
			flags.Float64Var(p, flagName, f.Float(), flagHelp(ft))
		case reflect.Uint:
			p := f.Addr().Interface().(*uint)
			val := uint(f.Uint())
			flags.UintVar(p, flagName, val, flagHelp(ft))
		case reflect.Uint64:
			p := f.Addr().Interface().(*uint64)
			flags.Uint64Var(p, flagName, f.Uint(), flagHelp(ft))
		case reflect.Struct:
			var newprefix string
			if prefix != "" {
				newprefix = prefix + "." + flagName
			} else {
				newprefix = flagName
			}
			err := setFlags(flags, f.Addr().Interface(), newprefix)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("encountered unsupported field type/kind: %#v at %s", f, prefix)
		}
	}
	return nil
}

func flagName(field reflect.StructField, prefix string) (flagname string) {
	var ok bool
	// unnecessary and confusing... but awesome.
	defer func() {
		if prefix != "" {
			flagname = prefix + "." + flagname
		}
	}()
	if flagname, ok = field.Tag.Lookup("flag"); ok {
		return flagname
	}

	if flagname, ok = field.Tag.Lookup("json"); ok {
		return flagname
	}
	flagname = field.Name
	// TODO convert from camel case to lower with dashes
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

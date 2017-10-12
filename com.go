package commandeer

import (
	"fmt"
	"reflect"
	"strings"

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
				continue // ignored
			}
			p := f.Addr().Interface().(*string)
			flags.StringVarP(p, flagName, "", f.String(), flagHelp(ft))
		}
	}

	return com, nil
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

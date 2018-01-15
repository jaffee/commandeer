package cobrafy

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jaffee/commandeer"
	"github.com/spf13/cobra"
)

// Command takes a struct pointer (optionally with tagged fields), and produces a
// cobra.Command with flags set up to populate the values of the struct.
func Command(main interface{}) (*cobra.Command, error) {
	typ := reflect.TypeOf(main)
	if typ.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("value must be pointer to struct, but is %s", typ.Kind())
	}

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	if mainTyp.Kind() != reflect.Struct {
		return nil, fmt.Errorf("value must be pointer to struct, but is pointer to %s", typ.Kind())
	}
	mainPkg := mainTyp.PkgPath()
	mainPkgLast := mainPkg[strings.LastIndex(mainPkg, "/")+1:]
	com := &cobra.Command{
		Use: mainPkgLast,
		// TODO get short and long desc from docstrings somehow?
	}
	if commandeer.ImplementsRunner(typ) {
		com.RunE = func(cmd *cobra.Command, args []string) error {
			return main.(commandeer.Runner).Run()
		}
	}
	flags := com.Flags()
	err := commandeer.Flags(flags, main)
	if err != nil {
		return nil, err
	}

	return com, nil
}

// Execute creates a cobra.Command using cobrafy.Command and then calls its
// Execute() method, returning the error.
func Execute(main interface{}) error {
	com, err := Command(main)
	if err != nil {
		return fmt.Errorf("getting command: %v", err)
	}

	return com.Execute()
}

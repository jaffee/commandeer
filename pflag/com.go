package pflag

import (
	"os"

	"github.com/jaffee/commandeer"
	"github.com/spf13/pflag"
)

// Validate interface
var _ = commandeer.FlagNamer(&FlagSet{})

// FlagSet is an extension to *pflag.FlagSet that satisfies FlagNamer
type FlagSet struct {
	*pflag.FlagSet
}

// Flags returns a slice of flag names
func (f *FlagSet) Flags() (flags []string) {
	f.VisitAll(func(f *pflag.Flag) {
		flags = append(flags, f.Name)
	})
	return flags
}

// LoadEnv calls LoadArgsEnv with args from the command line and the
// default flag set.
func LoadEnv(main interface{}, envPrefix string, parseElsewhere func(main interface{}) error) error {
	return commandeer.LoadArgsEnv(&FlagSet{pflag.CommandLine}, main, os.Args[1:], envPrefix, parseElsewhere)
}

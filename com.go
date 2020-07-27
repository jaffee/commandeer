// Package commandeer sets up command line flags based on the fields and field
// tags of a struct. It helps ease common pains of CLI app development by
// allowing you to unobtrusively define flags in a library package while having
// a tiny main package which calls commandeer.Run* or commandeer.Flags.
//
// Run is the usual interface to commandeer, but it requires you to pass in a
// struct which has a "Run() error" method. RunArgs works similarly, but allows
// you to pass in the args to be parsed and the flag set to be used. In cases
// where your struct doesn't have a Run() method, or you don't want to call it,
// the Flags() function takes in a FlagSet and sets the flags based on the
// passed in struct in the same way.
package commandeer

import (
	"encoding"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Flags sets up the given Flagger (usually an instance of flag.FlagSet or
// pflag.FlagSet). The second argument, "main", must be a pointer to a struct. A
// flag will be created for each exported field of the struct which isn't
// explicitly ignored.
//
// Struct tags are used to control the behavior of Flags(), though none are
// necessary.
//
// 1. The "help" tag on a field is used to populate the usage string for that
// field's flag.
//
// 2. The "flag" tag on a field will be used as the name of that field's flag.
// Set it to "-" to ignore this field when creating flags. If it does not exist,
// the "json" tag will be used, and if it also does not exist, the field name
// will be downcased and converted from camel case to be dash separated.
//
// 3. The "short" tag on a field will be used as the shorthand flag for that
// field. It should be a single ascii character. This will only be used if the
// Flagger is also a PFlagger.
func Flags(flags Flagger, main interface{}) error {
	typ := reflect.TypeOf(main)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("value must be pointer to struct, but is %s", typ.Kind())
	}

	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()
	if mainTyp.Kind() != reflect.Struct {
		return fmt.Errorf("value must be pointer to struct, but is pointer to %s", typ.Kind())
	}

	return setFlags(newFlagTracker(flags), main, "")
}

type flagSet struct {
	*flag.FlagSet
}

func (f *flagSet) Flags() (flags []string) {
	f.VisitAll(func(f *flag.Flag) {
		flags = append(flags, f.Name)
	})
	return flags
}

var _ = FlagNamer(&flagSet{})

// Run runs "main" which must be a pointer to a struct which implements the
// Runner interface. It first calls Flags to set up command line flags based on
// "main" (see the documentation for Flags).
func Run(main interface{}) error {
	return RunArgs(&flagSet{flag.CommandLine}, main, os.Args[1:])
}

var replacer *strings.Replacer = strings.NewReplacer("-", "_", ".", "_")

func envNorm(name string) string {
	return strings.ToUpper(replacer.Replace(name))
}

// loadEnv visits each flag in the FlagSet and sets its value based on
// OS environment.
func loadEnv(flagger Flagger, prefix string) (err error) {
	if namer, ok := flagger.(FlagNamer); ok {
		for _, name := range namer.Flags() {
			envString := envNorm(prefix + name)
			val, ok := os.LookupEnv(envString)
			if ok {
				err = flagger.Set(name, val)
				if err != nil {
					return fmt.Errorf("couldn't set %s to %s from env %s: %v", name, val, envString, err)
				}
			}
		}
	} else {
		return fmt.Errorf("unable to load flags from environment: flagger does not implement FlagNamer")
	}
	return nil
}

// LoadEnv calls LoadArgsEnv with args from the command line and the
// default flag set.
func LoadEnv(main interface{}, envPrefix string, parseElsewhere func(main interface{}) error) error {
	return LoadArgsEnv(&flagSet{flag.CommandLine}, main, os.Args[1:], envPrefix, parseElsewhere)
}

// LoadArgsEnv uses Flags to define flags based on "main", then it
// tries setting each flag's value from the OS environment based on a
// prefix concatenated to the flag name. The flag name is normalized
// by removing any dashes or dots and replacing them with
// underscores.
//
// One may also pass a "configElsewhere" function which can operate on
// main arbitrarily. The purpose of this is to load config values from
// (e.g.) a file without this package needing to import packages for
// parsing specific file formats.
//
// Flags set via args take the highest precedence, followed by the
// environment, followed by configElsewhere (followed by
// defaults). Command line args and environment variables are set on
// main before it is passed to configElsewhere so that configElsewhere
// can be configured (such as with a path to a config file). Once
// configElsewhere runs, the environment and command line args are
// re-set since they take higher precedence.
func LoadArgsEnv(flags Flagger, main interface{}, args []string, envPrefix string, configElsewhere func(main interface{}) error) error {
	// setup flags
	err := Flags(flags, main)
	if err != nil {
		return fmt.Errorf("calling Flags: %v", err)
	}
	// set values based on environment
	err = loadEnv(flags, envPrefix)
	if err != nil {
		return fmt.Errorf("loading environment: %v", err)
	}
	// set values based on command line
	err = flags.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing command line args: %v", err)
	}
	// set values with configElsewhere
	if configElsewhere != nil {
		err = configElsewhere(main)
		if err != nil {
			return fmt.Errorf("executing external parsing func: %v", err)
		}
	}
	// reset values with environment (precedence over configElsewhere)
	err = loadEnv(flags, envPrefix)
	if err != nil {
		return fmt.Errorf("reloading environment: %v", err)
	}
	// reset values with command line args (highest precedence)
	err = flags.Parse(args)
	if err != nil {
		return fmt.Errorf("reparsing command line args: %v", err)
	}
	return nil
}

// RunArgs is similar to Run, but the caller must specify their own flag set and
// args to be parsed by that flag set.
func RunArgs(flags Flagger, main interface{}, args []string) error {
	err := Flags(flags, main)
	if err != nil {
		return fmt.Errorf("calling Flags: %v", err)
	}
	err = flags.Parse(args)
	if err != nil {
		return fmt.Errorf("parsing flags: %v", err)
	}

	if main, ok := main.(Runner); ok {
		return main.Run()
	}
	return fmt.Errorf("called 'Run' with something which doesn't implement the 'Run() error' method.")
}

type stringSliceValue struct {
	value *[]string
}

func (s stringSliceValue) Set(val string) error {
	*s.value = strings.Split(val, ",")
	return nil
}

func (s stringSliceValue) String() string {
	if s.value != nil {
		return "[" + strings.Join(*s.value, ",") + "]"
	}
	return ""
}

func (s stringSliceValue) Type() string {
	return "strings"
}

func setFlags(flags *flagTracker, main interface{}, prefix string) error {
	// TODO add tracking of flag names to ensure no duplicates
	mainVal := reflect.ValueOf(main).Elem()
	mainTyp := mainVal.Type()

	for i := 0; i < mainTyp.NumField(); i++ {
		ft := mainTyp.Field(i)
		f := mainVal.Field(i)
		if ft.PkgPath != "" {
			continue // this field is unexported
		}
		flagName := flagName(ft)
		if flagName == "-" || flagName == "" {
			continue // explicitly ignored
		}
		shorthand, err := flags.short(ft, flagName)
		if err != nil {
			return fmt.Errorf("getting shorthand for '%v': %v", ft.Name, err)
		}
		if prefix != "" {
			flagName = prefix + "." + flagName
		}

		// first check supported concrete types
		switch p := f.Addr().Interface().(type) {
		case *time.Duration:
			flags.duration(p, flagName, shorthand, time.Duration(f.Int()), flagHelp(ft))
			continue
		case *net.IPMask:
			if !flags.pflag {
				return fmt.Errorf("cannot support net.IPMask field at '%v' with stdlib flag pkg.", flagName)
			}
			flags.ipMask(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case *net.IPNet:
			if !flags.pflag {
				return fmt.Errorf("cannot support net.IPNet field at '%v' with stdlib flag pkg.", flagName)
			}
			flags.ipNet(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case *net.IP:
			if !flags.pflag {
				return fmt.Errorf("cannot support net.IP field at '%v' with stdlib flag pkg.", flagName)
			}
			flags.ip(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case *[]net.IP:
			if !flags.pflag {
				return fmt.Errorf("cannot support []net.IP field at '%v' with stdlib flag pkg.", flagName)
			}
			flags.ipSlice(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case *[]string:
			// special case support for string slice. multiple calls
			// to set the string slice value will replace it rather
			// than appending to it (as they would with
			// e.g. pflag). This is necessary for cascading
			// configuration from multiple sources (e.g. file, env,
			// command line).
			flags.vvarp(stringSliceValue{value: p}, flagName, shorthand, flagHelp(ft))
			continue
		case encodable:
			flags.vvarp(encodedValue{p}, flagName, shorthand, flagHelp(ft))
			continue
		}

		// now check basic kinds
		switch ft.Type.Kind() {
		case reflect.String:
			p := f.Addr().Interface().(*string)
			flags.string(p, flagName, shorthand, f.String(), flagHelp(ft))
		case reflect.Bool:
			p := f.Addr().Interface().(*bool)
			flags.bool(p, flagName, shorthand, f.Bool(), flagHelp(ft))
		case reflect.Int:
			p := f.Addr().Interface().(*int)
			val := int(f.Int())
			flags.int(p, flagName, shorthand, val, flagHelp(ft))
		case reflect.Int64:
			p := f.Addr().Interface().(*int64)
			flags.int64(p, flagName, shorthand, f.Int(), flagHelp(ft))
		case reflect.Float64:
			p := f.Addr().Interface().(*float64)
			flags.float64(p, flagName, shorthand, f.Float(), flagHelp(ft))
		case reflect.Uint:
			p := f.Addr().Interface().(*uint)
			val := uint(f.Uint())
			flags.uint(p, flagName, shorthand, val, flagHelp(ft))
		case reflect.Uint64:
			p := f.Addr().Interface().(*uint64)
			flags.uint64(p, flagName, shorthand, f.Uint(), flagHelp(ft))
		case reflect.Slice:
			if !flags.pflag {
				return fmt.Errorf("cannot support slice field at '%v' with stdlib flag pkg.", flagName)
			}
			switch ft.Type.Elem().Kind() {
			case reflect.String:
				p := f.Addr().Interface().(*[]string)
				flags.stringSlice(p, flagName, shorthand, *p, flagHelp(ft))
			case reflect.Bool:
				p := f.Addr().Interface().(*[]bool)
				flags.boolSlice(p, flagName, shorthand, *p, flagHelp(ft))
			case reflect.Int:
				p := f.Addr().Interface().(*[]int)
				flags.intSlice(p, flagName, shorthand, *p, flagHelp(ft))
			case reflect.Uint:
				p := f.Addr().Interface().(*[]uint)
				flags.uintSlice(p, flagName, shorthand, *p, flagHelp(ft))
			default:
				return fmt.Errorf("encountered unsupported slice type/kind: %#v at %s", f, prefix)
			}
		case reflect.Float32:
			if !flags.pflag {
				return fmt.Errorf("cannot support float32 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*float32)
			flags.float32(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Int16:
			if !flags.pflag {
				return fmt.Errorf("cannot support int16 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*int16)
			flags.int16(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Int32:
			if !flags.pflag {
				return fmt.Errorf("cannot support int32 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*int32)
			flags.int32(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Uint16:
			if !flags.pflag {
				return fmt.Errorf("cannot support uint16 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*uint16)
			flags.uint16(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Uint32:
			if !flags.pflag {
				return fmt.Errorf("cannot support uint32 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*uint32)
			flags.uint32(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Uint8:
			if !flags.pflag {
				return fmt.Errorf("cannot support uint8 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*uint8)
			flags.uint8(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Int8:
			if !flags.pflag {
				return fmt.Errorf("cannot support int8 field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*int8)
			flags.int8(p, flagName, shorthand, *p, flagHelp(ft))
		case reflect.Struct:
			var newprefix string
			// TODO test, what happens if there are flag name
			// collisions (e.g. the struct at this level and the
			// !embed struct have a field with the same name)?
			if flagName == "!embed" {
				newprefix = prefix
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

// flagName finds a field's flag name. It first looks for a "flag" tag, then
// tries to use the "json" tag, and final falls back to using the name of the
// field after running it through "downcaseAndDash".
func flagName(field reflect.StructField) (flagname string) {
	var ok bool
	if flagname, ok = field.Tag.Lookup("flag"); ok {
		return flagname
	}

	if flagname, ok = field.Tag.Lookup("json"); ok {
		return flagname
	}
	flagname = field.Name

	return downcaseAndDash(flagname)
}

// downcaseAndDash converts a field name (expected to be camel case) to an all
// lower case flag name with dashes between words that were previously cameled.
// It attempts to handle upper case acronyms properly as well.
func downcaseAndDash(input string) string {
	ret := make([]rune, 0)
	lastUpper := false
	nextUpper := false
	for i, chr := range input {
		if i+1 < len(input) {
			nextUpper = unicode.IsUpper(rune(input[i+1]))
		}
		if unicode.IsUpper(chr) {
			if len(ret) == 0 || (lastUpper && nextUpper) {
				ret = append(ret, unicode.ToLower(chr))
			} else {
				ret = append(ret, '-', unicode.ToLower(chr))
			}
			lastUpper = true
		} else {
			ret = append(ret, chr)
			lastUpper = false
		}
	}
	return string(ret)
}

// flagHelp gets the help text from a field's tag or returns an empty string.
func flagHelp(field reflect.StructField) (flaghelp string) {
	if flaghelp, ok := field.Tag.Lookup("help"); ok {
		return flaghelp
	}
	return ""
}

type encodable interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

type encodedValue struct {
	encodable
}

func (v encodedValue) Set(str string) error {
	return v.UnmarshalText([]byte(str))
}

func (v encodedValue) String() string {
	dat, err := v.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(dat)
}

func (v encodedValue) Type() string {
	t := reflect.TypeOf(v.encodable)
	if name := t.Name(); name != "" {
		return name
	}
	return t.String()
}

// flagTracker has methods for managing the set up of flags - it will utilize
// pflag methods if flagger is a PFlagger, and set up short flags as well.
type flagTracker struct {
	flagger  Flagger
	pflagger PFlagger
	pflag    bool
	shorts   map[rune]struct{}
}

// newFlagTracker sets up a flagTracker based on a flagger.
func newFlagTracker(flagger Flagger) *flagTracker {
	fTr := &flagTracker{
		flagger: flagger,
		shorts: map[rune]struct{}{
			'h': {}, // "h" is always used for help, so we can't set it.
		},
	}
	fTr.pflagger, fTr.pflag = flagger.(PFlagger)
	return fTr
}

// short gets the shorthand for a flag. flagName is the non-prefixed name
// returned by flagName.
func (fTr *flagTracker) short(field reflect.StructField, flagName string) (letter string, err error) {
	if short, ok := field.Tag.Lookup("short"); ok {
		if short == "" {
			return "", nil // explicitly set to no shorthand
		}
		runeVal, width := utf8.DecodeRuneInString(short)
		if runeVal == utf8.RuneError || width > 1 {
			return "", fmt.Errorf("'%s' is not a valid single ascii character.", short)
		}
		if _, ok := fTr.shorts[runeVal]; ok {
			return "", fmt.Errorf("'%s' has already been used.", short)
		}
		fTr.shorts[runeVal] = struct{}{}
		return short, nil
	}
	return "", nil // no shorthand char available, but that's ok
}

func (fTr *flagTracker) string(p *string, name, shorthand, value, usage string) {
	if fTr.pflag {
		fTr.pflagger.StringVarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.StringVar(p, name, value, usage)
	}
}
func (fTr *flagTracker) int(p *int, name, shorthand string, value int, usage string) {
	if fTr.pflag {
		fTr.pflagger.IntVarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.IntVar(p, name, value, usage)
	}
}
func (fTr *flagTracker) int64(p *int64, name, shorthand string, value int64, usage string) {
	if fTr.pflag {
		fTr.pflagger.Int64VarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.Int64Var(p, name, value, usage)
	}
}
func (fTr *flagTracker) bool(p *bool, name, shorthand string, value bool, usage string) {
	if fTr.pflag {
		fTr.pflagger.BoolVarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.BoolVar(p, name, value, usage)
	}
}
func (fTr *flagTracker) uint(p *uint, name, shorthand string, value uint, usage string) {
	if fTr.pflag {
		fTr.pflagger.UintVarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.UintVar(p, name, value, usage)
	}
}
func (fTr *flagTracker) uint64(p *uint64, name, shorthand string, value uint64, usage string) {
	if fTr.pflag {
		fTr.pflagger.Uint64VarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.Uint64Var(p, name, value, usage)
	}
}
func (fTr *flagTracker) float64(p *float64, name, shorthand string, value float64, usage string) {
	if fTr.pflag {
		fTr.pflagger.Float64VarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.Float64Var(p, name, value, usage)
	}
}
func (fTr *flagTracker) duration(p *time.Duration, name, shorthand string, value time.Duration, usage string) {
	if fTr.pflag {
		fTr.pflagger.DurationVarP(p, name, shorthand, value, usage)
	} else {
		fTr.flagger.DurationVar(p, name, value, usage)
	}
}

func (fTr *flagTracker) stringSlice(p *[]string, name, shorthand string, value []string, usage string) {
	fTr.pflagger.StringSliceVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) boolSlice(p *[]bool, name, shorthand string, value []bool, usage string) {
	fTr.pflagger.BoolSliceVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) uintSlice(p *[]uint, name, shorthand string, value []uint, usage string) {
	fTr.pflagger.UintSliceVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) intSlice(p *[]int, name, shorthand string, value []int, usage string) {
	fTr.pflagger.IntSliceVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) ipSlice(p *[]net.IP, name, shorthand string, value []net.IP, usage string) {
	fTr.pflagger.IPSliceVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) float32(p *float32, name, shorthand string, value float32, usage string) {
	fTr.pflagger.Float32VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) ipMask(p *net.IPMask, name, shorthand string, value net.IPMask, usage string) {
	fTr.pflagger.IPMaskVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) ipNet(p *net.IPNet, name, shorthand string, value net.IPNet, usage string) {
	fTr.pflagger.IPNetVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) ip(p *net.IP, name, shorthand string, value net.IP, usage string) {
	fTr.pflagger.IPVarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) uint8(p *uint8, name, shorthand string, value uint8, usage string) {
	fTr.pflagger.Uint8VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) uint16(p *uint16, name, shorthand string, value uint16, usage string) {
	fTr.pflagger.Uint16VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) uint32(p *uint32, name, shorthand string, value uint32, usage string) {
	fTr.pflagger.Uint32VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) int8(p *int8, name, shorthand string, value int8, usage string) {
	fTr.pflagger.Int8VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) int16(p *int16, name, shorthand string, value int16, usage string) {
	fTr.pflagger.Int16VarP(p, name, shorthand, value, usage)
}
func (fTr *flagTracker) int32(p *int32, name, shorthand string, value int32, usage string) {
	fTr.pflagger.Int32VarP(p, name, shorthand, value, usage)
}

// vvarp reflectively calls VarP or Var on the underlying pflagger or
// flagger. We can't add VarP to the pflagger interface because it
// takes a pflag.Value and referring to that would necessitate
// importing pflag. This was added because pflag's implementation of string slices appends to the string slice if it is called multiple times rather than resetting it.
func (fTr *flagTracker) vvarp(value Value, name, shorthand, usage string) {
	var flagImpl reflect.Value
	if fTr.pflag {
		flagImpl = reflect.ValueOf(fTr.pflagger)
	} else {
		flagImpl = reflect.ValueOf(fTr.flagger)
	}
	varPMethod := flagImpl.MethodByName("VarP")
	zv := reflect.Value{}
	if varPMethod != zv {
		out := varPMethod.Call([]reflect.Value{reflect.ValueOf(value), reflect.ValueOf(name), reflect.ValueOf(shorthand), reflect.ValueOf(usage)})
		if len(out) > 0 {
			panic("unexpected result after reflectively calling Var on flagger implementation")
		}
		return
	}
	varMethod := flagImpl.MethodByName("Var")
	if varMethod == zv {
		panic("the given flag implementation does not have a Var method")
	}
	out := varMethod.Call([]reflect.Value{reflect.ValueOf(value), reflect.ValueOf(name), reflect.ValueOf(usage)})
	if len(out) > 0 {
		panic("unexpected result after reflectively calling Var on flagger implementation")
	}
}

// Value is a copy of the pflag Value interface which is a superset of flag.Value
type Value interface {
	String() string
	Set(string) error
	Type() string
}

// Flagger is an interface satisfied by flag.FlagSet and other implementations
// of flags.
type Flagger interface {
	Parse([]string) error
	StringVar(p *string, name string, value string, usage string)
	IntVar(p *int, name string, value int, usage string)
	Int64Var(p *int64, name string, value int64, usage string)
	BoolVar(p *bool, name string, value bool, usage string)
	UintVar(p *uint, name string, value uint, usage string)
	Uint64Var(p *uint64, name string, value uint64, usage string)
	Float64Var(p *float64, name string, value float64, usage string)
	DurationVar(p *time.Duration, name string, value time.Duration, usage string)
	Set(name string, value string) error
}

// PFlagger is an extension of the Flagger interface which is implemented by the
// pflag package (github.com/ogier/pflag, or github.com/spf13/pflag)
type PFlagger interface {
	Parse([]string) error
	StringSliceVarP(p *[]string, name string, shorthand string, value []string, usage string)
	BoolSliceVarP(p *[]bool, name string, shorthand string, value []bool, usage string)
	UintSliceVarP(p *[]uint, name string, shorthand string, value []uint, usage string)
	IntSliceVarP(p *[]int, name string, shorthand string, value []int, usage string)
	IPSliceVarP(p *[]net.IP, name string, shorthand string, value []net.IP, usage string)
	Float32VarP(p *float32, name string, shorthand string, value float32, usage string)
	IPMaskVarP(p *net.IPMask, name string, shorthand string, value net.IPMask, usage string)
	IPNetVarP(p *net.IPNet, name string, shorthand string, value net.IPNet, usage string)
	IPVarP(p *net.IP, name string, shorthand string, value net.IP, usage string)
	Int16VarP(p *int16, name string, shorthand string, value int16, usage string)
	Int32VarP(p *int32, name string, shorthand string, value int32, usage string)
	Uint16VarP(p *uint16, name string, shorthand string, value uint16, usage string)
	Uint32VarP(p *uint32, name string, shorthand string, value uint32, usage string)
	Uint8VarP(p *uint8, name string, shorthand string, value uint8, usage string)
	Int8VarP(p *int8, name string, shorthand string, value int8, usage string)

	StringVarP(p *string, name string, shorthand string, value string, usage string)
	IntVarP(p *int, name string, shorthand string, value int, usage string)
	Int64VarP(p *int64, name string, shorthand string, value int64, usage string)
	BoolVarP(p *bool, name string, shorthand string, value bool, usage string)
	UintVarP(p *uint, name string, shorthand string, value uint, usage string)
	Uint64VarP(p *uint64, name string, shorthand string, value uint64, usage string)
	Float64VarP(p *float64, name string, shorthand string, value float64, usage string)
	DurationVarP(p *time.Duration, name string, shorthand string, value time.Duration, usage string)
}

// FlagNamer is an interface that Flaggers may use to list the available flags.
type FlagNamer interface {
	Flags() []string
}

// Runner must be implemented by things passed to the Run and RunArgs methods.
type Runner interface {
	Run() error
}

package commandeer

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
		switch f.Interface().(type) {
		case time.Duration:
			p := f.Addr().Interface().(*time.Duration)
			flags.duration(p, flagName, shorthand, time.Duration(f.Int()), flagHelp(ft))
			continue
		case net.IPMask:
			if !flags.pflag {
				return fmt.Errorf("cannot support net.IPMask field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*net.IPMask)
			flags.ipMask(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case net.IPNet:
			if !flags.pflag {
				return fmt.Errorf("can support net.IPNet field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*net.IPNet)
			flags.ipNet(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case net.IP:
			if !flags.pflag {
				return fmt.Errorf("can support net.IP field at '%v' with stdlib flag pkg.", flagName)
			}
			p := f.Addr().Interface().(*net.IP)
			flags.ip(p, flagName, shorthand, *p, flagHelp(ft))
			continue
		case []net.IP:
			if !flags.pflag {
				return fmt.Errorf("can support []net.IP field at '%v' with stdlib flag pkg.", flagName)
			} else {
				return
			}
			p := f.Addr().Interface().(*[]net.IP)
			flags.ipSlice(p, flagName, shorthand, *p, flagHelp(ft))
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

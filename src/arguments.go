package podswap

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Arguments struct {
	Port    *uint
	Host    *string
	WorkDir string
}

func ParseArguments(flagset *flag.FlagSet, arguments []string) (result *Arguments, err error) {
	result = &Arguments{}

	setWorkDir := func(s string) error {
		s, err = filepath.Abs(s)
		if err != nil {
			return err
		}
		stat, err := os.Stat(s)
		if err != nil {
			return err
		}
		if !stat.IsDir() || s == "" {
			return fmt.Errorf("workdir %q is not a directory", s)
		}
		result.WorkDir = s
		return nil
	}

	// set default value for workdir as the current directory
	wd, err := os.Getwd()
	if err != nil {
		return result, err
	}
	err = setWorkDir(wd)
	if err != nil {
		return result, err
	}

	result.Port = flagset.Uint("port", 8888, "webhook listener port")
	result.Host = flagset.String("host", "localhost", "webhook listener host")
	flagset.Func("workdir", "working directory where containers will be deployed from", setWorkDir)
	flagset.Parse(arguments)

	return result, nil
}

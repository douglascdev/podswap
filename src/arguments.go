package podswap

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Arguments struct {
	BuildCommand  *string
	DeployCommand *string
	WorkDir       string
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

	result.BuildCommand = flagset.String("build-cmd", "docker compose build", "command to run after the webhook is triggered")
	result.DeployCommand = flagset.String("deploy-cmd", "docker compose up -d", "command to run after the build command is done")
	flagset.Func("workdir", "working directory where containers will be deployed from", setWorkDir)
	flagset.Parse(arguments)

	return result, nil
}

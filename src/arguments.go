package podswap

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Arguments struct {
	BuildCommand  *exec.Cmd
	DeployCommand *exec.Cmd
	WorkDir       string
}

func ParseArguments(flagset *flag.FlagSet, arguments []string) (result *Arguments, err error) {
	result = &Arguments{}

	// workdir
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
		slog.Debug("set workdir", slog.String("s", s))
		result.WorkDir = s
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return result, err
	}
	err = setWorkDir(wd)
	if err != nil {
		return result, err
	}
	flagset.Func("workdir", "working directory where containers will be deployed from", setWorkDir)

	// build-cmd
	flagset.Func("build-cmd", "command to run after the webhook is triggered", func(s string) error {
		cmds := strings.Split(s, " ")
		if len(cmds) < 1 {
			return errors.New("build-cmd was not set")
		}
		var (
			cmd     string = cmds[0]
			cmdPath string
			err     error
		)
		if cmdPath, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("command %q not found in path: %v", cmd, err)
		}

		result.BuildCommand = exec.Command(cmdPath, cmds[1:]...)
		slog.Debug("set build-cmd", slog.String("cmdPath", cmdPath), slog.Any("args", cmds[1:]))

		return nil
	})

	// deploy-cmd
	flagset.Func("deploy-cmd", "command to run after the webhook is triggered", func(s string) error {
		cmds := strings.Split(s, " ")
		if len(cmds) < 1 {
			return errors.New("deploy-cmd was not set")
		}
		var (
			cmd     string = cmds[0]
			cmdPath string
			err     error
		)
		if cmdPath, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("command %q not found in path: %v", cmd, err)
		}

		result.DeployCommand = exec.Command(cmdPath, cmds[1:]...)
		slog.Debug("set deploy-cmd", slog.String("cmdPath", cmdPath), slog.Any("args", cmds[1:]))

		return nil
	})

	flagset.Parse(arguments)

	return result, nil
}

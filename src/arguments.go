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
	PreBuildCommand *string
	BuildCommand    *string
	DeployCommand   *string
	WorkDir         string
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
		slog.Debug("set workdir", slog.String("s", s))
		result.WorkDir = s
		return nil
	}

	setBuildCmd := func(s string) error {
		cmds := strings.Split(s, " ")
		if len(cmds) < 1 {
			return errors.New("build-cmd was not set")
		}
		var (
			cmd string = cmds[0]
			err error
		)
		if _, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("command %q not found in path: %v", cmd, err)
		}

		result.BuildCommand = &s
		slog.Debug("set build-cmd", slog.String("command", s))

		return nil
	}

	setDeployCmd := func(s string) error {
		cmds := strings.Split(s, " ")
		if len(cmds) < 1 {
			return errors.New("deploy-cmd was not set")
		}
		var (
			cmd string = cmds[0]
			err error
		)
		if _, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("command %q not found in path: %v", cmd, err)
		}

		result.DeployCommand = &s
		slog.Debug("set deploy-cmd", slog.String("command", s))

		return nil
	}

	setPreBuildCmd := func(s string) error {
		cmds := strings.Split(s, " ")
		if len(cmds) < 1 {
			return errors.New("pre-build-cmd was not set")
		}
		var (
			cmd string = cmds[0]
			err error
		)
		if _, err = exec.LookPath(cmd); err != nil {
			return fmt.Errorf("command %q not found in path: %v", cmd, err)
		}

		result.PreBuildCommand = &s
		slog.Debug("set pre-build-cmd", slog.String("command", s))

		return nil
	}

	defaultExecutable := "docker"
	defaultPreBuild := "git pull"
	defaultBuild := fmt.Sprintf("%s compose build", defaultExecutable)
	defaultDeploy := fmt.Sprintf("%s compose up -d --force-recreate", defaultExecutable)

	flagset.Func("workdir", "working directory where containers will be deployed from(default: current directory).", setWorkDir)
	flagset.Func("pre-build-cmd", fmt.Sprintf("command to run after the webhook is triggered(default: %q).", defaultPreBuild), setPreBuildCmd)
	flagset.Func("build-cmd", fmt.Sprintf("command to run after the pre-build command(default: %q).", defaultBuild), setBuildCmd)
	flagset.Func("deploy-cmd", fmt.Sprintf("command to run after the build is finished(default: %q).", defaultDeploy), setDeployCmd)

	flagset.Parse(arguments)

	// workdir default
	if result.WorkDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return result, err
		}
		err = setWorkDir(wd)
		if err != nil {
			return result, err
		}
	}

	// build-cmd default
	if result.BuildCommand == nil {
		err = setBuildCmd(defaultBuild)
		if err != nil {
			return result, fmt.Errorf("default executable %q for argument build-cmd not found, please set build-cmd yourself.", defaultExecutable)
		}
	}

	// deploy-cmd default
	if result.DeployCommand == nil {
		err = setDeployCmd(defaultDeploy)
		if err != nil {
			return result, fmt.Errorf("default executable %q for argument deploy-cmd not found, please set deploy-cmd yourself.", defaultExecutable)
		}
	}

	// pre-build-cmd default
	if result.PreBuildCommand == nil {
		err = setPreBuildCmd(defaultPreBuild)
		if err != nil {
			return result, fmt.Errorf("default executable %q for argument pre-build-cmd not found, please set pre-build-cmd yourself.", "git")
		}
	}

	return result, nil
}

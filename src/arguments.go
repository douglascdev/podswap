package podswap

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

type Arguments struct {
	YmlPath     string
	ProjectPath string
}

func ParseArguments(fs *flag.FlagSet, arguments []string) (result *Arguments, err error) {
	var (
		ymlPath     string
		projectPath string
	)

	setProjectPath := func(s string) error {
		stat, err := os.Stat(s)
		if err != nil {
			return err
		}

		if !stat.IsDir() {
			return fmt.Errorf("path %q is not a directory", s)
		}

		projectPath = s
		slog.Debug("set projectPath", slog.String("path", s))
		return nil
	}

	setYmlPath := func(s string) error {
		stat, err := os.Stat(s)
		if err != nil {
			return err
		}

		if stat.IsDir() {
			return fmt.Errorf("path %q is a directory, should be a file", s)
		}

		ymlPath = s
		slog.Debug("set ymlPath", slog.String("path", s))
		return nil
	}

	fs.Func("f", "root path for the project(default: working directory)", setProjectPath)
	fs.Func("a", "path for the yml action file(default: working directory/.github/workflows/podswap.yml)", setYmlPath)

	err = fs.Parse(arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// set default values
	if ymlPath == "" {
		slog.Info("ymlPath not set, using default value")

		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		err = setYmlPath(filepath.Join(wd, ".github", "workflows", "podswap.yml"))
		if err != nil {
			return nil, fmt.Errorf("failed to set ymlPath to default value: %w", err)
		}
	}

	if projectPath == "" {
		slog.Info("projectPath not set, using default value")
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}

		err = setProjectPath(wd)
		if err != nil {
			return nil, fmt.Errorf("failed to set projectPath to default value %q: %w", wd, err)
		}
	}

	return &Arguments{
		YmlPath:     ymlPath,
		ProjectPath: projectPath,
	}, nil
}

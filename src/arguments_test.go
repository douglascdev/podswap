package podswap_test

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	podswap "podswap/src"
	"slices"
	"testing"
)

func TestParseArguments(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	wd, err = filepath.Abs(wd)
	if err != nil {
		t.Fatal(err)
	}

	// add a made-up executable to path to later see if it's correctly detected
	dir, err := os.MkdirTemp("", "glorp")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create temp dir: %w", err))
	}
	tempFile, err := os.OpenFile(filepath.Join(dir, "glorp"), os.O_CREATE|os.O_RDWR, 0765)
	_, err = tempFile.Write([]byte("#!/bin/sh\necho 'Hello from glorp'"))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	defer os.RemoveAll(tempFile.Name())
	tempFile.Close()
	err = os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), dir))
	if err != nil {
		t.Fatal(fmt.Errorf("failed to set $PATH: %w", err))
	}

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		flagset         *flag.FlagSet
		arguments       []string
		wantPanic       bool
		wantedErr       error
		resultValidator func(*podswap.Arguments) error
	}{
		{
			"Can't set workdir to non-existent dir",
			flag.NewFlagSet("", flag.PanicOnError),
			[]string{"--workdir", filepath.Join(os.TempDir(), rand.Text())},
			true,
			nil,
			nil,
		},
		{
			"Can set workdir to existent dir",
			flag.NewFlagSet("", flag.PanicOnError),
			[]string{"--workdir", os.TempDir()},
			false,
			nil,
			func(a *podswap.Arguments) error {
				expected := os.TempDir()
				if a.WorkDir != expected {
					return fmt.Errorf("expected workdir to be %q, got %q", expected, a.WorkDir)
				}
				return nil
			},
		},
		{
			"Can set build-cmd and deploy-cmd",
			flag.NewFlagSet("", flag.PanicOnError),
			[]string{"--build-cmd", "glorp compose build", "--deploy-cmd", "glorp compose up -d"},
			false,
			nil,
			func(a *podswap.Arguments) error {
				// build-cmd
				if a.BuildCommand.Path != tempFile.Name() {
					return fmt.Errorf("expected build-cmd path to be %s, got %s", tempFile.Name(), a.BuildCommand.Path)
				}
				expectedBuildArgs := []string{"compose", "build"}
				got := a.BuildCommand.Args[1:]
				if !slices.Equal(got, expectedBuildArgs) {
					return fmt.Errorf("expected buildArgs to be %v, got %v", expectedBuildArgs, got)
				}

				// deploy-cmd
				if a.DeployCommand.Path != tempFile.Name() {
					return fmt.Errorf("expected deploy-cmd path to be %s, got %s", tempFile.Name(), a.DeployCommand.Path)
				}
				expectedDeployArgs := []string{"compose", "up", "-d"}
				got = a.DeployCommand.Args[1:]
				if !slices.Equal(got, expectedDeployArgs) {
					return fmt.Errorf("expected deployArgs to be %v, got %v", expectedDeployArgs, got)
				}

				return nil
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil && test.wantPanic {
					t.Errorf("expected panic for %q", test.name)
				} else if err != nil && !test.wantPanic {
					t.Errorf("unexpected panic for %q: %v", test.name, err)
				}
			}()

			result, gotErr := podswap.ParseArguments(test.flagset, test.arguments)

			if gotErr != test.wantedErr {
				t.Errorf("ParseArguments() failed: %v, expected error %v", gotErr, test.wantedErr)
			}

			if test.resultValidator != nil {
				if err = test.resultValidator(result); err != nil {
					t.Errorf("Unexpected result for test %q: %v", test.name, err)
				}
			}
		})
	}
}

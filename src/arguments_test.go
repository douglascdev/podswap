package podswap_test

import (
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	podswap "podswap/src"
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

	dir, err := os.MkdirTemp("", "glorp")
	if err != nil {
		t.Fatal(fmt.Errorf("failed to create temp dir: %w", err))
	}
	defer os.RemoveAll(dir)

	// add a made-up executable to path to later see if it's correctly detected
	glorpTempFile, err := os.OpenFile(filepath.Join(dir, "glorp"), os.O_CREATE|os.O_RDWR, 0765)
	_, err = glorpTempFile.Write([]byte("#!/bin/sh\necho 'Hello from glorp'"))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	glorpTempFile.Close()
	err = os.Setenv("PATH", fmt.Sprintf("%s:%s", os.Getenv("PATH"), dir))
	if err != nil {
		t.Fatal(fmt.Errorf("failed to set $PATH: %w", err))
	}

	// add a fake docker executable to path to later see if it's correctly detected
	dockerTempFile, err := os.OpenFile(filepath.Join(dir, "docker"), os.O_CREATE|os.O_RDWR, 0765)
	_, err = dockerTempFile.Write([]byte("#!/bin/sh\necho 'Hello from docker'"))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	dockerTempFile.Close()
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
				expected := "glorp compose build"
				if *a.BuildCommand != expected {
					return fmt.Errorf("expected build-cmd path to be %s, got %s", expected, *a.BuildCommand)
				}

				expected = "glorp compose up -d"
				if *a.DeployCommand != expected {
					return fmt.Errorf("expected deploy-cmd path to be %s, got %s", expected, *a.DeployCommand)
				}

				return nil
			},
		},
		{
			"Check if it runs with default build-cmd and deploy-cmd",
			flag.NewFlagSet("", flag.PanicOnError),
			[]string{},
			false,
			nil,
			func(a *podswap.Arguments) error {
				// build-cmd
				expected := "docker compose build"
				if *a.BuildCommand != expected {
					return fmt.Errorf("expected build-cmd path to be %s, got %s", expected, *a.BuildCommand)
				}

				expected = "docker compose up -d --force-recreate"
				if *a.DeployCommand != expected {
					return fmt.Errorf("expected deploy-cmd path to be %s, got %s", expected, *a.DeployCommand)
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

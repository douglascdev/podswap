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
			"Can set port and host",
			flag.NewFlagSet("", flag.PanicOnError),
			[]string{"--port", "123", "--host", "abc"},
			false,
			nil,
			func(a *podswap.Arguments) error {
				var expectedPort uint = 123
				if *a.Port != expectedPort {
					return fmt.Errorf("expected port to be %d, got %d", expectedPort, a.Port)
				}

				expectedHost := "abc"
				if *a.Host != expectedHost {
					return fmt.Errorf("expected host to be %q, got %q", expectedHost, *a.Host)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				err := recover()
				if err == nil && tt.wantPanic {
					t.Errorf("expected panic for %q", tt.name)
				}
			}()
			result, gotErr := podswap.ParseArguments(tt.flagset, tt.arguments)

			if gotErr != tt.wantedErr {
				t.Errorf("ParseArguments() failed: %v, expected error %v", gotErr, tt.wantedErr)
			}

			if tt.resultValidator != nil {
				if err = tt.resultValidator(result); err != nil {
					t.Errorf("Unexpected result for test %q: %v", tt.name, err)
				}
			}
		})
	}
}

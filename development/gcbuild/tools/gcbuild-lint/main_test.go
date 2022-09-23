package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func Test_Flags(t *testing.T) {
	tc := []struct {
		name       string
		flags      []string
		expectOpts options
		expectErr  bool
	}{
		{
			name:       "flags parsed",
			expectOpts: options{baseSha: "b830e1032526165afd0e74fdaad8cca8d752cc2c", fromGit: true},
			expectErr:  false,
			flags: []string{
				"--base-sha=b830e1032526165afd0e74fdaad8cca8d752cc2c",
				"--from-git",
			},
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			fs := flag.NewFlagSet("flags", flag.ContinueOnError)
			o := options{}
			o.gatherOptions(fs)
			err := fs.Parse(c.flags)
			if err != nil && !c.expectErr {
				t.Errorf("got error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't get error, but wanted to")
			}
			if !reflect.DeepEqual(o, c.expectOpts) {
				t.Errorf("%v != %v", o, c.expectOpts)
			}
		})
	}
}

func Test_baseRef(t *testing.T) {
	tc := []struct {
		name          string
		env           map[string]string
		o             options
		expectErr     bool
		expectBaseSha string
	}{
		{
			name:      "in CI, baseRef not provided",
			o:         options{isCI: false},
			expectErr: true,
		},
		{
			name:          "in CI, baseSha provided through env variable",
			env:           map[string]string{"PULL_BASE_SHA": "b830e1032526165afd0e74fdaad8cca8d752cc2c"},
			o:             options{isCI: true},
			expectBaseSha: "b830e1032526165afd0e74fdaad8cca8d752cc2c",
			expectErr:     false,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			if len(c.env) > 0 {
				for k, v := range c.env {
					t.Setenv(k, v)
				}
			}
			b, err := baseRef(c.o)
			if err != nil && !c.expectErr {
				t.Errorf("got error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't get error, but wanted to")
			}
			if !reflect.DeepEqual(b, c.expectBaseSha) {
				t.Errorf("%v != %v", b, c.expectBaseSha)
			}
		})
	}
}

func TestGitProcess(t *testing.T) {
	if os.Getenv("GO_TEST_PROCESS") != "1" {
		return
	}
	fmt.Fprintf(os.Stdout, "test/cloudbuild.yaml\ngotest/file.go")
	os.Exit(0)
}

func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestGitProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_TEST_PROCESS=1"}
	return cmd
}

func Test_getFiles(t *testing.T) {
	execCmd = fakeExecCommand
	// this is not thread safe approach. needs further investigation
	defer func() { execCmd = exec.Command }()

	tc := []struct {
		name        string
		o           options
		expectErr   bool
		expectFiles []string
	}{
		{
			name:        "files from git, pass",
			o:           options{fromGit: true, baseSha: "main"},
			expectErr:   false,
			expectFiles: []string{"test/cloudbuild.yaml"},
		},
		{
			name:        "files from args, pass",
			o:           options{},
			expectErr:   false,
			expectFiles: []string{"test/cloudbuild.yaml"},
		},
	}
	for _, c := range tc {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		args := []string{
			"test/cloudbuild.yaml",
			"gotest/file.go",
		}
		fs.Parse(args)
		t.Run(c.name, func(t *testing.T) {
			f, err := getFiles(c.o, fs)
			if err != nil && !c.expectErr {
				t.Errorf("got error, but didn't want to: %v", err)
			}
			if err == nil && c.expectErr {
				t.Errorf("didn't get error, but wanted to")
			}
			if !reflect.DeepEqual(f, c.expectFiles) {
				t.Errorf("%v != %v", f, c.expectFiles)
			}
		})
	}
}

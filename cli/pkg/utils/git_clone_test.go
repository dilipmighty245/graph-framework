package utils

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func fakeModuleGo(t *testing.T) string {
	t.Helper()
	bin := t.TempDir()
	tool := filepath.Join(bin, "go")
	if err := os.WriteFile(tool, []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$TEST_GO_ARGS\"\nexit ${TEST_GO_EXIT:-0}\n"), 0755); err != nil {
		t.Fatal(err)
	}
	args := filepath.Join(bin, "args")
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TEST_GO_ARGS", args)
	return args
}

func TestGoModInitPassesModulePathAsOneArgument(t *testing.T) {
	argsFile := fakeModuleGo(t)
	modulePath := "example.com/model; touch " + filepath.Join(t.TempDir(), "injected")
	if err := GoModInit(modulePath, true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n"); !reflect.DeepEqual(got, []string{"mod", "init", modulePath}) {
		t.Fatalf("unexpected Go arguments: %v", got)
	}
}

func TestGoModInitReportsInitializationFailure(t *testing.T) {
	fakeModuleGo(t)
	t.Setenv("TEST_GO_EXIT", "1")
	if err := GoModInit("example.com/model", true); err == nil {
		t.Fatal("expected go mod init failure to be returned")
	}
}

func TestGoModInitUsesGoInferenceForEmptyName(t *testing.T) {
	argsFile := fakeModuleGo(t)
	if err := GoModInit("", true); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "mod\ninit\n" {
		t.Fatalf("unexpected Go arguments: %s", data)
	}
}

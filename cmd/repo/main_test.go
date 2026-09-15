package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cao7113/repo-manager/internal/config"
)

func TestRootCommandReportsVersion(t *testing.T) {
	command := newRootCommand()
	if command.Version != version {
		t.Fatalf("version = %q, want %q", command.Version, version)
	}
}

func TestVersionCommand(t *testing.T) {
	command := newRootCommand()
	versionCommand, _, err := command.Find([]string{"v"})
	if err != nil {
		t.Fatal(err)
	}
	if versionCommand.Name() != "version" {
		t.Fatalf("command name = %q, want %q", versionCommand.Name(), "version")
	}

	var output bytes.Buffer
	versionCommand.SetOut(&output)
	if err := versionCommand.RunE(versionCommand, nil); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(output.String()); got != version {
		t.Fatalf("version output = %q, want %q", got, version)
	}
}

func TestListCommandPrintsAlignedColumns(t *testing.T) {
	state := &app{file: config.File{Repos: []config.RepoItem{
		{Name: "one", Path: "/tmp/one", URL: "https://github.com/org/one.git"},
		{Name: "longer-name", Path: "/tmp/a/longer/path", URL: "git@github.com:org/two.git"},
	}}}
	command := newListCommand(state)
	var output bytes.Buffer
	command.SetOut(&output)

	if err := command.RunE(command, nil); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d output lines, want 3: %q", len(lines), output.String())
	}
	if lines[0] != "name         path                url" {
		t.Errorf("header = %q", lines[0])
	}
	if lines[1] != "one          /tmp/one            github:org/one" {
		t.Errorf("first row = %q", lines[1])
	}
	if lines[2] != "longer-name  /tmp/a/longer/path  github:org/two" {
		t.Errorf("second row = %q", lines[2])
	}
}

func TestShortURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{name: "github https", url: "https://github.com/org/repo.git", want: "github:org/repo"},
		{name: "github ssh", url: "git@github.com:org/repo.git", want: "github:org/repo"},
		{name: "other host", url: "https://gitlab.com/org/repo.git", want: "https://gitlab.com/org/repo.git"},
		{name: "missing url", url: "", want: "-"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := shortURL(test.url); got != test.want {
				t.Fatalf("shortURL(%q) = %q, want %q", test.url, got, test.want)
			}
		})
	}
}

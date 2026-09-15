package repository

import (
	"testing"

	"github.com/cao7113/repo-manager/internal/config"
)

func TestAddUpdatesURLForExistingPath(t *testing.T) {
	file := config.File{Repos: []config.RepoItem{{Name: "one", Path: "/tmp/one"}}}

	if err := Add(&file, config.RepoItem{Name: "one", Path: "/tmp/one", URL: "https://example.com/one.git"}); err != nil {
		t.Fatalf("update existing path: %v", err)
	}
	if len(file.Repos) != 1 || file.Repos[0].URL != "https://example.com/one.git" {
		t.Fatalf("repos after update = %#v", file.Repos)
	}

	if err := Add(&file, config.RepoItem{Name: "ONE", Path: "/tmp/two"}); err == nil {
		t.Fatal("expected duplicate name error")
	}
}

func TestSearchMatchesNameAndPath(t *testing.T) {
	repos := []config.RepoItem{
		{Name: "repo-manager", Path: "/work/go/repo-manager"},
		{Name: "other", Path: "/work/other"},
	}

	if got := Search(repos, "MANAGER"); len(got) != 1 || got[0].Name != "repo-manager" {
		t.Fatalf("search by name = %#v", got)
	}
	if got := Search(repos, "/work/other"); len(got) != 1 || got[0].Name != "other" {
		t.Fatalf("search by path = %#v", got)
	}
}

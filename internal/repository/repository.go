package repository

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/username/repo-manager/internal/config"
)

func Add(file *config.File, item config.RepoItem) error {
	for _, existing := range file.Repos {
		if existing.Path == item.Path {
			return fmt.Errorf("repository already tracked: %s", item.Path)
		}
		if strings.EqualFold(existing.Name, item.Name) {
			return fmt.Errorf("repository name already tracked: %s", item.Name)
		}
	}
	file.Repos = append(file.Repos, item)
	return nil
}

func Find(repos []config.RepoItem, query string) (config.RepoItem, error) {
	absolute, _ := filepath.Abs(query)
	for _, item := range repos {
		if item.Name == query || item.Path == absolute || item.Path == query {
			return item, nil
		}
	}
	return config.RepoItem{}, fmt.Errorf("repository not found: %s", query)
}

func Search(repos []config.RepoItem, query string) []config.RepoItem {
	if query == "" {
		return repos
	}
	query = strings.ToLower(query)
	result := make([]config.RepoItem, 0)
	for _, item := range repos {
		if strings.Contains(strings.ToLower(item.Name), query) || strings.Contains(strings.ToLower(item.Path), query) {
			result = append(result, item)
		}
	}
	return result
}

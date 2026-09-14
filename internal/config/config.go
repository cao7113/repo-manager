package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const fileName = "repos.yaml"

type RepoItem struct {
	Name string   `yaml:"name"`
	Path string   `yaml:"path"`
	URL  string   `yaml:"url,omitempty"`
	Desc string   `yaml:"desc,omitempty"`
	Tags []string `yaml:"tags,omitempty"`
}

type File struct {
	Version int        `yaml:"version"`
	Repos   []RepoItem `yaml:"repos"`
}

func Path() (string, error) {
	if value := os.Getenv("REPO_MANAGER_CONFIG"); value != "" {
		return expandPath(value)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "repo-manager", fileName), nil
}

func Load(path string) (File, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return File{Version: 1, Repos: []RepoItem{}}, nil
	}
	if err != nil {
		return File{}, err
	}
	var file File
	if err := yaml.Unmarshal(data, &file); err != nil {
		return File{}, fmt.Errorf("parse config: %w", err)
	}
	if file.Version == 0 {
		file.Version = 1
	}
	return file, nil
}

func Save(path string, file File) error {
	if file.Version == 0 {
		file.Version = 1
	}
	data, err := yaml.Marshal(file)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".repos-*.yaml")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func ExpandPath(path string) (string, error) {
	return expandPath(path)
}

func expandPath(path string) (string, error) {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolute), nil
}

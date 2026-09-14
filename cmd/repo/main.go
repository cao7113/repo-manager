package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/username/repo-manager/internal/config"
	"github.com/username/repo-manager/internal/git"
	"github.com/username/repo-manager/internal/repository"
)

type app struct {
	configPath string
	file       config.File
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "repo:", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	state := &app{}
	root := &cobra.Command{
		Use:   "repo",
		Short: "Manage local Git repositories",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			if state.configPath != "" {
				path, err = config.ExpandPath(state.configPath)
				if err != nil {
					return err
				}
			}
			state.configPath = path
			state.file, err = config.Load(path)
			return err
		},
	}
	root.PersistentFlags().StringVar(&state.configPath, "config", "", "config file path")
	root.AddCommand(newAddCommand(state), newCloneCommand(state), newListCommand(state), newShowCommand(state), newStatCommand(state))
	return root
}

func newAddCommand(state *app) *cobra.Command {
	return &cobra.Command{
		Use:   "add PATH",
		Short: "Track a local Git repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return add(cmd.Context(), state.configPath, &state.file, args[0])
		},
	}
}

func newCloneCommand(state *app) *cobra.Command {
	command := &cobra.Command{Use: "clone URL [PATH]", Short: "Clone a repository and track it"}
	command.Args = func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "all" {
			return nil
		}
		return cobra.RangeArgs(1, 2)(cmd, args)
	}
	command.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && args[0] == "all" {
			return cloneAll(cmd.Context(), state.configPath, &state.file)
		}
		return cloneOne(cmd.Context(), state.configPath, &state.file, args)
	}
	return command
}

func newListCommand(state *app) *cobra.Command {
	return &cobra.Command{
		Use:     "ls [QUERY]",
		Aliases: []string{"l"},
		Short:   "List tracked repositories",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) == 1 {
				query = args[0]
			}
			writer := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(writer, "name\tpath\turl")
			for _, item := range repository.Search(state.file.Repos, query) {
				fmt.Fprintf(writer, "%s\t%s\t%s\n", item.Name, item.Path, shortURL(item.URL))
			}
			return writer.Flush()
		},
	}
}

func shortURL(raw string) string {
	if raw == "" {
		return "-"
	}

	if strings.HasPrefix(raw, "git@github.com:") {
		return "github:" + strings.TrimSuffix(strings.TrimPrefix(raw, "git@github.com:"), ".git")
	}

	parsed, err := url.Parse(raw)
	if err == nil && parsed.Host == "github.com" {
		return "github:" + strings.TrimSuffix(strings.TrimPrefix(parsed.Path, "/"), ".git")
	}
	return raw
}

func newShowCommand(state *app) *cobra.Command {
	return &cobra.Command{
		Use: "show NAME_OR_PATH", Short: "Show repository details", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			item, err := repository.Find(state.file.Repos, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\nPath: %s\nURL: %s\nDescription: %s\nTags: %s\n", item.Name, item.Path, item.URL, item.Desc, strings.Join(item.Tags, ", "))
			status, err := git.StatusOf(cmd.Context(), item.Path)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Branch: %s\nStatus: %s\n", status.Branch, statusLabel(status))
			return nil
		},
	}
}

func newStatCommand(state *app) *cobra.Command {
	return &cobra.Command{
		Use: "stat", Short: "Show status for all tracked repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, item := range state.file.Repos {
				status, err := git.StatusOf(cmd.Context(), item.Path)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%-20s missing/invalid\n", item.Name)
					continue
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-20s %-12s %s\n", item.Name, status.Branch, statusLabel(status))
			}
			return nil
		},
	}
}

func add(ctx context.Context, path string, file *config.File, input string) error {
	input, err := config.ExpandPath(input)
	if err != nil {
		return err
	}
	root, err := git.Root(ctx, input)
	if err != nil {
		return fmt.Errorf("not a git repository: %s", input)
	}
	root, err = config.ExpandPath(root)
	if err != nil {
		return err
	}
	url, err := git.RemoteURL(ctx, root)
	if err != nil {
		return err
	}
	item := config.RepoItem{Name: filepath.Base(root), Path: root, URL: url}
	if err := repository.Add(file, item); err != nil {
		return err
	}
	if err := config.Save(path, *file); err != nil {
		return err
	}
	fmt.Printf("added %s\n", item.Name)
	return nil
}

func cloneOne(ctx context.Context, path string, file *config.File, args []string) error {
	if err := git.Clone(ctx, args...); err != nil {
		return err
	}
	clonePath := ""
	if len(args) == 2 {
		clonePath, _ = config.ExpandPath(args[1])
	} else {
		clonePath = filepath.Join(mustWorkingDirectory(), filepath.Base(strings.TrimSuffix(args[0], ".git")))
	}
	return add(ctx, path, file, clonePath)
}

func cloneAll(ctx context.Context, path string, file *config.File) error {
	for _, item := range file.Repos {
		if _, err := os.Stat(item.Path); err == nil {
			continue
		}
		if item.URL == "" {
			fmt.Printf("skip %s: no remote URL\n", item.Name)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(item.Path), 0o755); err != nil {
			return err
		}
		fmt.Printf("cloning %s\n", item.URL)
		if err := git.Clone(ctx, item.URL, item.Path); err != nil {
			return err
		}
	}
	return nil
}

func statusLabel(status git.Status) string {
	if status.Conflicts {
		return "conflicted"
	}
	if status.Dirty {
		if status.Untracked {
			return "modified, untracked"
		}
		return "modified"
	}
	if status.Ahead > 0 && status.Behind > 0 {
		return "diverged"
	}
	if status.Ahead > 0 {
		return "ahead"
	}
	if status.Behind > 0 {
		return "behind"
	}
	return "clean"
}

func mustWorkingDirectory() string {
	path, err := os.Getwd()
	if err != nil {
		return "."
	}
	return path
}

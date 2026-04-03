// Package config handles loading and saving the wtree configuration file at
// ~/.config/wtree/config.yml.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// HooksConfig holds paths to executable hook scripts for a project.
// post_navigation receives: <ref> <project_name> <worktree_dir>
// post_delete receives:     <ref>
type HooksConfig struct {
	PostNavigation string `yaml:"post_navigation,omitempty"`
	PostDelete     string `yaml:"post_delete,omitempty"`
}

// ExpandedPostNavigation returns the PostNavigation path with ~ expanded.
func (h HooksConfig) ExpandedPostNavigation() string { return expandHome(h.PostNavigation) }

// ExpandedPostDelete returns the PostDelete path with ~ expanded.
func (h HooksConfig) ExpandedPostDelete() string { return expandHome(h.PostDelete) }

// Project holds per-project settings from the config file.
type Project struct {
	Name         string      `yaml:"name"`
	Path         string      `yaml:"path"`
	TicketPrefix string      `yaml:"ticket_prefix,omitempty"`
	BranchPrefix string      `yaml:"branch_prefix,omitempty"`
	Hooks        HooksConfig `yaml:"hooks,omitempty"`
}

// Config is the top-level structure of ~/.config/wtree/config.yml.
type Config struct {
	BaseDir  string    `yaml:"base_dir,omitempty"`
	Projects []Project `yaml:"projects,omitempty"`
}

// Load reads the config file. Returns an empty Config if the file does not
// exist yet.
func Load() (Config, error) {
	data, err := os.ReadFile(FilePath())
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save writes the config back to disk, creating parent directories as needed.
func (c *Config) Save() error {
	p := FilePath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// FindOrCreate returns the Project matching projectDir (by path, then by name).
// If no match is found it appends a new entry and returns it with created=true
// — the caller should call Save().
func (c *Config) FindOrCreate(name, projectDir string) (Project, bool) {
	abs := expandHome(projectDir)

	for _, p := range c.Projects {
		if expandHome(p.Path) == abs {
			return p, false
		}
	}
	for _, p := range c.Projects {
		if p.Name == name {
			return p, false
		}
	}

	p := Project{Name: name, Path: projectDir}
	c.Projects = append(c.Projects, p)
	return p, true
}

// ExpandedBaseDir returns BaseDir with ~ expanded.
func (c *Config) ExpandedBaseDir() string {
	return expandHome(c.BaseDir)
}

// FilePath returns the path to the config file. It is a variable so tests can
// override it without touching the real ~/.config/wtree/config.yml.
var FilePath = func() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "wtree", "config.yml")
}

func (c *Config) validate() error {
	names := make(map[string]struct{}, len(c.Projects))
	paths := make(map[string]struct{}, len(c.Projects))
	for _, p := range c.Projects {
		if _, dup := names[p.Name]; dup {
			return fmt.Errorf("config: duplicate project name %q", p.Name)
		}
		names[p.Name] = struct{}{}

		abs := expandHome(p.Path)
		if _, dup := paths[abs]; dup {
			return fmt.Errorf("config: duplicate project path %q", p.Path)
		}
		paths[abs] = struct{}{}
	}
	return nil
}

func expandHome(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}

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

// ProjectConfig holds per-project settings from the config file and runtime hooks.
type ProjectConfig struct {
	Name         string      `yaml:"name"`
	Path         string      `yaml:"path"`
	BaseDir      string      `yaml:"base_dir,omitempty"`
	TicketPrefix string      `yaml:"ticket_prefix,omitempty"`
	BranchPrefix string      `yaml:"branch_prefix,omitempty"`
	Hooks        HooksConfig `yaml:"hooks,omitempty"`

	PostNavigation func(ref, projectName, worktreeDir string) error `yaml:"-"`
	PostDelete     func(ref string) error                           `yaml:"-"`
}

// ExpandedBaseDir returns BaseDir with ~ expanded.
func (p ProjectConfig) ExpandedBaseDir() string { return expandHome(p.BaseDir) }

// EffectiveBaseDir returns the configured base dir with ~ expanded, or the
// default project-local base dir when BaseDir is empty.
func (p ProjectConfig) EffectiveBaseDir() string {
	if p.BaseDir == "" {
		return filepath.Join(expandHome(p.Path), ".wtree")
	}
	return p.ExpandedBaseDir()
}

// Config is the top-level structure of ~/.config/wtree/config.yml.
type Config struct {
	Projects []ProjectConfig `yaml:"projects,omitempty"`
}

// DefaultProjectConfig returns a ProjectConfig with sensible defaults.
func DefaultProjectConfig(name, projectDir string) ProjectConfig {
	return ProjectConfig{
		Name:         name,
		Path:         projectDir,
		BaseDir:      filepath.Join(projectDir, ".wtree"),
		TicketPrefix: "ABC-",
		BranchPrefix: "ob-",
	}
}

// Load reads the config file. Returns an empty Config if the file does not
// exist yet.
func Load() (Config, error) {
	data, err := os.ReadFile(ConfigPath())
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
	p := ConfigPath()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// FindProjectConfig returns the project config matching projectDir (by path, then by name).
func (c *Config) FindProjectConfig(name, projectDir string) (ProjectConfig, bool) {
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
	return ProjectConfig{}, true
}

// FindProjectConfig reads the config file and returns the matching project config.
func FindProjectConfig(name, projectDir string) (ProjectConfig, bool, error) {
	cfg, err := Load()
	if err != nil {
		return ProjectConfig{}, false, err
	}
	projectCfg, missing := cfg.FindProjectConfig(name, projectDir)
	return projectCfg, !missing, nil
}

// LoadOrCreateProjectConfig returns the matching project config from disk,
// or creates, saves, and returns a default one when no entry exists yet.
func LoadOrCreateProjectConfig(name, projectDir string) (ProjectConfig, error) {
	cfg, err := Load()
	if err != nil {
		return ProjectConfig{}, err
	}

	projectCfg, missing := cfg.FindProjectConfig(name, projectDir)
	if !missing {
		return projectCfg, nil
	}

	projectCfg = DefaultProjectConfig(name, projectDir)
	cfg.Projects = append(cfg.Projects, projectCfg)
	if err := cfg.Save(); err != nil {
		return ProjectConfig{}, err
	}
	return projectCfg, nil
}

// ConfigPath returns the path to the config file. It is a variable so tests can
// override it without touching the real ~/.config/wtree/config.yml.
var ConfigPath = func() string {
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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	cases := []struct{ in, want string }{
		{"~/foo/bar", filepath.Join(home, "foo/bar")},
		{"/abs/path", "/abs/path"},
		{"relative", "relative"},
		{"~notexpanded", "~notexpanded"},
	}
	for _, c := range cases {
		if got := expandHome(c.in); got != c.want {
			t.Errorf("expandHome(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestValidate_DuplicateName(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/a"},
		{Name: "foo", Path: "/b"},
	}}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for duplicate name, got nil")
	} else if !strings.Contains(err.Error(), "foo") {
		t.Errorf("error should mention the duplicate name, got %q", err.Error())
	}
}

func TestValidate_DuplicatePath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/same"},
		{Name: "bar", Path: "/same"},
	}}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error for duplicate path, got nil")
	} else if !strings.Contains(err.Error(), "/same") {
		t.Errorf("error should mention the duplicate path, got %q", err.Error())
	}
}

func TestValidate_DuplicatePathExpandedTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "~/code"},
		{Name: "bar", Path: filepath.Join(home, "code")},
	}}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error when ~ and absolute path resolve to the same dir")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/a"},
		{Name: "bar", Path: "/b"},
	}}
	if err := cfg.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindProjectConfigByPath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/code/foo", TicketPrefix: "ABC-"},
	}}
	p, missing := cfg.FindProjectConfig("foo", "/code/foo")
	if missing {
		t.Error("expected missing=false for existing path")
	}
	if p.TicketPrefix != "ABC-" {
		t.Errorf("expected to return existing project, got %+v", p)
	}
}

func TestFindProjectConfigByName(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/old/path", TicketPrefix: "XYZ-"},
	}}
	p, missing := cfg.FindProjectConfig("foo", "/new/path")
	if missing {
		t.Error("expected missing=false for existing name")
	}
	if p.TicketPrefix != "XYZ-" {
		t.Errorf("expected to return existing project, got %+v", p)
	}
}

func TestFindProjectConfigMissing(t *testing.T) {
	cfg := Config{}
	p, missing := cfg.FindProjectConfig("newproject", "/code/newproject")
	if !missing {
		t.Error("expected missing=true for new project")
	}
	if p.Name != "" || p.Path != "" || p.BaseDir != "" || p.TicketPrefix != "" || p.BranchPrefix != "" {
		t.Errorf("expected empty project config, got %+v", p)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "config.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	cfg := Config{
		Projects: []ProjectConfig{
			{Name: "myproject", Path: "/code/myproject", BaseDir: "~/.wtree/myproject", TicketPrefix: "ABC-"},
		},
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(loaded.Projects))
	}
	if loaded.Projects[0].Name != "myproject" {
		t.Errorf("project name = %q, want %q", loaded.Projects[0].Name, "myproject")
	}
	if loaded.Projects[0].TicketPrefix != "ABC-" {
		t.Errorf("ticket_prefix = %q, want %q", loaded.Projects[0].TicketPrefix, "ABC-")
	}
	if loaded.Projects[0].BaseDir != "~/.wtree/myproject" {
		t.Errorf("base_dir = %q, want %q", loaded.Projects[0].BaseDir, "~/.wtree/myproject")
	}
}

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "nonexistent.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should return empty config when file missing, got error: %v", err)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

func TestLoadOrCreateProjectConfigCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "config.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	got, err := LoadOrCreateProjectConfig("myproject", "/code/myproject")
	if err != nil {
		t.Fatalf("LoadOrCreateProjectConfig() error: %v", err)
	}
	if got.Name != "myproject" || got.Path != "/code/myproject" {
		t.Fatalf("unexpected project config: %+v", got)
	}
	if got.BaseDir != "/code/myproject/.wtree" {
		t.Fatalf("BaseDir = %q, want %q", got.BaseDir, "/code/myproject/.wtree")
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(loaded.Projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(loaded.Projects))
	}
}

func TestProjectConfigEffectiveBaseDirUsesDefaultWhenEmpty(t *testing.T) {
	cfg := ProjectConfig{Path: "/code/myproject"}

	if got := cfg.EffectiveBaseDir(); got != "/code/myproject/.wtree" {
		t.Fatalf("EffectiveBaseDir() = %q, want %q", got, "/code/myproject/.wtree")
	}
}

func TestProjectConfigEffectiveBaseDirExpandsConfiguredValue(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := ProjectConfig{Path: "/code/myproject", BaseDir: "~/.wtree/myproject"}

	if got := cfg.EffectiveBaseDir(); got != filepath.Join(home, ".wtree", "myproject") {
		t.Fatalf("EffectiveBaseDir() = %q, want %q", got, filepath.Join(home, ".wtree", "myproject"))
	}
}

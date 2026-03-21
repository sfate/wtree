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
	cfg := Config{Projects: []Project{
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
	cfg := Config{Projects: []Project{
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
	cfg := Config{Projects: []Project{
		{Name: "foo", Path: "~/code"},
		{Name: "bar", Path: filepath.Join(home, "code")},
	}}
	if err := cfg.validate(); err == nil {
		t.Fatal("expected error when ~ and absolute path resolve to the same dir")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{Projects: []Project{
		{Name: "foo", Path: "/a"},
		{Name: "bar", Path: "/b"},
	}}
	if err := cfg.validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFindOrCreate_ByPath(t *testing.T) {
	cfg := Config{Projects: []Project{
		{Name: "foo", Path: "/code/foo", TicketPrefix: "ABC-"},
	}}
	p, created := cfg.FindOrCreate("foo", "/code/foo")
	if created {
		t.Error("expected created=false for existing path")
	}
	if p.TicketPrefix != "ABC-" {
		t.Errorf("expected to return existing project, got %+v", p)
	}
}

func TestFindOrCreate_ByName(t *testing.T) {
	cfg := Config{Projects: []Project{
		{Name: "foo", Path: "/old/path", TicketPrefix: "XYZ-"},
	}}
	p, created := cfg.FindOrCreate("foo", "/new/path")
	if created {
		t.Error("expected created=false for existing name")
	}
	if p.TicketPrefix != "XYZ-" {
		t.Errorf("expected to return existing project, got %+v", p)
	}
}

func TestFindOrCreate_NewEntry(t *testing.T) {
	cfg := Config{}
	p, created := cfg.FindOrCreate("newproject", "/code/newproject")
	if !created {
		t.Error("expected created=true for new project")
	}
	if p.Name != "newproject" || p.Path != "/code/newproject" {
		t.Errorf("unexpected project: %+v", p)
	}
	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project in config, got %d", len(cfg.Projects))
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	origFilePath := FilePath
	FilePath = func() string { return filepath.Join(dir, "config.yml") }
	t.Cleanup(func() { FilePath = origFilePath })

	cfg := Config{
		BaseDir: "~/.worktrees",
		Projects: []Project{
			{Name: "myproject", Path: "/code/myproject", TicketPrefix: "ABC-"},
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
}

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	origFilePath := FilePath
	FilePath = func() string { return filepath.Join(dir, "nonexistent.yml") }
	t.Cleanup(func() { FilePath = origFilePath })

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should return empty config when file missing, got error: %v", err)
	}
	if len(cfg.Projects) != 0 {
		t.Errorf("expected empty config, got %+v", cfg)
	}
}

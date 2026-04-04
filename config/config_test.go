package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
		require.Equal(t, c.want, expandHome(c.in))
	}
}

func TestValidate_DuplicateName(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/a"},
		{Name: "foo", Path: "/b"},
	}}
	err := cfg.validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "foo")
}

func TestValidate_DuplicatePath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/same"},
		{Name: "bar", Path: "/same"},
	}}
	err := cfg.validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "/same")
}

func TestValidate_DuplicatePathExpandedTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "~/code"},
		{Name: "bar", Path: filepath.Join(home, "code")},
	}}
	require.Error(t, cfg.validate())
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/a"},
		{Name: "bar", Path: "/b"},
	}}
	require.NoError(t, cfg.validate())
}

func TestFindProjectConfigByPath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/code/foo", TicketPrefix: "ABC-"},
	}}
	p, missing := cfg.FindProjectConfig("foo", "/code/foo")
	require.False(t, missing)
	require.Equal(t, "ABC-", p.TicketPrefix)
}

func TestFindProjectConfigByName(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Name: "foo", Path: "/old/path", TicketPrefix: "XYZ-"},
	}}
	p, missing := cfg.FindProjectConfig("foo", "/new/path")
	require.False(t, missing)
	require.Equal(t, "XYZ-", p.TicketPrefix)
}

func TestFindProjectConfigMissing(t *testing.T) {
	cfg := Config{}
	p, missing := cfg.FindProjectConfig("newproject", "/code/newproject")
	require.True(t, missing)
	require.Equal(t, ProjectConfig{}, ProjectConfig{
		Name:         p.Name,
		Path:         p.Path,
		BaseDir:      p.BaseDir,
		TicketPrefix: p.TicketPrefix,
		BranchPrefix: p.BranchPrefix,
	})
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
	require.NoError(t, cfg.Save())

	loaded, err := Load()
	require.NoError(t, err)
	require.Len(t, loaded.Projects, 1)
	require.Equal(t, "myproject", loaded.Projects[0].Name)
	require.Equal(t, "ABC-", loaded.Projects[0].TicketPrefix)
	require.Equal(t, "~/.wtree/myproject", loaded.Projects[0].BaseDir)
}

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "nonexistent.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	cfg, err := Load()
	require.NoError(t, err)
	require.Len(t, cfg.Projects, 0)
}

func TestLoadOrCreateProjectConfigCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "config.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	got, err := LoadOrCreateProjectConfig("myproject", "/code/myproject")
	require.NoError(t, err)
	require.Equal(t, "myproject", got.Name)
	require.Equal(t, "/code/myproject", got.Path)
	require.Equal(t, "/code/myproject/.wtree", got.BaseDir)

	loaded, err := Load()
	require.NoError(t, err)
	require.Len(t, loaded.Projects, 1)
}

func TestProjectConfigEffectiveBaseDirUsesDefaultWhenEmpty(t *testing.T) {
	cfg := ProjectConfig{Path: "/code/myproject"}

	require.Equal(t, "/code/myproject/.wtree", cfg.EffectiveBaseDir())
}

func TestProjectConfigEffectiveBaseDirExpandsConfiguredValue(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := ProjectConfig{Path: "/code/myproject", BaseDir: "~/.wtree/myproject"}

	require.Equal(t, filepath.Join(home, ".wtree", "myproject"), cfg.EffectiveBaseDir())
}

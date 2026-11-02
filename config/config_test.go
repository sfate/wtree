package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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

func TestValidate_DuplicatePath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Path: "/same"},
		{Path: "/same"},
	}}
	err := cfg.validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "/same")
}

func TestValidate_DuplicatePathExpandedTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	cfg := Config{Projects: []ProjectConfig{
		{Path: "~/code"},
		{Path: filepath.Join(home, "code")},
	}}
	require.Error(t, cfg.validate())
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Path: "/a"},
		{Path: "/b"},
	}}
	require.NoError(t, cfg.validate())
}

func TestFindProjectConfigByPath(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Path: "/code/foo", TicketPrefix: TicketPrefixes{"AB-"}},
	}}
	p, missing := cfg.FindProjectConfig("/code/foo")
	require.False(t, missing)
	require.Equal(t, TicketPrefixes{"AB-"}, p.TicketPrefix)
}

func TestFindProjectConfigDoesNotFallbackByName(t *testing.T) {
	cfg := Config{Projects: []ProjectConfig{
		{Path: "/old/path", TicketPrefix: TicketPrefixes{"XYZ-"}},
	}}
	_, missing := cfg.FindProjectConfig("/new/path")
	require.True(t, missing)
}

func TestFindProjectConfigMissing(t *testing.T) {
	cfg := Config{}
	p, missing := cfg.FindProjectConfig("/code/newproject")
	require.True(t, missing)
	require.Equal(t, ProjectConfig{}, ProjectConfig{
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
			{Path: "/code/myproject", BaseDir: "~/.wtree/myproject", TicketPrefix: TicketPrefixes{"AB-"}},
		},
	}
	require.NoError(t, cfg.Save())

	loaded, err := Load()
	require.NoError(t, err)
	require.Len(t, loaded.Projects, 1)
	require.Equal(t, TicketPrefixes{"AB-"}, loaded.Projects[0].TicketPrefix)
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

	got, err := LoadOrCreateProjectConfig("/code/myproject")
	require.NoError(t, err)
	require.Equal(t, "/code/myproject", got.Path)
	require.Equal(t, "/code/myproject/.wtree", got.BaseDir)

	loaded, err := Load()
	require.NoError(t, err)
	require.Len(t, loaded.Projects, 1)
}

func TestLoadOrCreateProjectConfigDifferentPathsCreateSeparateEntries(t *testing.T) {
	dir := t.TempDir()
	origConfigPath := ConfigPath
	ConfigPath = func() string { return filepath.Join(dir, "config.yml") }
	t.Cleanup(func() { ConfigPath = origConfigPath })

	first, err := LoadOrCreateProjectConfig("/code/team-a/api")
	require.NoError(t, err)

	second, err := LoadOrCreateProjectConfig("/code/team-b/api")
	require.NoError(t, err)

	require.Equal(t, "/code/team-a/api", first.Path)
	require.Equal(t, "/code/team-b/api", second.Path)

	loaded, err := Load()
	require.NoError(t, err)
	require.Len(t, loaded.Projects, 2)
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

func TestTicketPrefixUnmarshalScalar(t *testing.T) {
	var p ProjectConfig
	require.NoError(t, yaml.Unmarshal([]byte("path: /a\nticket_prefix: AB-\n"), &p))
	require.Equal(t, TicketPrefixes{"AB-"}, p.TicketPrefix)
}

func TestTicketPrefixUnmarshalList(t *testing.T) {
	var p ProjectConfig
	require.NoError(t, yaml.Unmarshal([]byte("path: /a\nticket_prefix: [AB-, BCD-]\n"), &p))
	require.Equal(t, TicketPrefixes{"AB-", "BCD-"}, p.TicketPrefix)
}

func TestTicketPrefixUnmarshalEmptyScalar(t *testing.T) {
	var p ProjectConfig
	require.NoError(t, yaml.Unmarshal([]byte("path: /a\nticket_prefix: \"\"\n"), &p))
	require.Empty(t, p.TicketPrefix)
}

func TestTicketPrefixUnmarshalInvalid(t *testing.T) {
	var p ProjectConfig
	err := yaml.Unmarshal([]byte("path: /a\nticket_prefix: {a: b}\n"), &p)
	require.ErrorContains(t, err, "ticket_prefix")
}

func TestTicketPrefixMarshal(t *testing.T) {
	cases := []struct {
		name   string
		prefix TicketPrefixes
		want   string
	}{
		{"single stays scalar", TicketPrefixes{"AB-"}, "path: /a\nticket_prefix: AB-\n"},
		{"many as list", TicketPrefixes{"AB-", "BCD-"}, "path: /a\nticket_prefix:\n    - AB-\n    - BCD-\n"},
		{"empty omitted", nil, "path: /a\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			data, err := yaml.Marshal(ProjectConfig{Path: "/a", TicketPrefix: c.prefix})
			require.NoError(t, err)
			require.Equal(t, c.want, string(data))

			var p ProjectConfig
			require.NoError(t, yaml.Unmarshal(data, &p))
			require.Equal(t, c.prefix, p.TicketPrefix)
		})
	}
}

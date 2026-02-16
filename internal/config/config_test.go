package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_AllFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".git-sandwich.yml")
	content := `start: "BEGIN SANDWICH"
end: "END SANDWICH"
base: "origin/develop"
head: "feature-branch"
allow_nesting: true
allow_boundary_with_outside: true
json: true
include:
  - "*.go"
  - "*.rb"
exclude:
  - "vendor/**"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Start != "BEGIN SANDWICH" {
		t.Errorf("Start = %q, want %q", cfg.Start, "BEGIN SANDWICH")
	}
	if cfg.End != "END SANDWICH" {
		t.Errorf("End = %q, want %q", cfg.End, "END SANDWICH")
	}
	if cfg.Base != "origin/develop" {
		t.Errorf("Base = %q, want %q", cfg.Base, "origin/develop")
	}
	if cfg.Head != "feature-branch" {
		t.Errorf("Head = %q, want %q", cfg.Head, "feature-branch")
	}
	if !cfg.AllowNesting {
		t.Error("AllowNesting = false, want true")
	}
	if !cfg.AllowBoundaryWithOutside {
		t.Error("AllowBoundaryWithOutside = false, want true")
	}
	if !cfg.JSON {
		t.Error("JSON = false, want true")
	}
	if len(cfg.Include) != 2 || cfg.Include[0] != "*.go" || cfg.Include[1] != "*.rb" {
		t.Errorf("Include = %v, want [*.go *.rb]", cfg.Include)
	}
	if len(cfg.Exclude) != 1 || cfg.Exclude[0] != "vendor/**" {
		t.Errorf("Exclude = %v, want [vendor/**]", cfg.Exclude)
	}
}

func TestLoad_PartialFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".git-sandwich.yml")
	content := `start: "# BEGIN"
end: "# END"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Start != "# BEGIN" {
		t.Errorf("Start = %q, want %q", cfg.Start, "# BEGIN")
	}
	if cfg.End != "# END" {
		t.Errorf("End = %q, want %q", cfg.End, "# END")
	}
	if cfg.Base != "" {
		t.Errorf("Base = %q, want empty", cfg.Base)
	}
	if cfg.Head != "" {
		t.Errorf("Head = %q, want empty", cfg.Head)
	}
	if cfg.AllowNesting {
		t.Error("AllowNesting = true, want false")
	}
	if cfg.JSON {
		t.Error("JSON = true, want false")
	}
	if cfg.Include != nil {
		t.Errorf("Include = %v, want nil", cfg.Include)
	}
	if cfg.Exclude != nil {
		t.Errorf("Exclude = %v, want nil", cfg.Exclude)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".git-sandwich.yml")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Start != "" {
		t.Errorf("Start = %q, want empty", cfg.Start)
	}
	if cfg.End != "" {
		t.Errorf("End = %q, want empty", cfg.End)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".git-sandwich.yml")
	content := `start: [invalid yaml
  : broken`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/.git-sandwich.yml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

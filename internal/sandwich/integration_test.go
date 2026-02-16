package sandwich

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init")
	run("checkout", "-b", "main")

	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func commit(t *testing.T, dir, msg string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
	run("add", "-A")
	run("commit", "-m", msg)
}

func makeCfg() *Config {
	return &Config{
		StartMarkerRegex:         regexp.MustCompile(`# START`),
		EndMarkerRegex:           regexp.MustCompile(`# END`),
		BaseRef:                  "main",
		HeadRef:                  "HEAD",
		AllowNesting:             false,
		AllowBoundaryWithOutside: false,
	}
}

func TestIntegration_InsideChange_PASS(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	// Create feature branch and modify inside block
	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	writeFile(t, dir, "app.rb", "line 1\n# START\nmodified\n# END\nline 5\n")
	commit(t, dir, "change inside")

	cfg := makeCfg()
	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got failure: %+v", result.Files)
	}
}

func TestIntegration_OutsideChange_FAIL(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	writeFile(t, dir, "app.rb", "CHANGED\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "change outside")

	cfg := makeCfg()
	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for outside change")
	}
}

func TestIntegration_BoundaryOnlyChange_PASS(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	// Change boundary marker (add more content to START line)
	writeFile(t, dir, "app.rb", "line 1\n# START v2\noriginal\n# END v2\nline 5\n")
	commit(t, dir, "change boundary")

	cfg := makeCfg()
	// Need to update regex to match both versions
	cfg.StartMarkerRegex = regexp.MustCompile(`# START`)
	cfg.EndMarkerRegex = regexp.MustCompile(`# END`)

	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success for boundary-only change, got failure: %+v", result.Files)
	}
}

func TestIntegration_BoundaryPlusOutside_FAIL(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	// Change both boundary and outside
	writeFile(t, dir, "app.rb", "CHANGED\n# START v2\noriginal\n# END v2\nline 5\n")
	commit(t, dir, "change boundary + outside")

	cfg := makeCfg()
	cfg.StartMarkerRegex = regexp.MustCompile(`# START`)
	cfg.EndMarkerRegex = regexp.MustCompile(`# END`)

	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for boundary + outside change")
	}
}

func TestIntegration_BoundaryPlusOutside_AllowFlag_PASS(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	writeFile(t, dir, "app.rb", "CHANGED\n# START v2\noriginal\n# END v2\nline 5\n")
	commit(t, dir, "change boundary + outside")

	cfg := makeCfg()
	cfg.StartMarkerRegex = regexp.MustCompile(`# START`)
	cfg.EndMarkerRegex = regexp.MustCompile(`# END`)
	cfg.AllowBoundaryWithOutside = true

	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success with allow-boundary-with-outside, got failure: %+v", result.Files)
	}
}

func TestIntegration_NewFile_PASS(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "existing.rb", "no blocks here\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	writeFile(t, dir, "new.rb", "# START\ncontent\n# END\n")
	commit(t, dir, "add new file")

	cfg := makeCfg()
	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success for new file, got failure: %+v", result.Files)
	}
}

func TestIntegration_BrokenBlocks_FAIL(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "# START\noriginal\n# END\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	// Break block structure: remove END
	writeFile(t, dir, "app.rb", "# START\nmodified\n")
	commit(t, dir, "break blocks")

	cfg := makeCfg()
	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for broken blocks")
	}
	if len(result.Files) > 0 && result.Files[0].BlockError == "" {
		t.Error("expected block error")
	}
}

func TestIntegration_NoBlocks_Skip(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	writeFile(t, dir, "app.rb", "line 1\nline 2\nline 3\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	writeFile(t, dir, "app.rb", "changed\nline 2\nline 3\n")
	commit(t, dir, "change file without blocks")

	cfg := makeCfg()
	result, err := Validate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Error("expected success for file without blocks")
	}
}

func TestIntegration_IncludeExclude(t *testing.T) {
	dir := setupTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Base: two files with blocks, one outside change would fail
	writeFile(t, dir, "src/app.go", "line 1\n# START\noriginal\n# END\nline 5\n")
	writeFile(t, dir, "vendor/lib.go", "line 1\n# START\noriginal\n# END\nline 5\n")
	writeFile(t, dir, "docs/guide.md", "line 1\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "base")

	cmd := exec.Command("git", "checkout", "-b", "feature")
	cmd.Dir = dir
	cmd.Run()

	// Change inside block in src/app.go (PASS), outside block in vendor/lib.go and docs/guide.md (FAIL)
	writeFile(t, dir, "src/app.go", "line 1\n# START\nmodified\n# END\nline 5\n")
	writeFile(t, dir, "vendor/lib.go", "CHANGED\n# START\noriginal\n# END\nline 5\n")
	writeFile(t, dir, "docs/guide.md", "CHANGED\n# START\noriginal\n# END\nline 5\n")
	commit(t, dir, "changes in multiple files")

	t.Run("no filter fails", func(t *testing.T) {
		cfg := makeCfg()
		result, err := Validate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Success {
			t.Error("expected failure without filters")
		}
	})

	t.Run("include only src passes", func(t *testing.T) {
		cfg := makeCfg()
		cfg.IncludePatterns = []string{"src/**"}
		result, err := Validate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success with include=src/**, got failure: %+v", result.Files)
		}
		if len(result.Files) != 1 {
			t.Errorf("expected 1 file, got %d", len(result.Files))
		}
	})

	t.Run("exclude vendor and docs passes", func(t *testing.T) {
		cfg := makeCfg()
		cfg.ExcludePatterns = []string{"vendor", "docs"}
		result, err := Validate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success with exclude=vendor,docs, got failure: %+v", result.Files)
		}
		if len(result.Files) != 1 {
			t.Errorf("expected 1 file, got %d", len(result.Files))
		}
	})

	t.Run("include and exclude combined", func(t *testing.T) {
		cfg := makeCfg()
		cfg.IncludePatterns = []string{"**/*.go"}
		cfg.ExcludePatterns = []string{"vendor"}
		result, err := Validate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Errorf("expected success with include=**/*.go exclude=vendor, got failure: %+v", result.Files)
		}
		if len(result.Files) != 1 {
			t.Errorf("expected 1 file, got %d", len(result.Files))
		}
	})
}

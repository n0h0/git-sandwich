package sandwich

import (
	"testing"

	"github.com/n0h0/git-sandwich/internal/diff"
)

func makeFileDiffs(paths ...string) []diff.FileDiff {
	var fds []diff.FileDiff
	for _, p := range paths {
		fds = append(fds, diff.FileDiff{NewPath: p, OldPath: p})
	}
	return fds
}

func filePaths(fds []diff.FileDiff) []string {
	var paths []string
	for _, fd := range fds {
		paths = append(paths, fd.NewPath)
	}
	return paths
}

func TestFilter_NoPatterns(t *testing.T) {
	files := makeFileDiffs("a.go", "b.rb", "vendor/c.go")
	result := filterFiles(files, nil, nil)
	if len(result) != 3 {
		t.Errorf("expected 3 files, got %d", len(result))
	}
}

func TestFilter_IncludeOnly(t *testing.T) {
	files := makeFileDiffs("src/main.go", "src/util.go", "README.md", "docs/guide.md")
	result := filterFiles(files, []string{"**/*.go"}, nil)
	paths := filePaths(result)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
	for _, p := range paths {
		if p != "src/main.go" && p != "src/util.go" {
			t.Errorf("unexpected file: %s", p)
		}
	}
}

func TestFilter_ExcludeOnly(t *testing.T) {
	files := makeFileDiffs("src/main.go", "vendor/lib.go", "vendor/dep/dep.go")
	result := filterFiles(files, nil, []string{"vendor/**"})
	paths := filePaths(result)
	if len(paths) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(paths), paths)
	}
	if paths[0] != "src/main.go" {
		t.Errorf("expected src/main.go, got %s", paths[0])
	}
}

func TestFilter_IncludeAndExclude(t *testing.T) {
	files := makeFileDiffs("src/main.go", "src/main_test.go", "README.md")
	result := filterFiles(files, []string{"**/*.go"}, []string{"**/*_test.go"})
	paths := filePaths(result)
	if len(paths) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(paths), paths)
	}
	if paths[0] != "src/main.go" {
		t.Errorf("expected src/main.go, got %s", paths[0])
	}
}

func TestFilter_DirectoryPattern(t *testing.T) {
	files := makeFileDiffs("docs/guide.md", "docs/api/ref.md", "src/main.go")
	result := filterFiles(files, nil, []string{"docs"})
	paths := filePaths(result)
	if len(paths) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(paths), paths)
	}
	if paths[0] != "src/main.go" {
		t.Errorf("expected src/main.go, got %s", paths[0])
	}
}

func TestFilter_DeletedFile(t *testing.T) {
	files := []diff.FileDiff{
		{OldPath: "old/deleted.go", IsDeleted: true},
		{NewPath: "src/main.go", OldPath: "src/main.go"},
	}
	result := filterFiles(files, []string{"src/**"}, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result))
	}
	if result[0].NewPath != "src/main.go" {
		t.Errorf("expected src/main.go, got %s", result[0].NewPath)
	}
}

func TestFilter_StarGlob(t *testing.T) {
	files := makeFileDiffs("main.go", "util.go", "main.rb")
	result := filterFiles(files, []string{"*.go"}, nil)
	paths := filePaths(result)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
}

func TestFilter_DoubleStarGlob(t *testing.T) {
	files := makeFileDiffs("a.go", "pkg/b.go", "pkg/sub/c.go", "README.md")
	result := filterFiles(files, []string{"**/*.go"}, nil)
	paths := filePaths(result)
	if len(paths) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(paths), paths)
	}
}

func TestFilter_MultipleIncludePatterns(t *testing.T) {
	files := makeFileDiffs("main.go", "style.css", "index.html", "config.yaml")
	result := filterFiles(files, []string{"*.go", "*.css"}, nil)
	paths := filePaths(result)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
}

func TestFilter_MultipleExcludePatterns(t *testing.T) {
	files := makeFileDiffs("main.go", "vendor/lib.go", "generated/code.go", "src/app.go")
	result := filterFiles(files, nil, []string{"vendor", "generated"})
	paths := filePaths(result)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
}

func TestFilter_IncludeDirectory(t *testing.T) {
	files := makeFileDiffs("src/main.go", "src/sub/util.go", "lib/helper.go")
	result := filterFiles(files, []string{"src"}, nil)
	paths := filePaths(result)
	if len(paths) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(paths), paths)
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"main.go", "*.go", true},
		{"main.go", "*.rb", false},
		{"src/main.go", "**/*.go", true},
		{"src/sub/main.go", "**/*.go", true},
		{"docs/guide.md", "docs", true},
		{"docs/sub/guide.md", "docs", true},
		{"mydocs/file.md", "docs", false},
		{"src/main_test.go", "**/*_test.go", true},
		{"main_test.go", "*_test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_"+tt.pattern, func(t *testing.T) {
			got := matchesPattern(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("matchesPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

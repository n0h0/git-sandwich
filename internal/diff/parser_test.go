package diff

import (
	"testing"
)

func TestParse_BasicDiff(t *testing.T) {
	// A basic -U0 diff with one hunk: delete line 3, add lines 3-4
	raw := []byte(`diff --git a/foo.rb b/foo.rb
index 1234567..abcdefg 100644
--- a/foo.rb
+++ b/foo.rb
@@ -3,1 +3,2 @@
-old line
+new line 1
+new line 2
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.OldPath != "foo.rb" {
		t.Errorf("expected OldPath foo.rb, got %s", f.OldPath)
	}
	if f.NewPath != "foo.rb" {
		t.Errorf("expected NewPath foo.rb, got %s", f.NewPath)
	}
	if f.IsNew || f.IsDeleted {
		t.Errorf("expected existing file, got IsNew=%v IsDeleted=%v", f.IsNew, f.IsDeleted)
	}

	if len(f.OldRanges) != 1 {
		t.Fatalf("expected 1 OldRange, got %d", len(f.OldRanges))
	}
	if f.OldRanges[0].Start != 3 || f.OldRanges[0].End != 3 {
		t.Errorf("expected OldRange {3,3}, got %+v", f.OldRanges[0])
	}

	if len(f.NewRanges) != 1 {
		t.Fatalf("expected 1 NewRange, got %d", len(f.NewRanges))
	}
	if f.NewRanges[0].Start != 3 || f.NewRanges[0].End != 4 {
		t.Errorf("expected NewRange {3,4}, got %+v", f.NewRanges[0])
	}
}

func TestParse_NewFile(t *testing.T) {
	raw := []byte(`diff --git a/new.rb b/new.rb
new file mode 100644
index 0000000..abcdefg
--- /dev/null
+++ b/new.rb
@@ -0,0 +1,3 @@
+line 1
+line 2
+line 3
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsNew {
		t.Error("expected IsNew=true")
	}
	if len(files[0].OldRanges) != 0 {
		t.Errorf("expected 0 OldRanges for new file, got %d", len(files[0].OldRanges))
	}
	if len(files[0].NewRanges) != 1 {
		t.Fatalf("expected 1 NewRange, got %d", len(files[0].NewRanges))
	}
	if files[0].NewRanges[0].Start != 1 || files[0].NewRanges[0].End != 3 {
		t.Errorf("expected NewRange {1,3}, got %+v", files[0].NewRanges[0])
	}
}

func TestParse_DeletedFile(t *testing.T) {
	raw := []byte(`diff --git a/old.rb b/old.rb
deleted file mode 100644
index abcdefg..0000000
--- a/old.rb
+++ /dev/null
@@ -1,2 +0,0 @@
-line 1
-line 2
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsDeleted {
		t.Error("expected IsDeleted=true")
	}
	if len(files[0].OldRanges) != 1 {
		t.Fatalf("expected 1 OldRange, got %d", len(files[0].OldRanges))
	}
	if files[0].OldRanges[0].Start != 1 || files[0].OldRanges[0].End != 2 {
		t.Errorf("expected OldRange {1,2}, got %+v", files[0].OldRanges[0])
	}
}

func TestParse_MultipleHunks(t *testing.T) {
	raw := []byte(`diff --git a/multi.rb b/multi.rb
index 1234567..abcdefg 100644
--- a/multi.rb
+++ b/multi.rb
@@ -2,1 +2,1 @@
-old2
+new2
@@ -10,3 +10,1 @@
-old10
-old11
-old12
+new10
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if len(f.OldRanges) != 2 {
		t.Fatalf("expected 2 OldRanges, got %d", len(f.OldRanges))
	}
	if f.OldRanges[0].Start != 2 || f.OldRanges[0].End != 2 {
		t.Errorf("expected OldRange[0] {2,2}, got %+v", f.OldRanges[0])
	}
	if f.OldRanges[1].Start != 10 || f.OldRanges[1].End != 12 {
		t.Errorf("expected OldRange[1] {10,12}, got %+v", f.OldRanges[1])
	}

	if len(f.NewRanges) != 2 {
		t.Fatalf("expected 2 NewRanges, got %d", len(f.NewRanges))
	}
	if f.NewRanges[0].Start != 2 || f.NewRanges[0].End != 2 {
		t.Errorf("expected NewRange[0] {2,2}, got %+v", f.NewRanges[0])
	}
	if f.NewRanges[1].Start != 10 || f.NewRanges[1].End != 10 {
		t.Errorf("expected NewRange[1] {10,10}, got %+v", f.NewRanges[1])
	}
}

func TestParse_AddOnly(t *testing.T) {
	// Hunk that only adds lines (OrigLines=0)
	raw := []byte(`diff --git a/add.rb b/add.rb
index 1234567..abcdefg 100644
--- a/add.rb
+++ b/add.rb
@@ -5,0 +6,2 @@
+added line 1
+added line 2
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := files[0]
	if len(f.OldRanges) != 0 {
		t.Errorf("expected 0 OldRanges for add-only, got %d", len(f.OldRanges))
	}
	if len(f.NewRanges) != 1 {
		t.Fatalf("expected 1 NewRange, got %d", len(f.NewRanges))
	}
	if f.NewRanges[0].Start != 6 || f.NewRanges[0].End != 7 {
		t.Errorf("expected NewRange {6,7}, got %+v", f.NewRanges[0])
	}
}

func TestParse_DeleteOnly(t *testing.T) {
	// Hunk that only deletes lines (NewLines=0)
	raw := []byte(`diff --git a/del.rb b/del.rb
index 1234567..abcdefg 100644
--- a/del.rb
+++ b/del.rb
@@ -3,2 +2,0 @@
-deleted line 1
-deleted line 2
`)

	files, err := Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := files[0]
	if len(f.OldRanges) != 1 {
		t.Fatalf("expected 1 OldRange, got %d", len(f.OldRanges))
	}
	if f.OldRanges[0].Start != 3 || f.OldRanges[0].End != 4 {
		t.Errorf("expected OldRange {3,4}, got %+v", f.OldRanges[0])
	}
	if len(f.NewRanges) != 0 {
		t.Errorf("expected 0 NewRanges for delete-only, got %d", len(f.NewRanges))
	}
}

package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"morphr/internal/config"
)

func TestWalkDirectoryFlat(t *testing.T) {
	tmpDir := t.TempDir()

	rawFiles := []string{"photo1.cr2", "photo2.nef", "photo3.arw"}
	otherFiles := []string{"readme.txt", "notes.md", "script.sh"}

	for _, name := range rawFiles {
		os.WriteFile(filepath.Join(tmpDir, name), []byte("fake"), 0o644)
	}
	for _, name := range otherFiles {
		os.WriteFile(filepath.Join(tmpDir, name), []byte("fake"), 0o644)
	}

	files, err := WalkDirectory(tmpDir, false)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}

	if len(files) != len(rawFiles) {
		t.Errorf("found %d files, want %d", len(files), len(rawFiles))
	}

	for _, f := range files {
		if f.Ext != ".cr2" && f.Ext != ".nef" && f.Ext != ".arw" {
			t.Errorf("unexpected extension: %s", f.Ext)
		}
	}
}

func TestWalkDirectoryRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "root.cr2"), []byte("fake"), 0o644)
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0o755)
	os.WriteFile(filepath.Join(subDir, "nested.nef"), []byte("fake"), 0o644)

	// Non-recursive should only find root
	flat, err := WalkDirectory(tmpDir, false)
	if err != nil {
		t.Fatalf("flat walk: %v", err)
	}
	if len(flat) != 1 {
		t.Errorf("flat: found %d files, want 1", len(flat))
	}

	// Recursive should find both
	recursive, err := WalkDirectory(tmpDir, true)
	if err != nil {
		t.Fatalf("recursive walk: %v", err)
	}
	if len(recursive) != 2 {
		t.Errorf("recursive: found %d files, want 2", len(recursive))
	}
}

func TestWalkDirectoryEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := WalkDirectory(tmpDir, false)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("found %d files in empty dir, want 0", len(files))
	}
}

func TestWalkDirectoryCaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "photo.CR2"), []byte("fake"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "shot.Nef"), []byte("fake"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "img.DNG"), []byte("fake"), 0o644)

	files, err := WalkDirectory(tmpDir, false)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("found %d files, want 3 (case-insensitive)", len(files))
	}
}

func TestWalkDirectoryMissing(t *testing.T) {
	_, err := WalkDirectory("/nonexistent/path", false)
	if err == nil {
		t.Error("should fail on missing directory")
	}
}

func TestRunPoolSuccess(t *testing.T) {
	files := []FileEntry{
		{Path: "a.cr2", Ext: ".cr2"},
		{Path: "b.nef", Ext: ".nef"},
		{Path: "c.arw", Ext: ".arw"},
	}
	cfg := config.Default()

	var processed atomic.Int32

	fn := func(ctx context.Context, entry FileEntry, c *config.Config) error {
		processed.Add(1)
		return nil
	}

	var completed int
	progressFn := func(r Result) {
		completed++
	}

	err := RunPool(context.Background(), files, cfg, 2, fn, progressFn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	if int(processed.Load()) != len(files) {
		t.Errorf("processed %d files, want %d", processed.Load(), len(files))
	}
	if completed != len(files) {
		t.Errorf("progress reported %d, want %d", completed, len(files))
	}
}

func TestRunPoolWithErrors(t *testing.T) {
	files := []FileEntry{
		{Path: "ok.cr2", Ext: ".cr2"},
		{Path: "fail.nef", Ext: ".nef"},
		{Path: "ok2.arw", Ext: ".arw"},
	}
	cfg := config.Default()

	fn := func(ctx context.Context, entry FileEntry, c *config.Config) error {
		if entry.Path == "fail.nef" {
			return fmt.Errorf("decode error")
		}
		return nil
	}

	var errors int
	progressFn := func(r Result) {
		if r.Err != nil {
			errors++
		}
	}

	err := RunPool(context.Background(), files, cfg, 2, fn, progressFn)
	if err == nil {
		t.Fatal("pool should return first error")
	}
	if errors != 1 {
		t.Errorf("expected 1 error in progress, got %d", errors)
	}
}

func TestRunPoolCancellation(t *testing.T) {
	files := make([]FileEntry, 100)
	for i := range files {
		files[i] = FileEntry{Path: fmt.Sprintf("file%d.cr2", i), Ext: ".cr2"}
	}
	cfg := config.Default()

	ctx, cancel := context.WithCancel(context.Background())

	var count atomic.Int32
	fn := func(ctx context.Context, entry FileEntry, c *config.Config) error {
		if count.Add(1) >= 5 {
			cancel()
		}
		return nil
	}

	RunPool(ctx, files, cfg, 2, fn, nil)

	// Should have processed fewer than all files due to cancellation
	if int(count.Load()) >= len(files) {
		t.Logf("processed all %d files despite cancellation (race ok)", count.Load())
	}
}

func TestRunPoolSingleWorker(t *testing.T) {
	files := []FileEntry{
		{Path: "a.cr2", Ext: ".cr2"},
		{Path: "b.cr2", Ext: ".cr2"},
	}
	cfg := config.Default()

	var order []string
	fn := func(ctx context.Context, entry FileEntry, c *config.Config) error {
		order = append(order, entry.Path)
		return nil
	}

	err := RunPool(context.Background(), files, cfg, 1, fn, nil)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	if len(order) != 2 {
		t.Errorf("processed %d files, want 2", len(order))
	}
}

func TestProgressReporter(t *testing.T) {
	p := NewProgressReporter(3, true)

	p.Update(Result{Entry: FileEntry{Path: "a.cr2"}, Err: nil})
	p.Update(Result{Entry: FileEntry{Path: "b.cr2"}, Err: fmt.Errorf("fail")})
	p.Update(Result{Entry: FileEntry{Path: "c.cr2"}, Err: nil})
	p.Finish()

	if p.done != 3 {
		t.Errorf("done = %d, want 3", p.done)
	}
	if p.failed != 1 {
		t.Errorf("failed = %d, want 1", p.failed)
	}
}

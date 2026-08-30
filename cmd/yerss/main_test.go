package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResetDatabaseRemovesFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sqlite")
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := resetDatabase(path); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{path, path + "-wal", path + "-shm"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s should be removed, stat err = %v", p, err)
		}
	}
}

func TestResetDatabaseMissingIsNoOp(t *testing.T) {
	if err := resetDatabase(filepath.Join(t.TempDir(), "absent.sqlite")); err != nil {
		t.Errorf("missing database should not error, got %v", err)
	}
}
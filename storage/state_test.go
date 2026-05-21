package storage

import (
	"path/filepath"
	"testing"
)

func TestStateStorage(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_state.json")

	// Test Read missing file (should return 0)
	block, err := ReadLatestBlock(testFile)
	if err != nil {
		t.Fatalf("ReadLatestBlock on missing file failed: %v", err)
	}
	if block != 0 {
		t.Errorf("Expected block 0, got %d", block)
	}

	// Test Write
	err = WriteLatestBlock(testFile, 12345)
	if err != nil {
		t.Fatalf("WriteLatestBlock failed: %v", err)
	}

	// Test Read existing file
	block, err = ReadLatestBlock(testFile)
	if err != nil {
		t.Fatalf("ReadLatestBlock on existing file failed: %v", err)
	}
	if block != 12345 {
		t.Errorf("Expected block 12345, got %d", block)
	}
}
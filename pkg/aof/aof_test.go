package aof

import (
	"os"
	"redis-lite/pkg/cfg"
	"strings"
	"testing"
	"time"
)

func TestAof(t *testing.T) {
	f, err := os.CreateTemp("", "aof_test_*.aof")
	if err != nil {
		t.Fatal(err)
	}
	tempPath := f.Name()
	f.Close()
	defer os.Remove(tempPath)

	mockConfig := &cfg.Config{
		Host:            "localhost",
		Port:            "0",
		ServerType:      "tcp",
		AofPath:         tempPath,
		JanitorInterval: 10 * time.Millisecond,
	}

	fileName := f.Name()

	// Close the temp file handle so our AOF logic can open it freshly
	f.Close()
	defer os.Remove(fileName)

	// 2. Initialize AOF
	aof, err := NewAof(mockConfig)
	if err != nil {
		t.Fatalf("Failed to create AOF: %v", err)
	}

	// 3. Test WRITE
	// We simulate a few Redis commands
	cmd1 := "SET key value"
	cmd2 := "HSET user:1 name john"
	cmd3 := "LPUSH list item"

	if err := aof.Write(cmd1); err != nil {
		t.Errorf("Failed to write cmd1: %v", err)
	}
	if err := aof.Write(cmd2); err != nil {
		t.Errorf("Failed to write cmd2: %v", err)
	}
	if err := aof.Write(cmd3); err != nil {
		t.Errorf("Failed to write cmd3: %v", err)
	}

	// 4. Test READ (Verification 1: Immediate Read)
	// We read the file we just wrote to (simulating checking consistency)
	var lines []string

	err = aof.Read(func(cmd string) {
		// AOF stores with \n, so we trim it to compare
		lines = append(lines, strings.TrimSpace(cmd))
	})
	if err != nil {
		t.Fatalf("Failed to read AOF: %v", err)
	}

	if len(lines) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0] != cmd1 {
		t.Errorf("Line 1 mismatch. Want '%s', got '%s'", cmd1, lines[0])
	}
	if lines[1] != cmd2 {
		t.Errorf("Line 2 mismatch. Want '%s', got '%s'", cmd2, lines[1])
	}

	// 5. Test PERSISTENCE (Simulate Server Restart)
	// We close the file and open a new instance pointing to the same path
	if err := aof.Close(); err != nil {
		t.Fatalf("Failed to close AOF: %v", err)
	}

	aofRestart, err := NewAof(mockConfig)
	if err != nil {
		t.Fatalf("Failed to re-open AOF: %v", err)
	}
	defer aofRestart.Close()

	// Read again from the "new" server instance
	var restoredLines []string
	err = aofRestart.Read(func(cmd string) {
		restoredLines = append(restoredLines, strings.TrimSpace(cmd))
	})
	if err != nil {
		t.Fatalf("Failed to read AOF after restart: %v", err)
	}

	if len(restoredLines) != 3 {
		t.Fatalf("Expected 3 restored lines, got %d", len(restoredLines))
	}
	if restoredLines[2] != cmd3 {
		t.Errorf("Restored Line 3 mismatch. Want '%s', got '%s'", cmd3, restoredLines[2])
	}
}

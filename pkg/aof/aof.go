package aof

import (
	"bufio"
	"io"
	"os"
	"redis-lite/pkg/cfg"
	"sync"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

// NewAof opens (or creates) the database file.
func NewAof(config *cfg.Config) (*Aof, error) {
	f, err := os.OpenFile(config.AofPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	return &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}, nil
}

// Close ensures the file is properly closed
func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()
	return aof.file.Close()
}

// Write adds a new command to the file
// Ideally, we would batch this or use a channel,
// but for the MVP (our current structure) a Mutex + WriteString is safer to ensure order.
// TODO: move this to a background channel.
func (aof *Aof) Write(command string) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.WriteString(command + "\n")
	if err != nil {
		return err
	}

	// Optional: sync to disk immediately (safe but slow)
	// or let OS handle it (fast but risky)
	// for now, we will let OS handle it.
	return nil
}

func (aof *Aof) Read(callback func(command string)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	_, err := aof.file.Seek(0, 0)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(aof.file)

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// execute the callback which will be our command handler
		callback(line)
	}

	return nil
}

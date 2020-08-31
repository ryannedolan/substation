package appender

import (
	"time"
	"fmt"
	"os"
	"path"
)

type Appender struct {
	*os.File
}

// Create a new Appender in the given directory
func Create(dir string) (*Appender, error) {
	filebase := fmt.Sprintf("%d.log", time.Now().UnixNano())
	filename := path.Join(dir, filebase)
	if f, err := os.Create(filename); err != nil {
		return nil, fmt.Errorf("Could not create file %s: %v", filename, err)
	} else {
		return &Appender{f}, nil
	}
}

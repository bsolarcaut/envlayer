package resolver

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DotenvResolver resolves environment variables from a .env file.
type DotenvResolver struct {
	filePath string
	cache    map[string]string
}

// NewDotenvResolver creates a new DotenvResolver for the given file path.
func NewDotenvResolver(filePath string) *DotenvResolver {
	return &DotenvResolver{
		filePath: filePath,
		cache:    nil,
	}
}

// Resolve returns the value for the given key from the .env file.
// Returns an empty string and no error if the key is not found.
func (d *DotenvResolver) Resolve(key string) (string, error) {
	if d.cache == nil {
		if err := d.load(); err != nil {
			return "", fmt.Errorf("dotenv: failed to load file %q: %w", d.filePath, err)
		}
	}
	return d.cache[key], nil
}

// Name returns the resolver backend name.
func (d *DotenvResolver) Name() string {
	return "dotenv"
}

// load reads and parses the .env file into the cache.
func (d *DotenvResolver) load() error {
	f, err := os.Open(d.filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	d.cache = make(map[string]string)
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid syntax on line %d: %q", lineNum, line)
		}

		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		v = strings.Trim(v, `"'`)

		if k == "" {
			return fmt.Errorf("empty key on line %d", lineNum)
		}

		d.cache[k] = v
	}

	return scanner.Err()
}

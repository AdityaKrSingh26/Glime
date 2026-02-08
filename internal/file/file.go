package file

import (
	"bufio"
	"fmt"
	"os"
)

/**
* Load func - reads a file and returns its lines
* Example :
* 	File content :
* 		hello
* 		world
* returns => []string{"hello", "world"}
 */
func Load(filePath string) ([]string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file :%w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// ensure atlease one line
	if len(lines) == 0 {
		lines = []string{""}
	}

	return lines, nil
}

// saves write lines to a file
func Save(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	for i, line := range lines {
		if _, err := writer.WriteString(line); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}

		// add new line except for last line
		if i < len(lines)-1 {
			if _, err := writer.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write new line: %w", err)
			}
		}
	}

	return nil
}

// Exists check if file exists
func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return true
}

// IsReadable to check if a file is readable
func IsReadable(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}

	file.Close()
	return true
}

// Backup creates a backup copy of a file so that original data is not lost
func Backup(filePath string) error {
	if !Exists(filePath) {
		return nil // nothing to backup
	}

	backupPath := filePath + "~"
	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	if err := os.WriteFile(backupPath, input, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

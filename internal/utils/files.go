package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/thanhfphan/svg2sheet/internal/config"
)

// FileInfo holds information about a file for sorting
type FileInfo struct {
	Path  string
	Name  string
	CTime time.Time
}

// FileMapping holds the mapping between original files and processed PNG files
type FileMapping struct {
	PNGPath      string
	OriginalPath string
	IsTemporary  bool
}

// SortFiles sorts files according to the specified mode
func SortFiles(files []string, mode config.SortMode) ([]string, error) {
	if len(files) == 0 {
		return files, nil
	}

	switch mode {
	case config.SortByName:
		return sortByName(files), nil
	case config.SortByCTime:
		return sortByCTime(files)
	case config.SortManual:
		// Manual sorting - return as-is (user should provide files in desired order)
		return files, nil
	default:
		return nil, fmt.Errorf("unsupported sort mode: %s", mode)
	}
}

// sortByName sorts files alphabetically by filename
func sortByName(files []string) []string {
	sorted := make([]string, len(files))
	copy(sorted, files)

	sort.Slice(sorted, func(i, j int) bool {
		nameI := filepath.Base(sorted[i])
		nameJ := filepath.Base(sorted[j])
		return nameI < nameJ
	})

	return sorted
}

// sortByCTime sorts files by creation/modification time
func sortByCTime(files []string) ([]string, error) {
	fileInfos := make([]FileInfo, 0, len(files))

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("failed to stat file %s: %w", file, err)
		}

		fileInfos = append(fileInfos, FileInfo{
			Path:  file,
			Name:  filepath.Base(file),
			CTime: info.ModTime(), // Use ModTime as creation time approximation
		})
	}

	// Sort by creation time
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].CTime.Before(fileInfos[j].CTime)
	})

	// Extract sorted paths
	sorted := make([]string, len(fileInfos))
	for i, info := range fileInfos {
		sorted[i] = info.Path
	}

	return sorted, nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %w", err)
	}

	if err := dstFile.Chmod(srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	return nil
}

// CreateTempFile creates a temporary file with the given extension
func CreateTempFile(ext string) (string, error) {
	tempFile, err := os.CreateTemp("", "svg2sheet_*"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	tempPath := tempFile.Name()
	tempFile.Close()

	return tempPath, nil
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDirectory checks if a path is a directory
func IsDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// GetFileExtension returns the file extension (including the dot)
func GetFileExtension(path string) string {
	return filepath.Ext(path)
}

// GetFileNameWithoutExt returns the filename without extension
func GetFileNameWithoutExt(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	if ext == "" {
		return base
	}
	return base[:len(base)-len(ext)]
}

// ListFiles returns all files in a directory with the given extensions
func ListFiles(dir string, extensions []string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		for _, validExt := range extensions {
			if ext == validExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

// ValidateInputPath validates that an input path exists and is accessible
func ValidateInputPath(path string) error {
	if path == "" {
		return fmt.Errorf("input path cannot be empty")
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("input path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("failed to access input path %s: %w", path, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("cannot read directory %s: %w", path, err)
		}

		hasValidFiles := false
		for _, entry := range entries {
			if !entry.IsDir() {
				ext := filepath.Ext(entry.Name())
				if ext == ".svg" || ext == ".png" {
					hasValidFiles = true
					break
				}
			}
		}

		if !hasValidFiles {
			return fmt.Errorf("directory %s contains no valid SVG or PNG files", path)
		}
	} else {
		ext := filepath.Ext(path)
		if ext != ".svg" && ext != ".png" {
			return fmt.Errorf("file %s must be an SVG or PNG file", path)
		}
	}

	return nil
}

// ValidateOutputPath validates that an output path is writable
func ValidateOutputPath(path string, force bool) error {
	if path == "" {
		return fmt.Errorf("output path cannot be empty")
	}

	if FileExists(path) && !force {
		return fmt.Errorf("output file already exists: %s (use --force to overwrite)", path)
	}

	parentDir := filepath.Dir(path)
	if err := EnsureDir(parentDir); err != nil {
		return fmt.Errorf("cannot create output directory: %w", err)
	}

	tempFile, err := os.CreateTemp(parentDir, "svg2sheet_test_*")
	if err != nil {
		return fmt.Errorf("output directory is not writable: %w", err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())

	return nil
}

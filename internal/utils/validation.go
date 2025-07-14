package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thanhfphan/svg2sheet/internal/config"
)

// ValidateConfig performs comprehensive validation of the configuration
func ValidateConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Additional validation for file paths and permissions
	if err := ValidateInputPath(cfg.Input); err != nil {
		return fmt.Errorf("input validation failed: %w", err)
	}

	if err := ValidateOutputPath(cfg.Output, cfg.Force); err != nil {
		return fmt.Errorf("output validation failed: %w", err)
	}

	// Validate metadata output path if specified
	if cfg.Meta != "" {
		if err := ValidateMetadataPath(cfg.Meta, cfg.Force); err != nil {
			return fmt.Errorf("metadata path validation failed: %w", err)
		}
	}

	if cfg.IsSpritesheetMode() {
		if err := ValidateSpritesheetConfig(cfg); err != nil {
			return fmt.Errorf("spritesheet configuration validation failed: %w", err)
		}
	}

	return nil
}

// ValidateMetadataPath validates the metadata output path
func ValidateMetadataPath(path string, force bool) error {
	if path == "" {
		return fmt.Errorf("metadata path cannot be empty")
	}

	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".json" && ext != ".csv" {
		return fmt.Errorf("metadata file must have .json or .csv extension, got: %s", ext)
	}

	if FileExists(path) && !force {
		return fmt.Errorf("metadata file already exists: %s (use --force to overwrite)", path)
	}

	parentDir := filepath.Dir(path)
	if err := EnsureDir(parentDir); err != nil {
		return fmt.Errorf("cannot create metadata output directory: %w", err)
	}

	return nil
}

// ValidateSpritesheetConfig validates spritesheet-specific configuration
func ValidateSpritesheetConfig(cfg *config.Config) error {
	if cfg.TileWidth <= 0 || cfg.TileHeight <= 0 {
		return fmt.Errorf("tile dimensions must be positive: %dx%d", cfg.TileWidth, cfg.TileHeight)
	}

	// Check if tile dimensions are reasonable (not too large)
	maxTileSize := 2048
	if cfg.TileWidth > maxTileSize || cfg.TileHeight > maxTileSize {
		return fmt.Errorf("tile dimensions too large (max %d): %dx%d", maxTileSize, cfg.TileWidth, cfg.TileHeight)
	}

	if cfg.Cols <= 0 && cfg.Rows <= 0 {
		return fmt.Errorf("either cols or rows must be specified for spritesheet")
	}

	if cfg.Cols > 0 && cfg.Rows > 0 {
		return fmt.Errorf("cannot specify both cols and rows")
	}

	// Check reasonable limits
	maxGridSize := 100
	if cfg.Cols > maxGridSize || cfg.Rows > maxGridSize {
		return fmt.Errorf("grid size too large (max %d): cols=%d, rows=%d", maxGridSize, cfg.Cols, cfg.Rows)
	}

	if cfg.Padding < 0 {
		return fmt.Errorf("padding cannot be negative: %d", cfg.Padding)
	}

	maxPadding := 100
	if cfg.Padding > maxPadding {
		return fmt.Errorf("padding too large (max %d): %d", maxPadding, cfg.Padding)
	}

	return nil
}

// ValidateImageDimensions validates that image dimensions are reasonable
func ValidateImageDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("image dimensions must be positive: %dx%d", width, height)
	}

	maxDimension := 8192
	if width > maxDimension || height > maxDimension {
		return fmt.Errorf("image dimensions too large (max %d): %dx%d", maxDimension, width, height)
	}

	// Check for reasonable aspect ratio
	aspectRatio := float64(width) / float64(height)
	if aspectRatio > 10.0 || aspectRatio < 0.1 {
		return fmt.Errorf("extreme aspect ratio detected: %dx%d (ratio: %.2f)", width, height, aspectRatio)
	}

	return nil
}

// ValidateScale validates scale factor
func ValidateScale(scale float64) error {
	if scale <= 0 {
		return fmt.Errorf("scale must be positive: %f", scale)
	}

	if scale > 10.0 {
		return fmt.Errorf("scale too large (max 10.0): %f", scale)
	}

	if scale < 0.1 {
		return fmt.Errorf("scale too small (min 0.1): %f", scale)
	}

	return nil
}

// ValidateFileCount validates the number of files for processing
func ValidateFileCount(count int, mode string) error {
	if count <= 0 {
		return fmt.Errorf("no files to process")
	}

	maxFiles := 1000
	if count > maxFiles {
		return fmt.Errorf("too many files to process (max %d): %d", maxFiles, count)
	}

	// For spritesheet mode, check reasonable limits
	if mode == "spritesheet" {
		maxSprites := 256
		if count > maxSprites {
			return fmt.Errorf("too many sprites for spritesheet (max %d): %d", maxSprites, count)
		}
	}

	return nil
}

// ValidateMemoryUsage estimates and validates memory usage
func ValidateMemoryUsage(cfg *config.Config, fileCount int) error {
	// Estimate memory usage based on configuration
	tileSize := cfg.TileWidth * cfg.TileHeight * 4 // 4 bytes per pixel (RGBA)

	var estimatedMemory int64

	if cfg.IsSpritesheetMode() {
		// Memory for individual tiles + spritesheet
		tilesMemory := int64(fileCount * tileSize)

		// Calculate spritesheet dimensions
		cols := cfg.Cols
		rows := cfg.Rows
		if cols == 0 {
			cols = (fileCount + rows - 1) / rows
		}
		if rows == 0 {
			rows = (fileCount + cols - 1) / cols
		}

		spritesheetWidth := cols*cfg.TileWidth + (cols-1)*cfg.Padding
		spritesheetHeight := rows*cfg.TileHeight + (rows-1)*cfg.Padding
		spritesheetMemory := int64(spritesheetWidth * spritesheetHeight * 4)

		estimatedMemory = tilesMemory + spritesheetMemory
	} else {
		// Memory for individual conversions (assuming one at a time)
		estimatedMemory = int64(tileSize)
	}

	// Check against reasonable memory limit (500MB)
	maxMemory := int64(500 * 1024 * 1024)
	if estimatedMemory > maxMemory {
		return fmt.Errorf("estimated memory usage too high: %d MB (max 500 MB)", estimatedMemory/(1024*1024))
	}

	return nil
}

// ValidateOutputFormat validates the output file format
func ValidateOutputFormat(outputPath string) error {
	ext := strings.ToLower(filepath.Ext(outputPath))

	validExtensions := []string{".png", ".jpg", ".jpeg"}
	for _, validExt := range validExtensions {
		if ext == validExt {
			return nil
		}
	}

	return fmt.Errorf("unsupported output format: %s (supported: %v)", ext, validExtensions)
}

// ValidateSortMode validates the sort mode
func ValidateSortMode(mode string) error {
	validModes := []string{"name", "ctime", "manual"}

	for _, validMode := range validModes {
		if mode == validMode {
			return nil
		}
	}

	return fmt.Errorf("invalid sort mode: %s (valid: %v)", mode, validModes)
}

// CheckSystemRequirements checks if the system meets requirements
func CheckSystemRequirements() error {
	// Check if we can create temporary files
	tempFile, err := os.CreateTemp("", "svg2sheet_test_*")
	if err != nil {
		return fmt.Errorf("cannot create temporary files: %w", err)
	}
	tempFile.Close()
	os.Remove(tempFile.Name())

	// TODO: Implement disk space check

	return nil
}

package config

import (
	"fmt"
	"path/filepath"
)

// Config holds all configuration options for the svg2sheet tool
type Config struct {
	// Input/Output
	Input  string `json:"input"`
	Output string `json:"output"`

	// SVG Conversion
	Scale  float64 `json:"scale,omitempty"`
	Width  int     `json:"width,omitempty"`
	Height int     `json:"height,omitempty"`

	// Spritesheet Layout
	TileWidth  int `json:"tile_width,omitempty"`
	TileHeight int `json:"tile_height,omitempty"`
	Cols       int `json:"cols,omitempty"`
	Rows       int `json:"rows,omitempty"`
	Padding    int `json:"padding,omitempty"`

	// Options
	Sort      string `json:"sort,omitempty"`      // name, ctime, manual
	Meta      string `json:"meta,omitempty"`      // metadata output file
	Trim      bool   `json:"trim,omitempty"`      // trim transparent edges
	Force     bool   `json:"force,omitempty"`     // overwrite existing files
	Verbose   bool   `json:"verbose,omitempty"`   // verbose logging
	Converter string `json:"converter,omitempty"` // SVG converter backend
}

// SortMode represents different sorting options
type SortMode string

const (
	SortByName  SortMode = "name"
	SortByCTime SortMode = "ctime"
	SortManual  SortMode = "manual"
)

// ConverterType represents different SVG converter backends
type ConverterType string

const (
	ConverterOkSVG ConverterType = "oksvg"
	ConverterRod   ConverterType = "rod"
	ConverterRSVG  ConverterType = "rsvg"
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Input == "" {
		return fmt.Errorf("input path is required")
	}

	if c.Output == "" {
		return fmt.Errorf("output path is required")
	}

	// Validate scale and dimensions
	if c.Scale != 0 && (c.Width != 0 || c.Height != 0) {
		return fmt.Errorf("cannot specify both scale and width/height")
	}

	if c.Scale < 0 {
		return fmt.Errorf("scale must be positive")
	}

	if c.Width < 0 || c.Height < 0 {
		return fmt.Errorf("width and height must be positive")
	}

	// Validate spritesheet dimensions
	if c.TileWidth < 0 || c.TileHeight < 0 {
		return fmt.Errorf("tile dimensions must be positive")
	}

	if c.Cols < 0 || c.Rows < 0 {
		return fmt.Errorf("cols and rows must be positive")
	}

	if c.Cols > 0 && c.Rows > 0 {
		return fmt.Errorf("cannot specify both cols and rows")
	}

	if c.Padding < 0 {
		return fmt.Errorf("padding must be non-negative")
	}

	// Validate sort mode
	if c.Sort != "" {
		switch SortMode(c.Sort) {
		case SortByName, SortByCTime, SortManual:
			// valid
		default:
			return fmt.Errorf("invalid sort mode: %s (must be name, ctime, or manual)", c.Sort)
		}
	}

	// Validate converter type
	if c.Converter != "" {
		switch ConverterType(c.Converter) {
		case ConverterOkSVG, ConverterRod, ConverterRSVG:
			// valid
		default:
			return fmt.Errorf("invalid converter: %s (must be oksvg, rod, or rsvg)", c.Converter)
		}
	}

	return nil
}

// SetDefaults sets default values for the configuration
func (c *Config) SetDefaults() {
	if c.Scale == 0 && c.Width == 0 && c.Height == 0 {
		c.Scale = 1.0
	}

	if c.Sort == "" {
		c.Sort = string(SortByName)
	}

	if c.Converter == "" {
		c.Converter = string(ConverterOkSVG)
	}

	if c.TileWidth == 0 {
		c.TileWidth = 64
	}

	if c.TileHeight == 0 {
		c.TileHeight = 64
	}

	if c.Cols == 0 && c.Rows == 0 {
		c.Cols = 8
	}
}

// IsSpritesheetMode returns true if we're generating a spritesheet
func (c *Config) IsSpritesheetMode() bool {
	return c.TileWidth > 0 && c.TileHeight > 0 && (c.Cols > 0 || c.Rows > 0)
}

// IsSVGInput returns true if input appears to be SVG file(s)
func (c *Config) IsSVGInput() bool {
	ext := filepath.Ext(c.Input)
	return ext == ".svg"
}

// GetOutputExt returns the expected output file extension
func (c *Config) GetOutputExt() string {
	if c.Meta != "" {
		return ".json"
	}
	return ".png"
}

package svg

import (
	"fmt"
	"image"

	"github.com/thanhfphan/svg2sheet/internal/config"
)

// Converter handles SVG to PNG conversion using pluggable backends
type Converter struct {
	config   *config.Config
	backend  SVGConverter
	registry *ConverterRegistry
}

// NewConverter creates a new SVG converter with the specified backend
func NewConverter(cfg *config.Config) (*Converter, error) {
	registry := NewConverterRegistry()
	options := NewConversionOptions(cfg)

	// Create the specified converter backend
	converterType := config.ConverterType(cfg.Converter)
	backend, err := registry.Create(converterType, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s converter: %w", cfg.Converter, err)
	}

	return &Converter{
		config:   cfg,
		backend:  backend,
		registry: registry,
	}, nil
}

// ConvertFile converts a single SVG file to PNG using the configured backend
func (c *Converter) ConvertFile(inputPath, outputPath string) error {
	return c.backend.ConvertFile(inputPath, outputPath)
}

// ConvertToImage converts SVG data to an image.Image using the configured backend
func (c *Converter) ConvertToImage(svgData []byte) (image.Image, error) {
	return c.backend.ConvertToImage(svgData)
}

// GetImageDimensions returns the dimensions of an SVG file using the configured backend
func (c *Converter) GetImageDimensions(svgPath string) (int, int, error) {
	return c.backend.GetImageDimensions(svgPath)
}

// GetRegistry returns the converter registry for advanced operations
func (c *Converter) GetRegistry() *ConverterRegistry {
	return c.registry
}

// GetBackend returns the current converter backend
func (c *Converter) GetBackend() SVGConverter {
	return c.backend
}

package svg

import (
	"image"

	"github.com/thanhfphan/svg2sheet/internal/config"
)

// SVGConverter defines the interface that all SVG conversion backends must implement
type SVGConverter interface {
	// ConvertFile converts a single SVG file to PNG
	ConvertFile(inputPath, outputPath string) error

	// ConvertToImage converts SVG data to an image.Image
	ConvertToImage(svgData []byte) (image.Image, error)

	// GetImageDimensions returns the dimensions that would be used for conversion
	GetImageDimensions(svgPath string) (int, int, error)

	// IsAvailable checks if this converter is available on the system
	IsAvailable() error

	// Name returns the human-readable name of this converter
	Name() string

	// Description returns a description of this converter and its capabilities
	Description() string
}

// ConversionOptions holds options for SVG conversion
type ConversionOptions struct {
	Scale   float64
	Width   int
	Height  int
	Verbose bool
}

// NewConversionOptions creates ConversionOptions from config
func NewConversionOptions(cfg *config.Config) *ConversionOptions {
	return &ConversionOptions{
		Scale:   cfg.Scale,
		Width:   cfg.Width,
		Height:  cfg.Height,
		Verbose: cfg.Verbose,
	}
}

// CalculateDimensions determines the target width and height for conversion
// This is a common utility function that can be used by all converters
func (opts *ConversionOptions) CalculateDimensions(origWidth, origHeight float64) (int, int) {
	// If no dimensions specified, use original
	if opts.Scale == 0 && opts.Width == 0 && opts.Height == 0 {
		return int(origWidth), int(origHeight)
	}

	// If scale is specified, use it
	if opts.Scale > 0 {
		return int(origWidth * opts.Scale), int(origHeight * opts.Scale)
	}

	// If both width and height are specified, use them
	if opts.Width > 0 && opts.Height > 0 {
		return opts.Width, opts.Height
	}

	// If only width is specified, calculate height maintaining aspect ratio
	if opts.Width > 0 {
		aspectRatio := origHeight / origWidth
		return opts.Width, int(float64(opts.Width) * aspectRatio)
	}

	// If only height is specified, calculate width maintaining aspect ratio
	if opts.Height > 0 {
		aspectRatio := origWidth / origHeight
		return int(float64(opts.Height) * aspectRatio), opts.Height
	}

	// Fallback to original dimensions
	return int(origWidth), int(origHeight)
}

// ConverterRegistry manages available SVG converters
type ConverterRegistry struct {
	converters map[config.ConverterType]func(*ConversionOptions) SVGConverter
}

// NewConverterRegistry creates a new converter registry
func NewConverterRegistry() *ConverterRegistry {
	registry := &ConverterRegistry{
		converters: make(map[config.ConverterType]func(*ConversionOptions) SVGConverter),
	}

	// Register built-in converters
	registry.Register(config.ConverterOkSVG, NewOkSVGConverter)
	registry.Register(config.ConverterRod, NewRodConverter)
	registry.Register(config.ConverterRSVG, NewRSVGConverter)
	registry.Register(config.ConverterInkscape, NewInkscapeConverter)

	return registry
}

// Register adds a converter factory to the registry
func (r *ConverterRegistry) Register(converterType config.ConverterType, factory func(*ConversionOptions) SVGConverter) {
	r.converters[converterType] = factory
}

// Create creates a converter instance of the specified type
func (r *ConverterRegistry) Create(converterType config.ConverterType, opts *ConversionOptions) (SVGConverter, error) {
	factory, exists := r.converters[converterType]
	if !exists {
		return nil, &ConverterNotFoundError{ConverterType: string(converterType)}
	}

	converter := factory(opts)

	// Check if the converter is available on the system
	if err := converter.IsAvailable(); err != nil {
		return nil, &ConverterUnavailableError{
			ConverterType: string(converterType),
			Reason:        err.Error(),
		}
	}

	return converter, nil
}

// ListAvailable returns a list of available converters
func (r *ConverterRegistry) ListAvailable(opts *ConversionOptions) []config.ConverterType {
	var available []config.ConverterType

	for converterType, factory := range r.converters {
		converter := factory(opts)
		if converter.IsAvailable() == nil {
			available = append(available, converterType)
		}
	}

	return available
}

// GetConverterInfo returns information about a specific converter
func (r *ConverterRegistry) GetConverterInfo(converterType config.ConverterType, opts *ConversionOptions) (*ConverterInfo, error) {
	factory, exists := r.converters[converterType]
	if !exists {
		return nil, &ConverterNotFoundError{ConverterType: string(converterType)}
	}

	converter := factory(opts)
	return &ConverterInfo{
		Type:        converterType,
		Name:        converter.Name(),
		Description: converter.Description(),
		Available:   converter.IsAvailable() == nil,
	}, nil
}

// ConverterInfo holds information about a converter
type ConverterInfo struct {
	Type        config.ConverterType
	Name        string
	Description string
	Available   bool
}

// Error types for converter operations
type ConverterNotFoundError struct {
	ConverterType string
}

func (e *ConverterNotFoundError) Error() string {
	return "converter not found: " + e.ConverterType
}

type ConverterUnavailableError struct {
	ConverterType string
	Reason        string
}

func (e *ConverterUnavailableError) Error() string {
	return "converter " + e.ConverterType + " is not available: " + e.Reason
}

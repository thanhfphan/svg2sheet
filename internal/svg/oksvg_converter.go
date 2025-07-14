package svg

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// OkSVGConverter implements SVGConverter using the oksvg+rasterx libraries
type OkSVGConverter struct {
	options *ConversionOptions
}

// NewOkSVGConverter creates a new OkSVG-based converter
func NewOkSVGConverter(options *ConversionOptions) SVGConverter {
	return &OkSVGConverter{
		options: options,
	}
}

// Name returns the human-readable name of this converter
func (c *OkSVGConverter) Name() string {
	return "OkSVG"
}

// Description returns a description of this converter
func (c *OkSVGConverter) Description() string {
	return "Pure Go SVG renderer using oksvg+rasterx. Fast and lightweight, good for simple SVGs."
}

// IsAvailable checks if this converter is available
func (c *OkSVGConverter) IsAvailable() error {
	// OkSVG is always available since it's a pure Go library
	return nil
}

// ConvertFile converts a single SVG file to PNG
func (c *OkSVGConverter) ConvertFile(inputPath, outputPath string) error {
	if c.options.Verbose {
		fmt.Printf("Converting SVG with OkSVG: %s -> %s\n", inputPath, outputPath)
	}

	// Read SVG file
	svgData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read SVG file: %w", err)
	}

	// Convert to image
	img, err := c.ConvertToImage(svgData)
	if err != nil {
		return fmt.Errorf("failed to convert SVG to image: %w", err)
	}

	// Save as PNG
	return c.savePNG(img, outputPath)
}

// ConvertToImage converts SVG data to an image.Image
func (c *OkSVGConverter) ConvertToImage(svgData []byte) (image.Image, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG with OkSVG: %w", err)
	}

	// Calculate target dimensions
	width, height := c.calculateDimensions(icon)

	// Create and return raster image
	return c.rasterizeSVG(icon, width, height), nil
}

// GetImageDimensions returns the dimensions of an SVG file
func (c *OkSVGConverter) GetImageDimensions(svgPath string) (int, int, error) {
	svgData, err := os.ReadFile(svgPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read SVG file: %w", err)
	}

	icon, err := oksvg.ReadIconStream(bytes.NewReader(svgData))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse SVG with OkSVG: %w", err)
	}

	width, height := c.calculateDimensions(icon)
	return width, height, nil
}

// calculateDimensions determines the target width and height for the conversion
func (c *OkSVGConverter) calculateDimensions(icon *oksvg.SvgIcon) (int, int) {
	origWidth := icon.ViewBox.W
	origHeight := icon.ViewBox.H

	return c.options.CalculateDimensions(origWidth, origHeight)
}

// rasterizeSVG converts the SVG icon to a raster image
func (c *OkSVGConverter) rasterizeSVG(icon *oksvg.SvgIcon, width, height int) image.Image {
	icon.SetTarget(0, 0, float64(width), float64(height))

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	scanner := rasterx.NewScannerGV(width, height, img, img.Bounds())
	raster := rasterx.NewDasher(width, height, scanner)

	icon.Draw(raster, 1.0)

	return img
}

// savePNG saves the image as a PNG file
func (c *OkSVGConverter) savePNG(img image.Image, outputPath string) error {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

package svg

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// InkscapeConverter implements SVGConverter using the Inkscape command-line tool
type InkscapeConverter struct {
	options *ConversionOptions
}

// NewInkscapeConverter creates a new Inkscape-based converter
func NewInkscapeConverter(options *ConversionOptions) SVGConverter {
	return &InkscapeConverter{
		options: options,
	}
}

// Name returns the human-readable name of this converter
func (c *InkscapeConverter) Name() string {
	return "Inkscape"
}

// Description returns a description of this converter
func (c *InkscapeConverter) Description() string {
	return "Inkscape command-line tool. Professional-grade SVG rendering with excellent compatibility and features."
}

// IsAvailable checks if Inkscape is available on the system
func (c *InkscapeConverter) IsAvailable() error {
	cmd := exec.Command("inkscape", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("inkscape command not found - please install Inkscape (https://inkscape.org/)")
	}
	return nil
}

// ConvertFile converts a single SVG file to PNG
func (c *InkscapeConverter) ConvertFile(inputPath, outputPath string) error {
	if c.options.Verbose {
		fmt.Printf("Converting SVG with Inkscape: %s -> %s\n", inputPath, outputPath)
	}

	// Get SVG dimensions to calculate target size
	origWidth, origHeight, err := c.getSVGDimensions(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get SVG dimensions: %w", err)
	}

	// Calculate target dimensions
	width, height := c.options.CalculateDimensions(origWidth, origHeight)

	// Build inkscape command
	args := []string{
		"--export-type=png",
		"--export-width=" + strconv.Itoa(width),
		"--export-height=" + strconv.Itoa(height),
		"--export-filename=" + outputPath,
		inputPath,
	}

	cmd := exec.Command("inkscape", args...)

	if c.options.Verbose {
		fmt.Printf("Executing: inkscape %s\n", strings.Join(args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("inkscape failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ConvertToImage converts SVG data to an image.Image
func (c *InkscapeConverter) ConvertToImage(svgData []byte) (image.Image, error) {
	tmpSVG, err := os.CreateTemp("", "svg2sheet_*.svg")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary SVG file: %w", err)
	}
	defer os.Remove(tmpSVG.Name())
	defer tmpSVG.Close()

	if _, err := tmpSVG.Write(svgData); err != nil {
		return nil, fmt.Errorf("failed to write SVG data: %w", err)
	}
	tmpSVG.Close()

	tmpPNG, err := os.CreateTemp("", "svg2sheet_*.png")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary PNG file: %w", err)
	}
	defer os.Remove(tmpPNG.Name())
	tmpPNG.Close()

	// Convert using ConvertFile
	if err := c.ConvertFile(tmpSVG.Name(), tmpPNG.Name()); err != nil {
		return nil, fmt.Errorf("failed to convert SVG: %w", err)
	}

	// Read the PNG file back as image.Image
	pngFile, err := os.Open(tmpPNG.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open converted PNG: %w", err)
	}
	defer pngFile.Close()

	img, err := png.Decode(pngFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	return img, nil
}

// GetImageDimensions returns the dimensions that would be used for conversion
func (c *InkscapeConverter) GetImageDimensions(svgPath string) (int, int, error) {
	origWidth, origHeight, err := c.getSVGDimensions(svgPath)
	if err != nil {
		return 0, 0, err
	}

	width, height := c.options.CalculateDimensions(origWidth, origHeight)
	return width, height, nil
}

// getSVGDimensions gets the original dimensions of an SVG file using Inkscape
func (c *InkscapeConverter) getSVGDimensions(svgPath string) (float64, float64, error) {
	// Use inkscape to query SVG dimensions
	cmd := exec.Command("inkscape", "--query-width", "--query-height", svgPath)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query SVG dimensions: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("unexpected output from inkscape query: %s", string(output))
	}

	width, err := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse width: %w", err)
	}

	height, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse height: %w", err)
	}

	return width, height, nil
}

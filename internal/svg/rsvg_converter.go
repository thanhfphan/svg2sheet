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

// RSVGConverter implements SVGConverter using the rsvg-convert system command
type RSVGConverter struct {
	options *ConversionOptions
}

// NewRSVGConverter creates a new RSVG-based converter
func NewRSVGConverter(options *ConversionOptions) SVGConverter {
	return &RSVGConverter{
		options: options,
	}
}

// Name returns the human-readable name of this converter
func (c *RSVGConverter) Name() string {
	return "RSVG (librsvg)"
}

// Description returns a description of this converter
func (c *RSVGConverter) Description() string {
	return "System rsvg-convert command using librsvg. Excellent SVG compatibility and performance."
}

// IsAvailable checks if this converter is available
func (c *RSVGConverter) IsAvailable() error {
	// Check if rsvg-convert command is available
	_, err := exec.LookPath("rsvg-convert")
	if err != nil {
		return fmt.Errorf("rsvg-convert command not found. Please install librsvg2-bin (Ubuntu/Debian) or librsvg (macOS/Homebrew)")
	}

	// Test if the command works
	cmd := exec.Command("rsvg-convert", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("rsvg-convert command failed: %w", err)
	}

	return nil
}

// ConvertFile converts a single SVG file to PNG
func (c *RSVGConverter) ConvertFile(inputPath, outputPath string) error {
	if c.options.Verbose {
		fmt.Printf("Converting SVG with RSVG: %s -> %s\n", inputPath, outputPath)
	}

	// Get SVG dimensions to calculate target size
	origWidth, origHeight, err := c.getSVGDimensions(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get SVG dimensions: %w", err)
	}

	// Calculate target dimensions
	width, height := c.options.CalculateDimensions(origWidth, origHeight)

	// Build rsvg-convert command
	args := []string{
		"--format", "png",
		"--width", strconv.Itoa(width),
		"--height", strconv.Itoa(height),
		"--output", outputPath,
		inputPath,
	}

	cmd := exec.Command("rsvg-convert", args...)

	if c.options.Verbose {
		fmt.Printf("Executing: rsvg-convert %s\n", strings.Join(args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rsvg-convert failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// ConvertToImage converts SVG data to an image.Image
func (c *RSVGConverter) ConvertToImage(svgData []byte) (image.Image, error) {
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

// GetImageDimensions returns the dimensions of an SVG file
func (c *RSVGConverter) GetImageDimensions(svgPath string) (int, int, error) {
	origWidth, origHeight, err := c.getSVGDimensions(svgPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get SVG dimensions: %w", err)
	}

	width, height := c.options.CalculateDimensions(origWidth, origHeight)
	return width, height, nil
}

// getSVGDimensions gets the original dimensions of an SVG file using rsvg-convert
func (c *RSVGConverter) getSVGDimensions(svgPath string) (float64, float64, error) {
	// Use rsvg-convert to get SVG info
	cmd := exec.Command("rsvg-convert", "--width", "--height", svgPath)
	output, err := cmd.Output()
	if err != nil {
		// If the above fails, try a different approach
		return c.getSVGDimensionsAlternative(svgPath)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return c.getSVGDimensionsAlternative(svgPath)
	}

	width, err := strconv.ParseFloat(strings.TrimSpace(lines[0]), 64)
	if err != nil {
		return c.getSVGDimensionsAlternative(svgPath)
	}

	height, err := strconv.ParseFloat(strings.TrimSpace(lines[1]), 64)
	if err != nil {
		return c.getSVGDimensionsAlternative(svgPath)
	}

	return width, height, nil
}

// getSVGDimensionsAlternative gets SVG dimensions using a different rsvg-convert approach
func (c *RSVGConverter) getSVGDimensionsAlternative(svgPath string) (float64, float64, error) {
	// Try to get dimensions by converting to a 1x1 PNG and checking the natural size
	// This is a fallback method
	cmd := exec.Command("rsvg-convert", "--format", "png", "--width", "1", "--height", "1", svgPath)

	// Capture stderr which might contain dimension info
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 100.0, 100.0, nil // Default fallback
	}

	if err := cmd.Start(); err != nil {
		return 100.0, 100.0, nil // Default fallback
	}

	// Read stderr output
	stderrOutput := make([]byte, 1024)
	n, _ := stderr.Read(stderrOutput)
	stderr.Close()
	cmd.Wait()

	// Parse stderr for dimension information (this is implementation-specific)
	stderrStr := string(stderrOutput[:n])
	if strings.Contains(stderrStr, "x") {
		// Try to extract dimensions from error messages
		// This is a best-effort approach
	}

	// If all else fails, read the SVG file and try to parse dimensions manually
	return c.parseSVGDimensionsFromFile(svgPath)
}

// parseSVGDimensionsFromFile manually parses SVG file for dimensions
func (c *RSVGConverter) parseSVGDimensionsFromFile(svgPath string) (float64, float64, error) {
	data, err := os.ReadFile(svgPath)
	if err != nil {
		return 100.0, 100.0, nil // Default fallback
	}

	// Use the same parsing logic as the Rod converter
	return c.parseSVGDimensions(data)
}

// parseSVGDimensions extracts width and height from SVG data
func (c *RSVGConverter) parseSVGDimensions(svgData []byte) (float64, float64, error) {
	svgStr := string(svgData)

	// Default dimensions if not found
	width, height := 100.0, 100.0

	// Look for viewBox attribute first
	if viewBoxStart := strings.Index(svgStr, "viewBox=\""); viewBoxStart != -1 {
		viewBoxStart += 9 // length of "viewBox=\""
		if viewBoxEnd := strings.Index(svgStr[viewBoxStart:], "\""); viewBoxEnd != -1 {
			viewBox := svgStr[viewBoxStart : viewBoxStart+viewBoxEnd]
			parts := strings.Fields(viewBox)
			if len(parts) >= 4 {
				// viewBox format: "x y width height"
				if w, err := parseFloatRSVG(parts[2]); err == nil {
					width = w
				}
				if h, err := parseFloatRSVG(parts[3]); err == nil {
					height = h
				}
			}
		}
	}

	// Look for width and height attributes
	if widthStart := strings.Index(svgStr, "width=\""); widthStart != -1 {
		widthStart += 7 // length of "width=\""
		if widthEnd := strings.Index(svgStr[widthStart:], "\""); widthEnd != -1 {
			widthStr := svgStr[widthStart : widthStart+widthEnd]
			if w, err := parseFloatRSVG(widthStr); err == nil {
				width = w
			}
		}
	}

	if heightStart := strings.Index(svgStr, "height=\""); heightStart != -1 {
		heightStart += 8 // length of "height=\""
		if heightEnd := strings.Index(svgStr[heightStart:], "\""); heightEnd != -1 {
			heightStr := svgStr[heightStart : heightStart+heightEnd]
			if h, err := parseFloatRSVG(heightStr); err == nil {
				height = h
			}
		}
	}

	return width, height, nil
}

// parseFloatRSVG parses a float from a string, handling units
func parseFloatRSVG(s string) (float64, error) {
	// Remove common SVG units
	s = strings.TrimSuffix(s, "px")
	s = strings.TrimSuffix(s, "pt")
	s = strings.TrimSuffix(s, "em")
	s = strings.TrimSuffix(s, "rem")

	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

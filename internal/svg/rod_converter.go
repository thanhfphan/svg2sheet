package svg

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// RodConverter implements SVGConverter using Rod browser automation
type RodConverter struct {
	options *ConversionOptions
	browser *rod.Browser
}

// NewRodConverter creates a new Rod-based converter
func NewRodConverter(options *ConversionOptions) SVGConverter {
	return &RodConverter{
		options: options,
	}
}

// Name returns the human-readable name of this converter
func (c *RodConverter) Name() string {
	return "Rod Browser"
}

// Description returns a description of this converter
func (c *RodConverter) Description() string {
	return "High-quality SVG rendering using Chrome/Chromium browser automation. Excellent compatibility and quality."
}

// IsAvailable checks if this converter is available
func (c *RodConverter) IsAvailable() error {
	l := launcher.New()
	if path := l.Get(""); path == "" {
		return fmt.Errorf("Chrome/Chromium browser not found")
	}

	return nil
}

// ConvertFile converts a single SVG file to PNG
func (c *RodConverter) ConvertFile(inputPath, outputPath string) error {
	if c.options.Verbose {
		fmt.Printf("Converting SVG with Rod Browser: %s -> %s\n", inputPath, outputPath)
	}

	svgData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read SVG file: %w", err)
	}

	img, err := c.ConvertToImage(svgData)
	if err != nil {
		return fmt.Errorf("failed to convert SVG to image: %w", err)
	}

	return c.savePNG(img, outputPath)
}

// ConvertToImage converts SVG data to an image.Image
func (c *RodConverter) ConvertToImage(svgData []byte) (image.Image, error) {
	if err := c.initBrowser(); err != nil {
		return nil, fmt.Errorf("failed to initialize browser: %w", err)
	}

	origWidth, origHeight, err := c.parseSVGDimensions(svgData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SVG dimensions: %w", err)
	}

	// Calculate target dimensions
	width, height := c.options.CalculateDimensions(origWidth, origHeight)

	html := c.createHTMLWithSVG(string(svgData), width, height)

	page := c.browser.MustPage()
	defer page.MustClose()

	page.MustSetViewport(width, height, 1, false)
	page.MustNavigate("data:text/html;charset=utf-8," + html)
	page.MustWaitLoad()

	screenshot, err := page.Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatPng,
		Quality: nil, // PNG doesn't use quality
	})
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}

	img, err := png.Decode(strings.NewReader(string(screenshot)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot PNG: %w", err)
	}

	return img, nil
}

// GetImageDimensions returns the dimensions of an SVG file
func (c *RodConverter) GetImageDimensions(svgPath string) (int, int, error) {
	svgData, err := os.ReadFile(svgPath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read SVG file: %w", err)
	}

	origWidth, origHeight, err := c.parseSVGDimensions(svgData)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse SVG dimensions: %w", err)
	}

	width, height := c.options.CalculateDimensions(origWidth, origHeight)
	return width, height, nil
}

// initBrowser initializes the browser instance if not already done
func (c *RodConverter) initBrowser() error {
	if c.browser != nil {
		return nil
	}

	launcher := launcher.New().
		Headless(true).
		NoSandbox(true).
		Set("disable-gpu").
		Set("disable-dev-shm-usage")

	url, err := launcher.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	c.browser = browser
	return nil
}

// parseSVGDimensions extracts width and height from SVG data
func (c *RodConverter) parseSVGDimensions(svgData []byte) (float64, float64, error) {
	// TODO: Improve SVG dimension parsing
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
				if w, err := parseFloatRod(parts[2]); err == nil {
					width = w
				}
				if h, err := parseFloatRod(parts[3]); err == nil {
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
			if w, err := parseFloatRod(widthStr); err == nil {
				width = w
			}
		}
	}

	if heightStart := strings.Index(svgStr, "height=\""); heightStart != -1 {
		heightStart += 8 // length of "height=\""
		if heightEnd := strings.Index(svgStr[heightStart:], "\""); heightEnd != -1 {
			heightStr := svgStr[heightStart : heightStart+heightEnd]
			if h, err := parseFloatRod(heightStr); err == nil {
				height = h
			}
		}
	}

	return width, height, nil
}

// parseFloatRod parses a float from a string, handling units
func parseFloatRod(s string) (float64, error) {
	// Remove common SVG units
	s = strings.TrimSuffix(s, "px")
	s = strings.TrimSuffix(s, "pt")
	s = strings.TrimSuffix(s, "em")
	s = strings.TrimSuffix(s, "rem")

	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}

// createHTMLWithSVG creates an HTML page containing the SVG
func (c *RodConverter) createHTMLWithSVG(svgContent string, width, height int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        body { margin: 0; padding: 0; background: transparent; }
        svg { display: block; width: %dpx; height: %dpx; }
    </style>
</head>
<body>
    %s
</body>
</html>`, width, height, svgContent)
}

// savePNG saves the image as a PNG file
func (c *RodConverter) savePNG(img image.Image, outputPath string) error {
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

// Close closes the browser instance
func (c *RodConverter) Close() error {
	if c.browser != nil {
		return c.browser.Close()
	}
	return nil
}

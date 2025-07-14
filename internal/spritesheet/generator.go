package spritesheet

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/thanhfphan/svg2sheet/internal/config"
	"github.com/thanhfphan/svg2sheet/internal/metadata"
	"github.com/thanhfphan/svg2sheet/internal/utils"
)

// Generator handles spritesheet generation
type Generator struct {
	config *config.Config
}

// NewGenerator creates a new spritesheet generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
	}
}

// Generate creates a spritesheet from the given PNG files
func (g *Generator) Generate(fileMappings []utils.FileMapping, outputPath string) (*metadata.SpritesheetMetadata, error) {
	if len(fileMappings) == 0 {
		return nil, fmt.Errorf("no PNG files provided")
	}

	if g.config.Verbose {
		fmt.Printf("Generating spritesheet from %d files\n", len(fileMappings))
	}

	// Load and process images
	images, err := g.loadImages(fileMappings)
	if err != nil {
		return nil, fmt.Errorf("failed to load images: %w", err)
	}

	// Calculate layout
	layout := g.calculateLayout(len(images))

	// Create spritesheet
	spritesheet, metadata, err := g.createSpritesheet(images, layout)
	if err != nil {
		return nil, fmt.Errorf("failed to create spritesheet: %w", err)
	}

	// Save spritesheet
	if err := g.saveSpritesheet(spritesheet, outputPath); err != nil {
		return nil, fmt.Errorf("failed to save spritesheet: %w", err)
	}

	return metadata, nil
}

// ImageInfo holds information about a loaded image
type ImageInfo struct {
	Image        image.Image
	Filename     string
	OriginalPath string
	Width        int
	Height       int
}

// Layout holds spritesheet layout information
type Layout struct {
	Cols       int
	Rows       int
	TileWidth  int
	TileHeight int
	Padding    int
	Width      int
	Height     int
}

// loadImages loads all PNG files and returns image information
func (g *Generator) loadImages(fileMappings []utils.FileMapping) ([]*ImageInfo, error) {
	var images []*ImageInfo

	for _, mapping := range fileMappings {
		if g.config.Verbose {
			fmt.Printf("Loading image: %s\n", mapping.PNGPath)
		}

		img, err := g.loadImage(mapping.PNGPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", mapping.PNGPath, err)
		}

		// Process image (resize, trim if needed)
		processedImg := g.processImage(img)

		// Use original filename for sprite naming
		originalName := filepath.Base(mapping.OriginalPath)
		if ext := filepath.Ext(originalName); ext != "" {
			originalName = originalName[:len(originalName)-len(ext)]
		}

		images = append(images, &ImageInfo{
			Image:        processedImg,
			Filename:     originalName,
			OriginalPath: mapping.OriginalPath,
			Width:        processedImg.Bounds().Dx(),
			Height:       processedImg.Bounds().Dy(),
		})
	}

	return images, nil
}

// loadImage loads a single PNG file
func (g *Generator) loadImage(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// processImage processes an image (resize, trim, etc.)
func (g *Generator) processImage(img image.Image) image.Image {
	if g.config.Trim {
		img = utils.TrimTransparent(img)
	}

	// Resize to tile dimensions if they don't match
	bounds := img.Bounds()
	if bounds.Dx() != g.config.TileWidth || bounds.Dy() != g.config.TileHeight {
		img = utils.ResizeImage(img, g.config.TileWidth, g.config.TileHeight)
	}

	return img
}

// calculateLayout determines the spritesheet layout
func (g *Generator) calculateLayout(imageCount int) *Layout {
	var cols, rows int

	if g.config.Cols > 0 {
		cols = g.config.Cols
		rows = int(math.Ceil(float64(imageCount) / float64(cols)))
	} else if g.config.Rows > 0 {
		rows = g.config.Rows
		cols = int(math.Ceil(float64(imageCount) / float64(rows)))
	} else {
		// Default: try to make it roughly square
		cols = int(math.Ceil(math.Sqrt(float64(imageCount))))
		rows = int(math.Ceil(float64(imageCount) / float64(cols)))
	}

	width := cols*g.config.TileWidth + (cols-1)*g.config.Padding
	height := rows*g.config.TileHeight + (rows-1)*g.config.Padding

	return &Layout{
		Cols:       cols,
		Rows:       rows,
		TileWidth:  g.config.TileWidth,
		TileHeight: g.config.TileHeight,
		Padding:    g.config.Padding,
		Width:      width,
		Height:     height,
	}
}

// createSpritesheet creates the actual spritesheet image and metadata
func (g *Generator) createSpritesheet(images []*ImageInfo, layout *Layout) (image.Image, *metadata.SpritesheetMetadata, error) {
	spritesheet := image.NewRGBA(image.Rect(0, 0, layout.Width, layout.Height))

	// Create metadata
	meta := &metadata.SpritesheetMetadata{
		Width:      layout.Width,
		Height:     layout.Height,
		TileWidth:  layout.TileWidth,
		TileHeight: layout.TileHeight,
		Cols:       layout.Cols,
		Rows:       layout.Rows,
		Padding:    layout.Padding,
		Sprites:    make([]metadata.SpriteInfo, 0, len(images)),
	}

	// Place images on the spritesheet
	for i, imgInfo := range images {
		col := i % layout.Cols
		row := i / layout.Cols

		x := col * (layout.TileWidth + layout.Padding)
		y := row * (layout.TileHeight + layout.Padding)

		destRect := image.Rect(x, y, x+layout.TileWidth, y+layout.TileHeight)
		draw.Draw(spritesheet, destRect, imgInfo.Image, image.Point{}, draw.Over)

		sprite := metadata.SpriteInfo{
			Name:   g.getSpriteName(imgInfo.Filename),
			X:      x,
			Y:      y,
			Width:  layout.TileWidth,
			Height: layout.TileHeight,
			Index:  i,
		}
		meta.Sprites = append(meta.Sprites, sprite)

		if g.config.Verbose {
			fmt.Printf("Placed sprite %d: %s at (%d, %d)\n", i, sprite.Name, x, y)
		}
	}

	return spritesheet, meta, nil
}

// getSpriteName extracts the sprite name from filename (already processed in loadImages)
func (g *Generator) getSpriteName(filename string) string {
	return filename
}

// saveSpritesheet saves the spritesheet to a file
func (g *Generator) saveSpritesheet(img image.Image, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

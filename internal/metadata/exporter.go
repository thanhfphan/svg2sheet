package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thanhfphan/svg2sheet/internal/config"
)

// Exporter handles metadata export
type Exporter struct {
	config *config.Config
}

// NewExporter creates a new metadata exporter
func NewExporter(cfg *config.Config) *Exporter {
	return &Exporter{
		config: cfg,
	}
}

// SpritesheetMetadata contains information about the generated spritesheet
type SpritesheetMetadata struct {
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	TileWidth  int          `json:"tile_width"`
	TileHeight int          `json:"tile_height"`
	Cols       int          `json:"cols"`
	Rows       int          `json:"rows"`
	Padding    int          `json:"padding"`
	Sprites    []SpriteInfo `json:"sprites"`
}

// SpriteInfo contains information about individual sprites
type SpriteInfo struct {
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Index  int    `json:"index"`
}

// Export saves the metadata to a JSON file
func (e *Exporter) Export(metadata *SpritesheetMetadata, outputPath string) error {
	if e.config.Verbose {
		fmt.Printf("Exporting metadata to: %s\n", outputPath)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	if e.config.Verbose {
		fmt.Printf("Metadata exported successfully with %d sprites\n", len(metadata.Sprites))
	}

	return nil
}

// ExportCSV exports metadata in CSV format (alternative format)
func (e *Exporter) ExportCSV(metadata *SpritesheetMetadata, outputPath string) error {
	if e.config.Verbose {
		fmt.Printf("Exporting metadata to CSV: %s\n", outputPath)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create CSV content
	csvContent := "name,x,y,width,height,index\n"
	for _, sprite := range metadata.Sprites {
		csvContent += fmt.Sprintf("%s,%d,%d,%d,%d,%d\n",
			sprite.Name, sprite.X, sprite.Y, sprite.Width, sprite.Height, sprite.Index)
	}

	if err := os.WriteFile(outputPath, []byte(csvContent), 0644); err != nil {
		return fmt.Errorf("failed to write CSV file: %w", err)
	}

	return nil
}

// LoadMetadata loads metadata from a JSON file
func (e *Exporter) LoadMetadata(inputPath string) (*SpritesheetMetadata, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata SpritesheetMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return &metadata, nil
}

// ValidateMetadata validates the metadata structure
func (e *Exporter) ValidateMetadata(metadata *SpritesheetMetadata) error {
	if metadata.Width <= 0 || metadata.Height <= 0 {
		return fmt.Errorf("invalid spritesheet dimensions: %dx%d", metadata.Width, metadata.Height)
	}

	if metadata.TileWidth <= 0 || metadata.TileHeight <= 0 {
		return fmt.Errorf("invalid tile dimensions: %dx%d", metadata.TileWidth, metadata.TileHeight)
	}

	if metadata.Cols <= 0 || metadata.Rows <= 0 {
		return fmt.Errorf("invalid grid dimensions: %dx%d", metadata.Cols, metadata.Rows)
	}

	if len(metadata.Sprites) == 0 {
		return fmt.Errorf("no sprites in metadata")
	}

	// Validate individual sprites
	for i, sprite := range metadata.Sprites {
		if sprite.Name == "" {
			return fmt.Errorf("sprite %d has empty name", i)
		}

		if sprite.X < 0 || sprite.Y < 0 {
			return fmt.Errorf("sprite %s has invalid position: (%d, %d)", sprite.Name, sprite.X, sprite.Y)
		}

		if sprite.Width <= 0 || sprite.Height <= 0 {
			return fmt.Errorf("sprite %s has invalid dimensions: %dx%d", sprite.Name, sprite.Width, sprite.Height)
		}

		// Check if sprite is within spritesheet bounds
		if sprite.X+sprite.Width > metadata.Width || sprite.Y+sprite.Height > metadata.Height {
			return fmt.Errorf("sprite %s extends beyond spritesheet bounds", sprite.Name)
		}
	}

	return nil
}

// GetSpriteByName finds a sprite by name in the metadata
func (e *Exporter) GetSpriteByName(metadata *SpritesheetMetadata, name string) (*SpriteInfo, error) {
	for _, sprite := range metadata.Sprites {
		if sprite.Name == name {
			return &sprite, nil
		}
	}
	return nil, fmt.Errorf("sprite not found: %s", name)
}

// GetSpriteByIndex finds a sprite by index in the metadata
func (e *Exporter) GetSpriteByIndex(metadata *SpritesheetMetadata, index int) (*SpriteInfo, error) {
	if index < 0 || index >= len(metadata.Sprites) {
		return nil, fmt.Errorf("sprite index out of range: %d", index)
	}
	return &metadata.Sprites[index], nil
}

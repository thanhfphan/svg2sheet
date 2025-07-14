package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thanhfphan/svg2sheet/internal/config"
	"github.com/thanhfphan/svg2sheet/internal/metadata"
	"github.com/thanhfphan/svg2sheet/internal/spritesheet"
	"github.com/thanhfphan/svg2sheet/internal/svg"
	"github.com/thanhfphan/svg2sheet/internal/utils"
)

// Processor handles the main processing logic
type Processor struct {
	config    *config.Config
	converter *svg.Converter
	generator *spritesheet.Generator
	exporter  *metadata.Exporter
}

// NewProcessor creates a new processor instance
func NewProcessor(cfg *config.Config) (*Processor, error) {
	converter, err := svg.NewConverter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create SVG converter: %w", err)
	}

	return &Processor{
		config:    cfg,
		converter: converter,
		generator: spritesheet.NewGenerator(cfg),
		exporter:  metadata.NewExporter(cfg),
	}, nil
}

// Process executes the main processing logic based on configuration
func (p *Processor) Process() error {
	inputInfo, err := os.Stat(p.config.Input)
	if err != nil {
		return fmt.Errorf("failed to stat input: %w", err)
	}

	if inputInfo.IsDir() {
		return p.processDirectory()
	} else {
		return p.processFile()
	}
}

// processFile handles single file processing
func (p *Processor) processFile() error {
	if p.config.Verbose {
		fmt.Printf("Processing single file: %s\n", p.config.Input)
	}

	if !p.config.IsSVGInput() {
		return fmt.Errorf("single file input must be an SVG file")
	}

	return p.converter.ConvertFile(p.config.Input, p.config.Output)
}

// processDirectory handles directory processing
func (p *Processor) processDirectory() error {
	if p.config.Verbose {
		fmt.Printf("Processing directory: %s\n", p.config.Input)
	}

	files, err := p.getInputFiles()
	if err != nil {
		return fmt.Errorf("failed to get input files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no valid input files found in directory")
	}

	if p.config.Verbose {
		fmt.Printf("Found %d files to process\n", len(files))
	}

	sortedFiles, err := utils.SortFiles(files, config.SortMode(p.config.Sort))
	if err != nil {
		return fmt.Errorf("failed to sort files: %w", err)
	}

	if p.config.IsSpritesheetMode() {
		return p.generateSpritesheet(sortedFiles)
	} else {
		return p.convertFiles(sortedFiles)
	}
}

// getInputFiles returns a list of valid input files from the input directory
func (p *Processor) getInputFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(p.config.Input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".svg" || ext == ".png" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// convertFiles converts multiple files individually
func (p *Processor) convertFiles(files []string) error {
	if err := os.MkdirAll(p.config.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for i, file := range files {
		if p.config.Verbose {
			fmt.Printf("Converting file %d/%d: %s\n", i+1, len(files), file)
		}

		baseName := filepath.Base(file)
		nameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]
		outputFile := filepath.Join(p.config.Output, nameWithoutExt+".png")

		ext := filepath.Ext(file)
		if ext == ".svg" {
			if err := p.converter.ConvertFile(file, outputFile); err != nil {
				return fmt.Errorf("failed to convert %s: %w", file, err)
			}
		} else if ext == ".png" {
			if err := utils.CopyFile(file, outputFile); err != nil {
				return fmt.Errorf("failed to copy %s: %w", file, err)
			}
		}
	}

	return nil
}

// generateSpritesheet creates a spritesheet from the input files
func (p *Processor) generateSpritesheet(files []string) error {
	if p.config.Verbose {
		fmt.Printf("Generating spritesheet with %d files\n", len(files))
	}

	// Convert SVG files to PNG if needed (in-memory or temporary files)
	fileMappings, cleanup, err := p.preparePNGFiles(files)
	if err != nil {
		return fmt.Errorf("failed to prepare PNG files: %w", err)
	}
	defer cleanup()

	// Generate the spritesheet
	metadata, err := p.generator.Generate(fileMappings, p.config.Output)
	if err != nil {
		return fmt.Errorf("failed to generate spritesheet: %w", err)
	}

	// Export metadata if requested
	if p.config.Meta != "" {
		if err := p.exporter.Export(metadata, p.config.Meta); err != nil {
			return fmt.Errorf("failed to export metadata: %w", err)
		}
	}

	if p.config.Verbose {
		fmt.Printf("Spritesheet generated successfully: %s\n", p.config.Output)
		if p.config.Meta != "" {
			fmt.Printf("Metadata exported: %s\n", p.config.Meta)
		}
	}

	return nil
}

// preparePNGFiles converts SVG files to PNG and returns a list of PNG files with mappings
func (p *Processor) preparePNGFiles(files []string) ([]utils.FileMapping, func(), error) {
	var fileMappings []utils.FileMapping
	var tempFiles []string

	cleanup := func() {
		for _, tempFile := range tempFiles {
			os.Remove(tempFile)
		}
	}

	for _, file := range files {
		ext := filepath.Ext(file)
		if ext == ".png" {
			fileMappings = append(fileMappings, utils.FileMapping{
				PNGPath:      file,
				OriginalPath: file,
				IsTemporary:  false,
			})
		} else if ext == ".svg" {
			// Create temporary PNG file
			tempFile, err := utils.CreateTempFile(".png")
			if err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("failed to create temp file: %w", err)
			}

			if err := p.converter.ConvertFile(file, tempFile); err != nil {
				cleanup()
				return nil, nil, fmt.Errorf("failed to convert %s: %w", file, err)
			}

			fileMappings = append(fileMappings, utils.FileMapping{
				PNGPath:      tempFile,
				OriginalPath: file,
				IsTemporary:  true,
			})
			tempFiles = append(tempFiles, tempFile)
		}
	}

	return fileMappings, cleanup, nil
}

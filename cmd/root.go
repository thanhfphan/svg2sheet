package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thanhfphan/svg2sheet/internal/config"
)

var cfg config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "svg2sheet",
	Short: "Convert SVG files to PNG and generate spritesheets",
	Long: `svg2sheet is a command-line tool that converts SVG files into PNGs 
and can generate spritesheets from a folder of SVG or PNG files.

Examples:
  # Convert single SVG to PNG
  svg2sheet --input icon.svg --output icon.png --scale 2.0

  # Convert folder of SVGs to PNGs
  svg2sheet --input ./svg-folder --output ./png-folder

  # Generate spritesheet from SVGs
  svg2sheet --input ./svg --output sheet.png --tile-width 64 --tile-height 64 --cols 5

  # Generate spritesheet with metadata
  svg2sheet --input ./svg --output sheet.png --tile-width 64 --tile-height 64 --cols 5 --meta sheet.json

  # Use different converter backends
  svg2sheet --input icon.svg --output icon.png --converter rod
  svg2sheet --input icon.svg --output icon.png --converter rsvg --scale 2.0

  # List available converters
  svg2sheet converters`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSvg2Sheet()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Input/Output flags
	rootCmd.Flags().StringVarP(&cfg.Input, "input", "i", "", "Input SVG file or directory (required)")
	rootCmd.Flags().StringVarP(&cfg.Output, "output", "o", "", "Output PNG file or directory (required)")
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")

	// SVG conversion flags
	rootCmd.Flags().Float64Var(&cfg.Scale, "scale", 0, "Scale factor for SVG conversion (e.g., 2.0)")
	rootCmd.Flags().IntVar(&cfg.Width, "width", 0, "Target width for SVG conversion")
	rootCmd.Flags().IntVar(&cfg.Height, "height", 0, "Target height for SVG conversion")

	// Spritesheet layout flags
	rootCmd.Flags().IntVar(&cfg.TileWidth, "tile-width", 0, "Width of each tile in spritesheet")
	rootCmd.Flags().IntVar(&cfg.TileHeight, "tile-height", 0, "Height of each tile in spritesheet")
	rootCmd.Flags().IntVar(&cfg.Cols, "cols", 0, "Number of columns in spritesheet")
	rootCmd.Flags().IntVar(&cfg.Rows, "rows", 0, "Number of rows in spritesheet")
	rootCmd.Flags().IntVar(&cfg.Padding, "padding", 0, "Padding between tiles in pixels")

	// Options flags
	rootCmd.Flags().StringVar(&cfg.Sort, "sort", "", "Sort mode: name, ctime, or manual")
	rootCmd.Flags().StringVar(&cfg.Meta, "meta", "", "Output metadata JSON file")
	rootCmd.Flags().BoolVar(&cfg.Trim, "trim", false, "Trim transparent edges from images")
	rootCmd.Flags().BoolVar(&cfg.Force, "force", false, "Overwrite existing output files")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVar(&cfg.Converter, "converter", "", "SVG converter backend: oksvg, rod, or rsvg (default: oksvg)")
}

func runSvg2Sheet() error {
	// Set defaults and validate configuration
	cfg.SetDefaults()
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("Configuration: %+v\n", cfg)
	}

	if _, err := os.Stat(cfg.Input); os.IsNotExist(err) {
		return fmt.Errorf("input path does not exist: %s", cfg.Input)
	}

	if _, err := os.Stat(cfg.Output); err == nil && !cfg.Force {
		return fmt.Errorf("output file already exists: %s (use --force to overwrite)", cfg.Output)
	}

	return executeOperation()
}

func executeOperation() error {
	processor, err := NewProcessor(&cfg)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}
	return processor.Process()
}

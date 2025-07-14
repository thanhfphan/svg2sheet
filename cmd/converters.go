package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/thanhfphan/svg2sheet/internal/config"
	"github.com/thanhfphan/svg2sheet/internal/svg"
)

// convertersCmd represents the converters command
var convertersCmd = &cobra.Command{
	Use:   "converters",
	Short: "List available SVG converter backends",
	Long: `List all available SVG converter backends and their status.

This command shows which converters are available on your system and provides
information about their capabilities and requirements.

Examples:
  # List all converters
  svg2sheet converters

  # List converters with verbose output
  svg2sheet converters --verbose`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConvertersList()
	},
}

func init() {
	rootCmd.AddCommand(convertersCmd)
	convertersCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Show detailed converter information")
}

func runConvertersList() error {
	// Create a temporary config for testing converters
	tempConfig := &config.Config{
		Converter: "oksvg", // Default for testing
		Verbose:   cfg.Verbose,
	}
	tempConfig.SetDefaults()

	// Create registry and options
	registry := svg.NewConverterRegistry()
	options := svg.NewConversionOptions(tempConfig)

	// Get all converter types
	converterTypes := []config.ConverterType{
		config.ConverterOkSVG,
		config.ConverterRod,
		config.ConverterRSVG,
	}

	fmt.Println("SVG Converter Backends")
	fmt.Println("======================")
	fmt.Println()

	if cfg.Verbose {
		// Detailed output
		for _, converterType := range converterTypes {
			info, err := registry.GetConverterInfo(converterType, options)
			if err != nil {
				fmt.Printf("❌ %s: Error getting info - %v\n", converterType, err)
				continue
			}

			status := "❌ Not Available"
			if info.Available {
				status = "✅ Available"
			}

			fmt.Printf("%s %s (%s)\n", status, info.Name, converterType)
			fmt.Printf("   %s\n", info.Description)

			if !info.Available {
				// Try to get more specific error information
				_, err := registry.Create(converterType, options)
				if err != nil {
					fmt.Printf("   Error: %v\n", err)
				}
			}
			fmt.Println()
		}
	} else {
		// Compact table output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "STATUS\tNAME\tTYPE\tDESCRIPTION")
		fmt.Fprintln(w, "------\t----\t----\t-----------")

		for _, converterType := range converterTypes {
			info, err := registry.GetConverterInfo(converterType, options)
			if err != nil {
				fmt.Fprintf(w, "❌\t%s\t%s\tError: %v\n", "Unknown", converterType, err)
				continue
			}

			status := "❌"
			if info.Available {
				status = "✅"
			}

			// Truncate description for table format
			description := info.Description
			if len(description) > 50 {
				description = description[:47] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", status, info.Name, converterType, description)
		}

		w.Flush()
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  Use --converter flag to specify which backend to use")
	fmt.Println("  Example: svg2sheet --converter rod --input icon.svg --output icon.png")
	fmt.Println()

	// Show available converters
	available := registry.ListAvailable(options)
	if len(available) > 0 {
		fmt.Printf("Available converters: %v\n", available)
		fmt.Printf("Default converter: %s\n", config.ConverterOkSVG)
	} else {
		fmt.Println("⚠️  No converters are available on this system!")
		fmt.Println()
		fmt.Println("Installation instructions:")
		fmt.Println("- oksvg: Built-in (should always be available)")
		fmt.Println("- rod: Requires Chrome/Chromium browser")
		fmt.Println("- rsvg: Requires rsvg-convert command (install librsvg2-bin)")
	}

	return nil
}

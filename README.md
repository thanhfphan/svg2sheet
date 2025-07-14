# svg2sheet

A powerful command-line tool to convert SVG files into PNGs and generate spritesheets from folders of SVG or PNG files.

## Installation

### From Source
```bash
git clone https://github.com/thanhfphan/svg2sheet.git
cd svg2sheet
make build
# Binary will be in build/svg2sheet
```

### Using Go Install
```bash
go install github.com/thanhfphan/svg2sheet@latest
```

## Usage

### Basic Examples

#### Convert Single SVG to PNG
```bash
# Basic conversion
svg2sheet --input icon.svg --output icon.png

# With scaling
svg2sheet --input icon.svg --output icon.png --scale 2.0

# With specific dimensions
svg2sheet --input icon.svg --output icon.png --width 64 --height 64
```

#### Convert Folder of SVGs to PNGs
```bash
svg2sheet --input ./svg-folder --output ./png-folder --scale 1.5
```

#### Generate Spritesheet
```bash
# Basic spritesheet
svg2sheet --input ./icons --output spritesheet.png --tile-width 64 --tile-height 64 --cols 8

# With metadata export
svg2sheet --input ./icons --output spritesheet.png --tile-width 64 --tile-height 64 --cols 8 --meta spritesheet.json

# With padding and sorting
svg2sheet --input ./icons --output spritesheet.png --tile-width 64 --tile-height 64 --cols 8 --padding 2 --sort name
```

### Advanced Examples

#### Using Different Converter Backends
```bash
# Use OkSVG (default) - fast, pure Go
svg2sheet --input icon.svg --output icon.png --converter oksvg

# Use Rod browser - high quality, complex SVG support
svg2sheet --input complex.svg --output complex.png --converter rod --scale 2.0

# Use RSVG - excellent compatibility, system dependency
svg2sheet --input ./svg-folder --output ./png-folder --converter rsvg

# Use Inkscape - professional-grade rendering, system dependency
svg2sheet --input complex.svg --output complex.png --converter inkscape --scale 2.0

# Generate spritesheet with Rod converter for best quality
svg2sheet --input ./svg --output sheet.png --tile-width 64 --tile-height 64 --cols 5 --converter rod --meta sheet.json
```

#### Complete Spritesheet Generation
```bash
svg2sheet \
  --input ./svg \
  --output sheet.png \
  --tile-width 64 \
  --tile-height 64 \
  --cols 5 \
  --meta sheet.json \
  --scale 1.5 \
  --padding 4 \
  --sort name \
  --trim \
  --verbose \
  --force \
  --converter inkscape
```

## Command Line Options

### Required Flags
- `--input, -i`: Input SVG file or directory (required)
- `--output, -o`: Output PNG file or directory (required)

### SVG Conversion Options
- `--scale`: Scale factor for SVG conversion (e.g., 2.0)
- `--width`: Target width for SVG conversion
- `--height`: Target height for SVG conversion

### Spritesheet Layout Options
- `--tile-width`: Width of each tile in spritesheet
- `--tile-height`: Height of each tile in spritesheet
- `--cols`: Number of columns in spritesheet
- `--rows`: Number of rows in spritesheet (alternative to --cols)
- `--padding`: Padding between tiles in pixels

### Processing Options
- `--sort`: Sort mode: `name`, `ctime`, or `manual`
- `--trim`: Trim transparent edges from images
- `--meta`: Output metadata JSON file

### Converter Options
- `--converter`: SVG converter backend: `oksvg`, `rod`, `rsvg`, or `inkscape` (default: oksvg)

### General Options
- `--force`: Overwrite existing output files
- `--verbose, -v`: Enable verbose logging
- `--help, -h`: Show help message

## Metadata Format

When using `--meta`, svg2sheet exports a JSON file with the following structure:

```json
{
  "width": 192,
  "height": 128,
  "tile_width": 64,
  "tile_height": 64,
  "cols": 3,
  "rows": 2,
  "padding": 0,
  "sprites": [
    {
      "name": "icon1",
      "x": 0,
      "y": 0,
      "width": 64,
      "height": 64,
      "index": 0
    }
  ]
}
```

## SVG Converter Backends

svg2sheet supports multiple SVG rendering backends, each with different strengths:

### Available Converters

#### OkSVG (Default)
- **Type**: Pure Go library
- **Pros**: Fast, lightweight, no external dependencies, always available
- **Cons**: Limited SVG feature support compared to browser engines
- **Best for**: Simple SVGs, batch processing, deployment environments
- **Usage**: `--converter oksvg` (default)

#### Rod Browser
- **Type**: Chrome/Chromium browser automation
- **Pros**: Excellent SVG compatibility, high-quality rendering, supports complex SVGs
- **Cons**: Requires Chrome/Chromium browser, slower than pure Go solutions
- **Best for**: Complex SVGs, high-quality output, development environments
- **Usage**: `--converter rod`
- **Requirements**: Chrome or Chromium browser installed

#### RSVG (librsvg)
- **Type**: System command wrapper for rsvg-convert
- **Pros**: Excellent SVG compatibility, fast, mature library
- **Cons**: Requires system dependency installation
- **Best for**: Production environments with librsvg installed, complex SVGs
- **Usage**: `--converter rsvg`
- **Requirements**: `rsvg-convert` command (install `librsvg2-bin` on Ubuntu/Debian)

#### Inkscape
- **Type**: System command wrapper for Inkscape CLI
- **Pros**: Professional-grade SVG rendering, excellent compatibility, extensive feature support
- **Cons**: Requires Inkscape installation, larger dependency footprint
- **Best for**: Complex SVGs, professional workflows, maximum SVG feature compatibility
- **Usage**: `--converter inkscape`
- **Requirements**: Inkscape installed (download from [https://inkscape.org/](https://inkscape.org/))

### Checking Available Converters

```bash
# List all converters and their availability
svg2sheet converters

# Show detailed converter information
svg2sheet converters --verbose
```

### Installation Instructions

#### Installing Chrome/Chromium (for Rod converter)
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS (Homebrew)
brew install --cask google-chrome

# Or use Chromium
brew install --cask chromium
```

#### Installing librsvg (for RSVG converter)
```bash
# Ubuntu/Debian
sudo apt-get install librsvg2-bin

# macOS (Homebrew)
brew install librsvg

# Verify installation
rsvg-convert --version
```

#### Installing Inkscape (for Inkscape converter)
```bash
# Ubuntu/Debian
sudo apt-get install inkscape

# macOS (Homebrew)
brew install --cask inkscape

# Windows
# Download from https://inkscape.org/release/

# Verify installation
inkscape --version
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

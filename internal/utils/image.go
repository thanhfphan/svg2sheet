package utils

import (
	"image"
	"image/color"
	"image/draw"
)

// TrimTransparent removes transparent edges from an image
func TrimTransparent(img image.Image) image.Image {
	bounds := img.Bounds()

	// Find the actual content bounds by scanning for non-transparent pixels
	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y

	found := false

	// Scan the image to find non-transparent pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 { // Non-transparent pixel found
				if !found {
					minX, minY = x, y
					maxX, maxY = x, y
					found = true
				} else {
					if x < minX {
						minX = x
					}
					if x > maxX {
						maxX = x
					}
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		}
	}

	// If no non-transparent pixels found, return a 1x1 transparent image
	if !found {
		result := image.NewRGBA(image.Rect(0, 0, 1, 1))
		return result
	}

	// Create new image with trimmed bounds
	trimmedBounds := image.Rect(0, 0, maxX-minX+1, maxY-minY+1)
	result := image.NewRGBA(trimmedBounds)

	// Copy the non-transparent region
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			result.Set(x-minX, y-minY, img.At(x, y))
		}
	}

	return result
}

// ResizeImage resizes an image to the specified dimensions using nearest neighbor
func ResizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// If already the correct size, return as-is
	if srcWidth == width && srcHeight == height {
		return img
	}

	result := image.NewRGBA(image.Rect(0, 0, width, height))

	// Calculate scaling factors
	scaleX := float64(srcWidth) / float64(width)
	scaleY := float64(srcHeight) / float64(height)

	// Resize using nearest neighbor sampling
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)

			// Clamp to source bounds
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}

			result.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return result
}

// ResizeImageWithAspectRatio resizes an image while maintaining aspect ratio
func ResizeImageWithAspectRatio(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()

	// Calculate scaling factor to fit within max dimensions
	scaleX := float64(maxWidth) / float64(srcWidth)
	scaleY := float64(maxHeight) / float64(srcHeight)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}

	// Calculate new dimensions
	newWidth := int(float64(srcWidth) * scale)
	newHeight := int(float64(srcHeight) * scale)

	return ResizeImage(img, newWidth, newHeight)
}

// CenterImage centers an image within a canvas of the specified size
func CenterImage(img image.Image, canvasWidth, canvasHeight int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	canvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))

	// Calculate center position
	x := (canvasWidth - imgWidth) / 2
	y := (canvasHeight - imgHeight) / 2

	// Draw image centered on canvas
	destRect := image.Rect(x, y, x+imgWidth, y+imgHeight)
	draw.Draw(canvas, destRect, img, bounds.Min, draw.Over)

	return canvas
}

// PadImage adds padding around an image
func PadImage(img image.Image, padding int) image.Image {
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	// Create new image with padding
	newWidth := imgWidth + 2*padding
	newHeight := imgHeight + 2*padding
	result := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Draw original image with padding offset
	destRect := image.Rect(padding, padding, padding+imgWidth, padding+imgHeight)
	draw.Draw(result, destRect, img, bounds.Min, draw.Over)

	return result
}

// IsTransparent checks if a pixel is transparent
func IsTransparent(c color.Color) bool {
	_, _, _, a := c.RGBA()
	return a == 0
}

// GetImageBounds returns the actual content bounds of an image (excluding transparent areas)
func GetImageBounds(img image.Image) image.Rectangle {
	bounds := img.Bounds()

	minX, minY := bounds.Max.X, bounds.Max.Y
	maxX, maxY := bounds.Min.X, bounds.Min.Y
	found := false

	// Scan for non-transparent pixels
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if !IsTransparent(img.At(x, y)) {
				if !found {
					minX, minY = x, y
					maxX, maxY = x, y
					found = true
				} else {
					if x < minX {
						minX = x
					}
					if x > maxX {
						maxX = x
					}
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}
			}
		}
	}

	if !found {
		return image.Rect(0, 0, 0, 0)
	}

	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// CreateTransparentImage creates a transparent image of the specified size
func CreateTransparentImage(width, height int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, width, height))
}

// CopyImage creates a copy of an image
func CopyImage(img image.Image) image.Image {
	bounds := img.Bounds()
	result := image.NewRGBA(bounds)
	draw.Draw(result, bounds, img, bounds.Min, draw.Src)
	return result
}

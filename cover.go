package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
)

type Interpolation int

const (
	NearestNeighbor Interpolation = iota
	Bilinear
)

var (
	white       = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	black       = color.RGBA{R: 5, G: 5, B: 5, A: 255}
	grey        = color.RGBA{R: 82, G: 82, B: 82, A: 255}
	blue        = color.RGBA{R: 51, G: 181, B: 229, A: 255}
	green       = color.RGBA{R: 153, G: 204, B: 0, A: 255}
	orange      = color.RGBA{R: 255, G: 187, B: 51, A: 255}
	red         = color.RGBA{R: 255, G: 68, B: 68, A: 255}
	purple      = color.RGBA{R: 170, G: 102, B: 204, A: 255}
	coverColors = []color.RGBA{blue, green, orange, red, purple}
)

func GenerateCover(text string) image.Image {
	cover := image.NewNRGBA(image.Rect(0, 0, 600, 600))
	coverColor := coverColors[int64(Hash(text))%int64(len(coverColors))]
	draw.Draw(cover, cover.Bounds(), &image.Uniform{C: white}, image.Point{}, draw.Src)
	draw.Draw(cover, image.Rect(12, 12, 588, 400), &image.Uniform{C: coverColor}, image.Point{}, draw.Src)
	verticalGradient(cover, image.Rect(12, 400, 588, 588), grey, black)
	return cover
}

func verticalGradient(img *image.NRGBA, rect image.Rectangle, fromColor, toColor color.RGBA) {
	for x := 0; x < rect.Dx(); x++ {
		for y := 0; y < rect.Dy(); y++ {
			d := float64(y) / float64(rect.Dy())
			img.Set(rect.Min.X+x, rect.Min.Y+y, color.RGBA{
				R: uint8(float64(fromColor.R) + d*(float64(toColor.R)-float64(fromColor.R))),
				G: uint8(float64(fromColor.G) + d*(float64(toColor.G)-float64(fromColor.G))),
				B: uint8(float64(fromColor.B) + d*(float64(toColor.B)-float64(fromColor.B))),
				A: 255,
			})
		}
	}
}

func OpenImage(filename string) image.Image {
	file := ProcessErrorArg(os.Open(filename)).(*os.File)
	defer Close(file)
	img, _, err := image.Decode(file)
	ProcessError(err)
	return img
}

func ResizeImage(source image.Image, scale float64, interpolation Interpolation) image.Image {
	width, height := int(float64(source.Bounds().Dx())*scale), int(float64(source.Bounds().Dy())*scale)
	target := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch interpolation {
			case NearestNeighbor:
				target.Set(x, y, nearestNeighborInterpolation(source, float64(x)/scale, float64(y)/scale))
			case Bilinear:
				target.Set(x, y, bilinearInterpolation(source, float64(x)/scale, float64(y)/scale))
			}
		}
	}
	return target
}

func nearestNeighborInterpolation(img image.Image, x, y float64) color.Color {
	return img.At(int(math.Floor(x)), int(math.Floor(y)))
}

func bilinearInterpolation(img image.Image, x, y float64) color.Color {
	gx, tx := math.Modf(x)
	gy, ty := math.Modf(y)
	srcX, srcY := int(gx), int(gy)
	r00, g00, b00, a00 := img.At(srcX+0, srcY+0).RGBA()
	r10, g10, b10, a10 := img.At(srcX+1, srcY+0).RGBA()
	r01, g01, b01, a01 := img.At(srcX+0, srcY+1).RGBA()
	r11, g11, b11, a11 := img.At(srcX+1, srcY+1).RGBA()
	return color.RGBA64{
		R: blerp(r00, r10, r01, r11, tx, ty),
		G: blerp(g00, g10, g01, g11, tx, ty),
		B: blerp(b00, b10, b01, b11, tx, ty),
		A: blerp(a00, a10, a01, a11, tx, ty),
	}
}

func blerp(c00, c10, c01, c11 uint32, tx, ty float64) uint16 {
	return uint16(lerp(lerp(float64(c00), float64(c10), tx), lerp(float64(c01), float64(c11), tx), ty))
}

func lerp(s, e, t float64) float64 {
	return s + (e-s)*t
}

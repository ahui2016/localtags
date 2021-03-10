package graphics

import (
	"bytes"
	"image"
	"image/jpeg"
	"math"

	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

const (
	defaultSize            = 128
	defaultQuality         = 85
	defaultLimit   float64 = 900
)

// ResizeLimit resizes the image if it's long side bigger than limit.
// Use default limit 900 if limit is set to zero.
// Use default quality 85 if quality is set to zero.
func ResizeLimit(img []byte, limit float64, quality int) (*bytes.Buffer, error) {
	src, err := ReadImage(img)
	if err != nil {
		return nil, err
	}
	w, h := limitWidthHeight(src.Bounds(), limit)
	small := imaging.Resize(src, w, h, imaging.Lanczos)
	return jpegEncode(small, quality)
}

// Thumbnail create a thumbnail of imgFile.
// Use default size(128) if size is set to zero.
// Use default quality(85) if quality is set to zero.
func Thumbnail(img []byte, size, quality int) (*bytes.Buffer, error) {
	if size == 0 {
		size = defaultSize
	}
	src, err := ReadImage(img)
	if err != nil {
		return nil, err
	}
	side := shortSide(src.Bounds())
	src = imaging.CropCenter(src, side, side)
	src = imaging.Resize(src, size, 0, imaging.Lanczos)
	return jpegEncode(src, quality)
}

// Use default quality(85) if quality is set to zero.
func jpegEncode(src image.Image, quality int) (*bytes.Buffer, error) {
	if quality == 0 {
		quality = defaultQuality
	}
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, src, &jpeg.Options{Quality: quality})
	return buf, err
}

// ReadImage converts bytes to image. Supports webp.
func ReadImage(img []byte) (image.Image, error) {
	r := bytes.NewReader(img)
	src, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		r.Reset(img)
		if src, err = webp.Decode(r); err != nil {
			return nil, err
		}
	}
	return src, nil
}

func shortSide(bounds image.Rectangle) int {
	if bounds.Dx() < bounds.Dy() {
		return bounds.Dx()
	}
	return bounds.Dy()
}

// Use default limit(900) if limit is set to zero.
func limitWidthHeight(bounds image.Rectangle, limit float64) (limitWidth, limitHeight int) {
	if limit == 0 {
		limit = defaultLimit
	}
	w := float64(bounds.Dx())
	h := float64(bounds.Dy())
	// 先限制宽度
	if w > limit {
		h *= limit / w
		w = limit
	}
	// 缩小后的高度仍有可能超过限制，因此要再判断一次
	if h > limit {
		w *= limit / h
		h = limit
	}
	limitWidth = int(math.Round(w))
	limitHeight = int(math.Round(h))
	return
}

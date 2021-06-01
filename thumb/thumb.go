package thumb

import (
	"bytes"
	"image"
	"image/jpeg"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"

	"github.com/ahui2016/localtags/util"
	"github.com/disintegration/imaging"
	"golang.org/x/image/webp"
)

const (
	ffmpeg                 = "ffmpeg"
	ffprobe                = "ffprobe"
	defaultSize            = 128
	defaultQuality         = 85
	defaultLimit   float64 = 900
)

// CheckImage 检查图片能否正常使用。
/*
func CheckImage(img []byte) error {
	_, err := Nail(img, 0, 0)
	return err
}
*/

// NailWrite reads an image from imgPath, creates a thumbnail of it,
// and write the thumbnail to thumbPath.
func NailWrite(imgPath, thumbPath string) error {
	img, err := os.ReadFile(imgPath)
	if err != nil {
		return err
	}
	return BytesToThumb(img, thumbPath)
}

// BytesToThumb creates a thumbnail from img, uses default size and default quality,
// and write the thumbnail to thumbPath.
func BytesToThumb(img []byte, thumbPath string) error {
	buf, err := Nail(img, 0, 0)
	if err != nil {
		return err
	}
	return util.CreateFile(thumbPath, buf)
}

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

// Nail creates a thumbnail of the img.
// Use default size(128) if size is set to zero.
// Use default quality(85) if quality is set to zero.
func Nail(img []byte, size, quality int) (*bytes.Buffer, error) {
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

// CheckFFmpeg 检查系统有没有安装 ffmpeg 和 ffprobe
func CheckFFmpeg() (ok bool) {
	ffmpegPath, err1 := exec.LookPath(ffmpeg)
	ffprobePath, err2 := exec.LookPath(ffprobe)
	err := util.WrapErrors(err1, err2)
	if err == nil {
		ok = true
	}
	log.Print(ffmpegPath, ffprobePath, err)
	return
}

// OneFrame 截取视频文件 in 的其中一帧 (第 n 秒)，保存到文件 out 中。
// 建议 out 文件名的后缀为 ".jpg"。
// 例: OneFrame(video.mp4, screenshot.jpg, 10)
func OneFrame(in, out string, n int) error {
	cmd := exec.Command(
		ffmpeg,                 // 命令名
		"-ss", strconv.Itoa(n), // 从视频开头算起第 n 秒
		"-i", in, // 视频文件名
		"-frames:v", "1", // 截取 1 帧
		"-q:v", "9", // 截图质量，好像是 1 最高、9 最低
		"-y", // 自动覆盖文件
		out,  // 截图保存位置
	)
	return cmd.Run()
}

// FrameNail 截取视频文件 in 中的一帧 (第 n 秒),
// 并剪裁成正方形缩略图保存到文件 out 中。
func FrameNail(in, out string, n int) error {
	err := OneFrame(in, out, n)
	if err != nil {
		return err
	}
	img, err := os.ReadFile(out)
	if err != nil {
		return err
	}
	return BytesToThumb(img, out)
}

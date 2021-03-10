package thumbnail

import (
	"log"
	"os/exec"
	"strconv"

	"github.com/ahui2016/localtags/util"
)

const (
	ffmpeg  = "ffmpeg"
	ffprobe = "ffprobe"
)

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
		"-ss", strconv.Itoa(n), // 从影片开头算起第 n 秒
		"-i", in, // 影片文件名
		"-frames:v", "1", // 截取 1 帧
		"-q:v", "2", // 截图质量，2 是较高质量
		"-y", // 自动覆盖文件
		out,  // 截图保存位置
	)
	return cmd.Run()
}

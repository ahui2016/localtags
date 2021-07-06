package util

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/blake2b"
)

// WrapErrors 把多个错误合并为一个错误.
func WrapErrors(allErrors ...error) (wrapped error) {
	for _, err := range allErrors {
		if err != nil {
			if wrapped == nil {
				wrapped = err
			} else {
				wrapped = fmt.Errorf("%v | %v", err, wrapped)
			}
		}
	}
	return
}

// ErrorContains returns NoCaseContains(err.Error(), substr)
// Returns false if err is nil.
func ErrorContains(err error, substr string) bool {
	if err == nil {
		return false
	}
	return noCaseContains(err.Error(), substr)
}

// noCaseContains reports whether substr is within s case-insensitive.
func noCaseContains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// Panic panics if err != nil
func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

// UserHomeDir 就是 os.UserHomeDir
func UserHomeDir() string {
	homeDir, err := os.UserHomeDir()
	Panic(err)
	return homeDir
}

// PathIsNotExist .
func PathIsNotExist(name string) (bool, error) {
	_, err := os.Lstat(name)
	if os.IsNotExist(err) {
		return true, nil
	}
	return false, err
}

// PathIsExist .
func PathIsExist(name string) (bool, error) {
	isNotExist, err := PathIsNotExist(name)
	return !isNotExist, err
}

// MustMkdir 确保有一个名为 dirName 的文件夹，
// 如果没有则自动创建，如果已存在则不进行任何操作。
func MustMkdir(dirName string) {
	ok, err := PathIsNotExist(dirName)
	Panic(err)
	if ok {
		Panic(os.Mkdir(dirName, 0700))
	}
}

// Sha256Hex 返回 sha256 的 hex 字符串。
// 虽然函数名是 Sha256, 但实际上采用 BLAKE2b 算法。
func Sha256Hex(data []byte) string {
	sum := blake2b.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// FileSha256Hex 返回文件 name 的 hex 字符串。
// 虽然函数名是 Sha256, 但实际上采用 BLAKE2b 算法。
func FileSha256Hex(name string) (string, error) {
	fileBytes, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	return Sha256Hex(fileBytes), nil
}

// MustMarshal .
func MustMarshal(data interface{}) []byte {
	dataJSON, err := json.Marshal(data)
	Panic(err)
	return dataJSON
}

// MustMarshalWrite 把 data 转换为漂亮格式的 JSON 并写入文件 name 中。
func MustMarshalWrite(data interface{}, name string) {
	Panic(MarshalWrite(data, name))
}

// MarshalWrite 把 data 转换为漂亮格式的 JSON 并写入文件 name 中。
func MarshalWrite(data interface{}, name string) error {
	dataJSON, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(name, dataJSON, 0600)
}

// CreateFile 把 src 的数据写入 filePath, 权限是 0600, 自动关闭 file.
func CreateFile(filePath string, src io.Reader) error {
	_, file, err := CreateReturnFile(filePath, src)
	if err == nil {
		file.Close()
	}
	return err
}

// CreateReturnFile 把 src 的数据写入 filePath, 权限是 0600,
// 会自动创建或覆盖文件，返回 file, 要记得关闭资源。
func CreateReturnFile(filePath string, src io.Reader) (int64, *os.File, error) {
	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return 0, nil, err
	}
	size, err := io.Copy(f, src)
	if err != nil {
		return 0, nil, err
	}
	return size, f, nil
}

func DeleteFiles(files []string) (err error) {
	for _, file := range files {
		e := os.Remove(file)
		err = WrapErrors(err, e)
	}
	return err
}

// https://stackoverflow.com/questions/50740902/move-a-file-to-a-different-drive-with-go
func MoveFile(destPath, sourcePath string) error {
	if err := CopyFile(destPath, sourcePath); err != nil {
		return err
	}
	return os.Remove(sourcePath)
}

// https://stackoverflow.com/questions/30376921/how-do-you-copy-a-file-in-go
func CopyFile(destPath, sourcePath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	outputFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	_, err1 := io.Copy(outputFile, inputFile)
	err2 := outputFile.Sync()
	return WrapErrors(err1, err2)
}

// HasString reports whether item is in the slice.
func HasString(slice []string, item string) bool {
	i := StringIndex(slice, item)
	return i >= 0
}

// StringIndex returns the index of item in the slice.
// returns -1 if not found.
func StringIndex(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// StrSliceRemove 删除 slice 中的第 i 个元素。
func StrSliceRemove(slice []string, i int) []string {
	return append(slice[:i], slice[i+1:]...)
}

// StrSliceDiff 对比新旧数组的差异，并返回需要新增的项目与需要删除的项目。
func StrSliceDiff(newArr, oldArr []string) (toAdd, toDelete []string) {
	// newArr 里有，oldArr 里没有的，需要添加到数据库。
	for _, newItem := range newArr {
		if !HasString(oldArr, newItem) {
			toAdd = append(toAdd, newItem)
		}
	}
	// oldTags 里有，newTags 里没有的，需要从数据库中删除。
	for _, oldItem := range oldArr {
		if !HasString(newArr, oldItem) {
			toDelete = append(toDelete, oldItem)
		}
	}
	return
}

// FirstLineLimit 返回第一行，并限定长度，其中 s 必须事先 TrimSpace 并确保不是空字串。
// 并且, s 也可以事先限制字数.
// 该函数会尽量确保最后一个字符是有效的 utf8 字符，但当第一行中的全部字符都无效时，
// 则不保证有效字符。
func FirstLineLimit(s string, limit int) string {
	if len(s) > limit {
		s = s[:limit]
	}
	s += "\n"
	i := firstLineBreak(s)
	firstLine := s[:i]
	for len(firstLine) > 0 {
		if utf8.ValidString(firstLine) {
			break
		}
		firstLine = firstLine[:len(firstLine)-1]
	}
	if firstLine == "" {
		firstLine = s[:i]
	}
	return firstLine
}

// firstLineBreak 获取第一个 \r\n 或第一个 \n 的位置
func firstLineBreak(s string) int {
	i := strings.Index(s, "\n")
	i2 := strings.Index(s, "\r\n")
	if i2 < 0 {
		return i
	}
	if i > i2 {
		i = i2
	}
	return i
}

// GetMarkdownTitle 截取 markdown 的开头内容获作为标题。
// 其中 title 应该限制字数 (使用 FirstLineLimit), 没必要传入 markdown 文件的全文。
// 注意 title 不可包含制表符、回车键等特殊字符。
func GetMarkdownTitle(title string) string {
	reTitle := regexp.MustCompile(`(^#{1,6}|>|1.|-|\*) (.+)`)
	matches := reTitle.FindStringSubmatch(title)
	// 这个 matches 要么为空，要么包含 3 个元素
	if len(matches) >= 3 {
		title = matches[2]
	}
	// 把 Windows 禁止用来做文件名的字符替换为空格
	forbidden := regexp.MustCompile(`[:/\|"?<>*]`)
	return forbidden.ReplaceAllString(title, " ")
}

// GetMIME returns the content-type of a file extension.
// https://github.com/gofiber/fiber/blob/master/utils/http.go (edited).
func GetMIME(extension string) (mime string) {
	const MIMEOctetStream = "application/octet-stream"
	extension = strings.ToLower(extension)

	if len(extension) == 0 {
		return mime
	}
	mime = mimeExtensions[extension]
	if len(mime) == 0 {
		return MIMEOctetStream
	}
	return mime
}

// MIME types were copied from
// https://github.com/gofiber/fiber/blob/master/utils/http.go
// https://github.com/nginx/nginx/blob/master/conf/mime.types
var mimeExtensions = map[string]string{
	"html":    "text/html",
	"htm":     "text/html",
	"shtml":   "text/html",
	"css":     "text/css",
	"gif":     "image/gif",
	"jpeg":    "image/jpeg",
	"jpg":     "image/jpeg",
	"xml":     "application/xml",
	"js":      "application/javascript",
	"atom":    "application/atom+xml",
	"rss":     "application/rss+xml",
	"mml":     "text/mathml",
	"txt":     "text/plain",
	"jad":     "text/vnd.sun.j2me.app-descriptor",
	"wml":     "text/vnd.wap.wml",
	"htc":     "text/x-component",
	"png":     "image/png",
	"svg":     "image/svg+xml",
	"svgz":    "image/svg+xml",
	"tif":     "image/tiff",
	"tiff":    "image/tiff",
	"wbmp":    "image/vnd.wap.wbmp",
	"webp":    "image/webp",
	"ico":     "image/x-icon",
	"jng":     "image/x-jng",
	"bmp":     "image/x-ms-bmp",
	"woff":    "font/woff",
	"woff2":   "font/woff2",
	"jar":     "application/java-archive",
	"war":     "application/java-archive",
	"ear":     "application/java-archive",
	"json":    "application/json",
	"hqx":     "application/mac-binhex40",
	"doc":     "application/msword",
	"pdf":     "application/pdf",
	"ps":      "application/postscript",
	"eps":     "application/postscript",
	"ai":      "application/postscript",
	"rtf":     "application/rtf",
	"m3u8":    "application/vnd.apple.mpegurl",
	"kml":     "application/vnd.google-earth.kml+xml",
	"kmz":     "application/vnd.google-earth.kmz",
	"xls":     "application/vnd.ms-excel",
	"eot":     "application/vnd.ms-fontobject",
	"ppt":     "application/vnd.ms-powerpoint",
	"odg":     "application/vnd.oasis.opendocument.graphics",
	"odp":     "application/vnd.oasis.opendocument.presentation",
	"ods":     "application/vnd.oasis.opendocument.spreadsheet",
	"odt":     "application/vnd.oasis.opendocument.text",
	"wmlc":    "application/vnd.wap.wmlc",
	"7z":      "application/x-7z-compressed",
	"cco":     "application/x-cocoa",
	"jardiff": "application/x-java-archive-diff",
	"jnlp":    "application/x-java-jnlp-file",
	"run":     "application/x-makeself",
	"pl":      "application/x-perl",
	"pm":      "application/x-perl",
	"prc":     "application/x-pilot",
	"pdb":     "application/x-pilot",
	"rar":     "application/x-rar-compressed",
	"rpm":     "application/x-redhat-package-manager",
	"sea":     "application/x-sea",
	"swf":     "application/x-shockwave-flash",
	"sit":     "application/x-stuffit",
	"tcl":     "application/x-tcl",
	"tk":      "application/x-tcl",
	"der":     "application/x-x509-ca-cert",
	"pem":     "application/x-x509-ca-cert",
	"crt":     "application/x-x509-ca-cert",
	"xpi":     "application/x-xpinstall",
	"xhtml":   "application/xhtml+xml",
	"xspf":    "application/xspf+xml",
	"zip":     "application/zip",
	"bin":     "application/octet-stream",
	"exe":     "application/octet-stream",
	"dll":     "application/octet-stream",
	"deb":     "application/octet-stream",
	"dmg":     "application/octet-stream",
	"iso":     "application/octet-stream",
	"img":     "application/octet-stream",
	"msi":     "application/octet-stream",
	"msp":     "application/octet-stream",
	"msm":     "application/octet-stream",
	"mid":     "audio/midi",
	"midi":    "audio/midi",
	"kar":     "audio/midi",
	"mp3":     "audio/mpeg",
	"ogg":     "audio/ogg",
	"m4a":     "audio/x-m4a",
	"ra":      "audio/x-realaudio",
	"3gpp":    "video/3gpp",
	"3gp":     "video/3gpp",
	"ts":      "video/mp2t",
	"mp4":     "video/mp4",
	"mpeg":    "video/mpeg",
	"mpg":     "video/mpeg",
	"mov":     "video/quicktime",
	"webm":    "video/webm",
	"flv":     "video/x-flv",
	"m4v":     "video/x-m4v",
	"mng":     "video/x-mng",
	"asx":     "video/x-ms-asf",
	"asf":     "video/x-ms-asf",
	"wmv":     "video/x-ms-wmv",
	"avi":     "video/x-msvideo",
}

package model

import (
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/ahui2016/localtags/util"
)

const (
	// FileNameMinLength 规定包括后缀名在内文件名长度不可小于 5.
	FileNameMinLength = 3
)

// File 文件
type File struct {
	ID      string
	Name    string
	Size    int64
	Type    string
	Thumb   bool   // has a thumbnail or not
	Hash    string // checksum
	Like    int64  // 点赞
	CTime   int64  // created at
	UTime   int64  // updated at
	Deleted bool
}

func NewFile(id string) *File {
	now := TimeNow()
	return &File{
		ID:    id,
		CTime: now,
		UTime: now,
	}
}

func TimeNow() int64 {
	return time.Now().Unix()
}

// SetNameType 同时设置 Name 和 Type.
// 使用 SetNameType 可确保正确设置 Type.
func (file *File) SetNameType(filename string) error {
	filename = strings.TrimSpace(filename)
	if len(filename) < FileNameMinLength {
		return errors.New("filename is too short")
	}
	file.Name = filename
	file.Type = typeByFilename(filename)
	return nil
}

func typeByFilename(filename string) (filetype string) {
	ext := filepath.Ext(filename)
	ext = strings.TrimPrefix(ext, ".")
	filetype = util.GetMIME(ext)
	switch ext {
	case "zip", "rar", "7z", "gz", "tar", "bz", "bz2", "xz":
		filetype = "compressed/" + ext
	case "md", "xml", "html", "xhtml", "htm":
		filetype = "text/" + ext
	case "doc", "docx", "ppt", "pptx", "rtf", "xls", "xlsx":
		filetype = "office/" + ext
	case "epub", "pdf", "mobi", "azw", "azw3", "djvu":
		filetype = "ebook/" + ext
	}
	return filetype
}

// Tag 标签
type Tag struct {
	ID    string // base64 of Name
	Name  string
	Count int
	CTime int // created at
}

// TagGroup 标签组，其中 Tags 应该除重和排序。
type TagGroup struct {
	ID        string // primary key, random
	Tags      []string
	CTime     int // created at
	UTime     int // updated at
	Protected bool
}

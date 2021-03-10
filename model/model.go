package model

import (
	"errors"
	"path/filepath"
	"strings"

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
	Thumb   string // the src of a thumbnail
	Hash    string // checksum
	Like    int    // 点赞
	CTime   int    // created at
	UTime   int    // updated at
	Deleted bool
}

// SetNameType 同时设置 Name 和 Type.
// 请勿直接设置 Name, 每次都应该使用 SetNameType 以确保同时设置 Type.
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

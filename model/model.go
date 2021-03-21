package model

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ahui2016/localtags/stringset"
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
	Count   int64 // how many files with the same name
	Size    int64
	Type    string
	Thumb   bool   // has a thumbnail or not
	Hash    string // checksum
	Like    int64  // 点赞
	CTime   int64  // created at
	UTime   int64  // updated at
	Deleted bool
	Tags    []string // 该项目不在数据库中，放在这里只是为了方便
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

// SetTags 对标签进行一些验证和处理（例如除重和排序）。
// 尽量不要直接操作 file.Tags
func (file *File) SetTags(names []string) error {
	tags := stringset.UniqueSort(names)
	if len(tags) < 2 {
		return errors.New("too few tags (at least two)")
	}
	file.Tags = purify(tags)
	return nil
}

// purify 清除标签中的非法字符。
func purify(tags []string) []string {
	re := regexp.MustCompile(`[#;,，'"/\+\n]`)
	for i := range tags {
		tags[i] = re.ReplaceAllString(tags[i], "")
	}
	return tags
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
	ID    string
	CTime int64 // created at
	Count int64 // 该项目不在数据库中，放在这里只是为了方便
}

func NewTag(id string) *Tag {
	return &Tag{
		ID:    id,
		CTime: TimeNow(),
	}
}

// TagGroup 标签组，其中 Tags 应该除重和排序。
type TagGroup struct {
	ID        string // primary key, random
	Tags      []string
	CTime     int64 // created at
	UTime     int64 // updated at
	Protected bool
}

// NewTagGroup .
func NewTagGroup() *TagGroup {
	now := TimeNow()
	return &TagGroup{
		ID:    RandomID(),
		CTime: now,
		UTime: now,
	}
}

func (group *TagGroup) SetTags(tags []string) {
	group.Tags = stringset.UniqueSort(tags)
}

func (group *TagGroup) String() string {
	tags := util.MustMarshal(group.Tags)
	return string(tags)
}

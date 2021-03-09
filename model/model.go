package model

// File 文件
type File struct {
	ID      string
	Name    string
	Size    int
	Type    string
	Hash    string // checksum
	Like    int    // 点赞
	CTime   int    // created at
	UTime   int    // updated at
	Deleted bool
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

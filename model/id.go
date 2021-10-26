package model

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex = sync.Mutex{}

// ShortID 用来生成 “可爱” 的自增 ID.
// 说它可爱是因为它的字符串形式:
// 1.短 2.由数字和字母组成,不分大小写,且确保以字母开头 3.趋势自增，但又不明显自增 4.可利用前缀分类
// 该 ID 由前缀、年份与自增数三部分组成，年份与自增数分别转 36 进制字符。
// 前缀只能是单个字母，因为 ID 还是短一些的好。
// 注意：前缀不分大小写, ShortID 本身也不分大小写。
type ShortID struct {
	Prefix string
	Year   int64
	Count  int64
}

// FirstID 生成对应前缀的初始 id, 后续使用 Next 函数来获取下一个 id.
// prefix 只能是单个英文字母。
func FirstID(prefix string) (id ShortID, err error) {
	if len(prefix) > 1 {
		err = fmt.Errorf("the prefix [%s] is too long", prefix)
		return
	}
	prefix = strings.ToUpper(prefix)
	if prefix < "A" || prefix > "Z" {
		err = fmt.Errorf("the prefix [%s] is not an English character", prefix)
		return
	}
	id.Prefix = prefix
	id.Year = int64(time.Now().Year())
	return
}

// ParseID 把字符串形式的 id 转换为 IncreaseID.
// (有“万年虫”问题，大概公元五万年时本算法会出错，当然，这个问题可以忽略。)
func ParseID(strID string) (id ShortID, err error) {
	prefix := strID[:1]
	strYear := strID[1:4] // 可以姑且认为年份总是占三个字符
	strCount := strID[4:]
	year, err := strconv.ParseInt(strYear, 36, 0)
	if err != nil {
		return id, err
	}
	count, err := strconv.ParseInt(strCount, 36, 0)
	if err != nil {
		return id, err
	}
	id.Prefix = prefix
	id.Year = year
	id.Count = count
	return
}

// Next 使 id 自增一次，输出自增后的新 id.
// 如果当前年份大于 id 中的年份，则年份进位，Count 重新计数。
// 否则，年份不变，Count 加一。
func (id ShortID) Next() ShortID {
	mutex.Lock()
	defer mutex.Unlock()

	nowYear := int64(time.Now().Year())
	if nowYear > id.Year {
		return ShortID{id.Prefix, nowYear, 0}
	}
	return ShortID{id.Prefix, id.Year, id.Count + 1}
}

// String 返回 id 的字符串形式。
func (id ShortID) String() string {
	year := strconv.FormatInt(id.Year, 36)
	count := strconv.FormatInt(id.Count, 36)
	strID := id.Prefix + year + count
	return strings.ToUpper(strID)
}

// RandomID 返回一个上升趋势的随机 id, 由时间戳与随机数组成。
// 时间戳确保其上升趋势（大致有序），随机数确保其随机性（防止被穷举, 防冲突）。
// RandomID 考虑了 “生成 id 的速度”、 “并发防冲突” 与 “id 长度”
// 这三者的平衡，适用于大多数中、小规模系统（当然，不适用于大型系统）。
func RandomID() string {
	var max int64 = 100_000_000
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	timestamp := time.Now().Unix()
	idInt64 := timestamp*max + n.Int64()
	return strconv.FormatInt(idInt64, 36)
}

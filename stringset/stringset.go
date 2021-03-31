package stringset

import (
	"sort"

	"github.com/ahui2016/localtags/util"
)

// Set .
type Set struct {
	Map map[string]bool
}

func NewSet() *Set {
	return &Set{make(map[string]bool)}
}

func From(arr []string) *Set {
	set := NewSet()
	for _, v := range arr {
		set.Map[v] = true
	}
	return set
}

// Has .
func (set *Set) Has(item string) bool {
	return set.Map[item]
}

// Add .
func (set *Set) Add(item string) {
	set.Map[item] = true
}

// Intersect .
func (set *Set) Intersect(other *Set) *Set {
	result := NewSet()
	for key := range set.Map {
		if other.Has(key) {
			result.Add(key)
		}
	}
	return result
}

// Slice convert the set to a string slice.
func (set *Set) Slice() (arr []string) {
	for key := range set.Map {
		if set.Has(key) {
			arr = append(arr, key)
		}
	}
	return
}

// UniqueSort 利用 Set 对 arr 进行除重和排序。
func UniqueSort(arr []string) (result []string) {
	if len(arr) == 0 {
		return
	}
	result = From(arr).Slice()
	sort.Strings(result)
	return
}

// Intersect 取 group 里全部集合的交集。
func Intersect(group []*Set) *Set {
	length := len(group)
	if length == 0 {
		return NewSet()
	}
	result := group[0]
	for i := 1; i < length; i++ {
		result = result.Intersect(group[i])
	}
	return result
}

func UniqueSortString(arr []string) string {
	sorted := UniqueSort(arr)
	blob := util.MustMarshal(sorted)
	return string(blob)
}

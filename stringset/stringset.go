package stringset

import "sort"

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

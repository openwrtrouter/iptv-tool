package util

import (
	"os"
	"path/filepath"
	"sort"
)

// GetCurrentAbPathByExecutable 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res, nil
}

// SortedMapKeys 对Map的Key进行排序
func SortedMapKeys[T any](maps map[string]T) []string {
	ret := make([]string, len(maps))
	i := 0
	for name := range maps {
		ret[i] = name
		i++
	}
	sort.Strings(ret)
	return ret
}

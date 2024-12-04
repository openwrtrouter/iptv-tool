package util

import (
	"os"
	"path/filepath"
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

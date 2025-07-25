package helper

import "os"

// ReadDir 读取目录下的所有文件, 返回文件名列表
func ReadDir(dir string) []string {
	var files []string

	dirList, _ := os.ReadDir(dir)

	for _, v := range dirList {
		if !v.IsDir() {
			files = append(files, v.Name())
		}
	}

	return files
}

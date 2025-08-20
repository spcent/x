package helper

import (
	"os"
)

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

// WriteFile will write the info on array of bytes b to filepath. It will set the file
// permission mode to 0660
// Returns an error in case there's any.
func WriteFile(filepath string, b []byte) error {
	return os.WriteFile(filepath, b, 0660)
}

// GetFile will open filepath.
// Returns a tuple with a file and an error in case there's any.
func GetFile(filepath string) (*os.File, error) {
	return os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0777)
}

// DeleteFile will delete filepath permanently.
// Returns an error in case there's any.
func DeleteFile(filepath string) error {
	_, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	err = os.Remove(filepath)
	return err
}

package utils

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
)

func TextToJsonOfFile(fileName string, fn string, root string) (string, error) {

	var newSlice []PathInfo

	infos, err := ReadLines(fileName)
	if err != nil {
		println(err.Error())
		return "", err
	}

	for _, v := range RemoveDuplicateElement(infos) {
		newTag := PathInfo{
			Path: v,
			Hits: 0,
		}
		newSlice = append(newSlice, newTag)

	}

	info, _ := CustomMarshal(newSlice)
	NewFilename := filepath.Join(root, "Data", "DefDict", fn+".json")
	_ = ioutil.WriteFile(NewFilename, []byte(info), 0644)

	return NewFilename, nil
}

// golang读取文件并且返回列表
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// 字符串列表去重函数
func RemoveDuplicateElement(infos []string) []string {
	result := make([]string, 0, len(infos))
	temp := map[string]struct{}{}
	for _, item := range infos {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

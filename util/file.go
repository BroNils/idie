package util

import "os"

func IsFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func CreateFile(filePath string) *os.File {
	if IsFileExists(filePath) {
		return nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	return file
}

func OpenFileOrCreate(filePath string) *os.File {
	if !IsFileExists(filePath) {
		return CreateFile(filePath)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		panic(err)
	}

	return file
}

func WriteStringToFile(file *os.File, s string) {
	_, err := file.WriteString(s)
	if err != nil {
		panic(err)
	}
}

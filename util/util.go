package util

import (
	"strconv"
	"strings"
)

func Explode(str string, glue string) []string {
	return strings.Split(str, glue)
}

func ExplodeToIntSlice(str string, glue string) []int {
	var result []int

	if str == "" {
		return result
	}

	for _, item := range Explode(str, glue) {
		i, err := strconv.Atoi(item)
		if err != nil {
			continue
		}
		result = append(result, i)
	}
	return result
}

func UniqueIntSlice(slice []int) []int {
	keys := make(map[int]bool)
	var list []int

	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

func IsIntSliceContains(slice []int, item int) bool {
	for _, sliceItem := range slice {
		if sliceItem == item {
			return true
		}
	}
	return false
}

// FillPrefixWithRune fills the postfix of the input string with the rune.
// input is the string to be filled.
// length is the length of the output string (input + postfix fill).
// rune is the rune to fill the postfix.
func FillPostfixWithRune(input string, length int, rune rune) string {
	if len(input) >= length {
		return input
	}

	return input + strings.Repeat(string(rune), length-len(input))
}

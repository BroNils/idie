package util

import (
	"reflect"
	"regexp"
	"strings"
)

// support partial input
func IsValidPortList(portList string) bool {
	portList = strings.ReplaceAll(portList, " ", "")

	if portList == "" {
		return true
	}

	// regex of valid port numbers (comma-sep)
	portListRegex := regexp.MustCompile(`^(\d{1,5})(,(\d{1,5}))*$`)

	return portListRegex.MatchString(portList)
}

func IsValidIPv4(input string) bool {
	ipv4Regex := regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}$`)

	return ipv4Regex.MatchString(input)
}

func IsStruct(v interface{}) (ok bool, val reflect.Value) {
	structVal := reflect.ValueOf(v)
	kind := structVal.Kind()
	if kind == reflect.Ptr {
		structVal = structVal.Elem()
		kind = structVal.Kind()
	}
	if kind != reflect.Struct {
		return false, structVal
	}

	return true, structVal
}

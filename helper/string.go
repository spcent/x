package helper

import (
	"crypto/rand"
	"encoding/json"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func EllipseTruncate(s string, max int) string {
	if max < 10 {
		max = 80
	}
	lastSpaceIdx := -1
	curIdx := 0
	len := 0
	for i, r := range s {
		if unicode.IsSpace(r) {
			lastSpaceIdx = i
		}
		curIdx = i
		len++
		if len >= max {
			if lastSpaceIdx != -1 && lastSpaceIdx <= max-3 {
				return s[:lastSpaceIdx] + "..."
			} else {
				return s[:curIdx-3] + "..."
			}
			// If here, string is longer than max, but has no spaces
		}
	}
	return s
}

// CutLastString 截取字符串中最后一段，以@beginChar开始,@endChar结束的字符
// @text 文本
// @beginChar 开始
func CutLastString(text, beginChar, endChar string) string {
	if text == "" || beginChar == "" || endChar == "" {
		return ""
	}

	textRune := []rune(text)

	beginIndex := strings.LastIndex(text, beginChar)
	endIndex := strings.LastIndex(text, endChar)
	if endIndex < 0 || endIndex < beginIndex {
		endIndex = len(textRune)
	}

	return string(textRune[beginIndex+1 : endIndex])
}

func IsBlank(value string) bool {
	return value == ""
}

func IsNotBlank(value string) bool {
	return value != ""
}

func ToUint(value string, def ...uint) (uint, bool) {
	val, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		if len(def) > 0 {
			return def[0], false
		}
		return 0, false
	}
	return uint(val), true
}

func ToUintD(value string, def ...uint) uint {
	val, _ := ToUint(value, def...)
	return val
}

func ToInt(value string, def ...int) (int, bool) {
	val, err := strconv.Atoi(value)
	if err != nil {
		if len(def) > 0 {
			return def[0], false
		}
		return 0, false
	}
	return val, true
}

func ToIntD(value string, def ...int) int {
	val, _ := ToInt(value, def...)
	return val
}

func ToInt32(value string, def ...int32) (int32, bool) {
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		if len(def) > 0 {
			return def[0], false
		}
		return 0, false
	}
	return int32(val), true
}

func ToInt32D(value string, def ...int32) int32 {
	val, _ := ToInt32(value, def...)
	return val
}

func ToInt64(value string, def ...int64) (int64, bool) {
	val, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		if len(def) > 0 {
			return def[0], false
		}
		return 0, false
	}
	return val, true
}

func ToInt64D(value string, def ...int64) int64 {
	val, _ := ToInt64(value, def...)
	return val
}

func ToString(value any) string {
	ret := ""

	if value == nil {
		return ret
	}

	switch t := value.(type) {
	case string:
		ret = t
	case int:
		ret = strconv.Itoa(t)
	case int32:
		ret = strconv.Itoa(int(t))
	case int64:
		ret = strconv.FormatInt(t, 10)
	case uint:
		ret = strconv.Itoa(int(t))
	case uint32:
		ret = strconv.Itoa(int(t))
	case uint64:
		ret = strconv.Itoa(int(t))
	default:
		v, _ := json.Marshal(t)
		ret = string(v)
	}

	return ret
}

func ToStringSlice(val []any) []string {
	var result []string
	for _, item := range val {
		v, ok := item.(string)
		if ok {
			result = append(result, v)
		}
	}
	return result
}

func SplitIndex(s, sep string, index int) (string, bool) {
	ret := strings.Split(s, sep)
	if index >= len(ret) {
		return "", false
	}
	return ret[index], true
}

// KebabToCamel 中划线转小驼峰
func KebabToCamel(str string) string {
	arr := strings.Split(str, "-")

	for i, v := range arr {
		if i > 0 {
			arr[i] = cases.Title(language.English, cases.NoLower).String(v)
		} else {
			arr[i] = strings.ToLower(v)
		}
	}

	return strings.Join(arr, "")
}

// PascalToCamel 大驼峰转小驼峰
func PascalToCamel(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// PascalToSnake 大驼峰转蛇形
func PascalToSnake(s string) string {
	var result string
	for i, v := range s {
		if unicode.IsUpper(v) {
			if i != 0 {
				result += "_"
			}
			result += string(unicode.ToLower(v))
		} else {
			result += string(v)
		}
	}
	return result
}

// SnakeToPascal 蛇形转大驼峰
func SnakeToPascal(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_'
	})

	for i := 0; i < len(words); i++ {
		runes := []rune(words[i])
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}

	return strings.Join(words, "")
}

// GenerateRandomString generates a cryptographically random, alphanumeric string of length n.
func GenerateRandomString(n int) (string, error) {
	const dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes), nil
}

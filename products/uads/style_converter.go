package uads

import (
	"bytes"
	"fmt"
	"strings"
)

type styleConverter interface {
	convertWithErr(string) (string, error)
	unconvertWithErr(string) (string, error)
	convert(string) string
	unconvert(string) string
}

const EnumUnknownString = "unknown"
const EnumUnknownInt = -1

type intConverter struct {
	c map[int]string
	r map[string]int
}

func newIntConverter(input map[int]string) intConverter {
	reversed := make(map[string]int)
	for k, v := range input {
		reversed[v] = k
	}
	return intConverter{
		c: input,
		r: reversed,
	}
}

func (c intConverter) convert(src int) string {
	v, _ := c.convertWithErr(src)
	return v
}

func (c intConverter) unconvert(dst string) int {
	v, _ := c.unconvertWithErr(dst)
	return v
}

func (c intConverter) convertWithErr(src int) (string, error) {
	if dst, ok := c.c[src]; ok {
		return dst, nil
	}
	return EnumUnknownString, fmt.Errorf("")
}

func (c intConverter) unconvertWithErr(dst string) (int, error) {
	if src, ok := c.r[dst]; ok {
		return src, nil
	}
	return EnumUnknownInt, fmt.Errorf("")
}

type upperCamelConverter struct{}

func newUpperCamelConverter(specials map[string]string) styleConverter {
	return &upperCamelConverter{}
}

func (cvt *upperCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if 'a' <= input[0] && input[0] <= 'z' {
		return "", fmt.Errorf("excepted upper camel should not be leading by lowercase character, got %q", input)
	}

	return lowerCamelToLower(strings.ToLower(input[:1]) + input[1:]), nil
}

func (cvt *upperCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt *upperCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}

	output := lowerToLowerCamel(input)
	return strings.ToUpper(output[:1]) + output[1:], nil
}

func (cvt *upperCamelConverter) unconvert(input string) string {
	output, _ := cvt.unconvertWithErr(input)
	return output
}

func lowerCamelToLower(input string) string {
	var state int
	var words []string
	buf := strings.Builder{}
	for i := 0; i < len(input); i++ {
		c, l1 := input[i], lookAhead(&input, i, 1)

		if l1 == 0 {
			buf.Write(bytes.ToLower([]byte{c}))
			words = append(words, buf.String())
			buf.Reset()
			break
		}

		if state == 0 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c)
				state = 1
				words = append(words, buf.String())
				buf.Reset()
			} else {
				buf.WriteByte(c)
			}

			continue
		}

		if state == 1 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c + ('a' - 'A'))
				state = 3
			} else {
				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}

		if state == 3 {
			if 'A' <= l1 && l1 <= 'Z' {
				buf.WriteByte(c + ('a' - 'A'))
			} else {
				words = append(words, buf.String())
				buf.Reset()

				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}
	}

	return strings.Join(words, "_")
}

func lowerToLowerCamel(input string) string {
	iL := strings.Split(input, "_")
	oL := make([]string, len(iL))
	for i, s := range iL {
		oL[i] = strings.Title(s)
	}
	output := strings.Join(oL, "")
	return strings.ToLower(output[:1]) + output[1:]
}

func lookAhead(input *string, index, forward int) byte {
	if len(*input) <= index+forward {
		return 0
	}
	return (*input)[index+forward]
}

var upperCamelCvt = newUpperCamelConverter(nil)

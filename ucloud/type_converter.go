package ucloud

import "fmt"

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

type boolConverter struct {
	c map[bool]string
	r map[string]bool
}

func newBoolConverter(input map[bool]string) boolConverter {
	reversed := make(map[string]bool)
	for k, v := range input {
		reversed[v] = k
	}
	return boolConverter{
		c: input,
		r: reversed,
	}
}

func (c boolConverter) convert(src bool) string {
	v, _ := c.convertWithErr(src)
	return v
}

func (c boolConverter) unconvert(dst string) bool {
	v, _ := c.unconvertWithErr(dst)
	return v
}

func (c boolConverter) convertWithErr(src bool) (string, error) {
	if dst, ok := c.c[src]; ok {
		return dst, nil
	}
	return EnumUnknownString, fmt.Errorf("")
}

func (c boolConverter) unconvertWithErr(dst string) (bool, error) {
	if src, ok := c.r[dst]; ok {
		return src, nil
	}
	return false, fmt.Errorf("")
}

type stringConverter struct {
	c map[string]string
	r map[string]string
}

func newStringConverter(input map[string]string) stringConverter {
	reversed := make(map[string]string)
	for k, v := range input {
		reversed[v] = k
	}
	return stringConverter{
		c: input,
		r: reversed,
	}
}

func (c stringConverter) convert(src string) string {
	if dst, ok := c.c[src]; ok {
		return dst
	}
	return src
}

func (c stringConverter) unconvert(dst string) string {
	if src, ok := c.r[dst]; ok {
		return src
	}
	return dst
}

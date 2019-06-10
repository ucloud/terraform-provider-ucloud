package ucloud

import (
	"bytes"
	"fmt"
	"strings"
)

// Converter is use for converting string to another string with specifical style
type styleConverter interface {
	convertWithErr(string) (string, error)
	unconvertWithErr(string) (string, error)
	convert(string) string
	unconvert(string) string
}

type upperConverter struct{}

func newUpperConverter(specials map[string]string) styleConverter {
	return &upperConverter{}
}

// convert is an utils used for converting upper case name with underscore into lower case with underscore.
func (cvt *upperConverter) convertWithErr(input string) (string, error) {
	if input != strings.ToUpper(input) {
		return "", fmt.Errorf("excepted input string is uppercase with underscore, got %q", input)
	}
	return cvt.convert(input), nil
}

func (cvt *upperConverter) convert(input string) string {
	return strings.ToLower(input)
}

// unconvert is an utils used for converting lower case with underscore into upper case name with underscore.
func (cvt *upperConverter) unconvertWithErr(input string) (string, error) {
	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}
	return strings.ToUpper(input), nil
}

func (cvt *upperConverter) unconvert(input string) string {
	return strings.ToUpper(input)
}

type lowerCamelConverter struct{}

func newLowerCamelConverter(specials map[string]string) styleConverter {
	return &lowerCamelConverter{}
}

func (cvt *lowerCamelConverter) convertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if 'A' <= input[0] && input[0] <= 'Z' {
		return "", fmt.Errorf("excepted lower camel should not be leading by uppercase character, got %q", input)
	}

	return lowerCamelToLower(input), nil
}

func (cvt *lowerCamelConverter) convert(input string) string {
	output, _ := cvt.convertWithErr(input)
	return output
}

func (cvt *lowerCamelConverter) unconvertWithErr(input string) (string, error) {
	if len(input) == 0 {
		return "", nil
	}

	if input != strings.ToLower(input) {
		return "", fmt.Errorf("excepted input string is lowercase with underscore, got %q", input)
	}

	return cvt.unconvert(input), nil
}

func (cvt *lowerCamelConverter) unconvert(input string) string {
	return lowerToLowerCamel(input)
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
	// eg. createFail -> create_fail; createUDBFAIL -> create_udb_fail -> createUdbFail
	var state int
	var words []string
	buf := strings.Builder{}
	for i := 0; i < len(input); i++ {
		c, l1 := input[i], lookAhead(&input, i, 1)

		// last character
		if l1 == 0 {
			buf.Write(bytes.ToLower([]byte{c}))
			words = append(words, buf.String())
			buf.Reset()
			break
		}

		if state == 0 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createing UDBInstance
				//         ^ ^
				//         | |
				//         c l1
				buf.WriteByte(c)
				state = 1

				words = append(words, buf.String())
				buf.Reset()
			} else {
				// createi ngUDBInstance
				//       ^ ^
				//       | |
				//       c l1
				buf.WriteByte(c)
			}

			continue
		}

		if state == 1 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createingU DBInstance
				//          ^ ^
				//          | |
				//          c l1
				buf.WriteByte(c + ('a' - 'A'))
				state = 3
			} else {
				// createingI nstance
				//          ^ ^
				//          | |
				//          c l1
				buf.WriteByte(c + ('a' - 'A'))
				state = 0
			}

			continue
		}

		if state == 3 {
			if 'A' <= l1 && l1 <= 'Z' {
				// createingUD BInstance
				//           ^ ^
				//           | |
				//           c l1
				buf.WriteByte(c + ('a' - 'A'))
			} else {
				// createingUDBI nstance
				//             ^ ^
				//             | |
				//             c l1
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
	if len((*input)) <= index+forward {
		return 0
	}
	return (*input)[index+forward]
}

package unet

type stringConverter struct {
	values  map[string]string
	reverse map[string]string
}

func newStringConverter(input map[string]string) stringConverter {
	reversed := make(map[string]string)
	for k, v := range input {
		reversed[v] = k
	}
	return stringConverter{
		values:  input,
		reverse: reversed,
	}
}

func (c stringConverter) convert(value string) string {
	if converted, ok := c.values[value]; ok {
		return converted
	}
	return value
}

func (c stringConverter) unconvert(value string) string {
	if converted, ok := c.reverse[value]; ok {
		return converted
	}
	return value
}

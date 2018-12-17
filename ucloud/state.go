package ucloud

func stateFuncTag(v interface{}) string {
	if len(v.(string)) == 0 {
		return defaultTag
	}
	return v.(string)
}

package llmproxy

// util
func StringPtr(s string) *string {
	return &s
}
func UInt32Prt(i uint32) *uint32 {
	return &i
}

func headerArrayToMap(headersArray [][2]string) map[string]string {
	headerMap := make(map[string]string)
	for _, header := range headersArray {
		headerMap[header[0]] = header[1]
	}
	return headerMap
}

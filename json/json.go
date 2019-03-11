package libjson

import "encoding/json"

func Marshal(d interface{}) string {
	data, _ := json.Marshal(d)
	return string(data)
}

func Unmarshal(d string, v interface{}) {
	json.Unmarshal([]byte(d), v)
}

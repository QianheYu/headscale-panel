package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Structs to json
func Struct2Json(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("[Struct2Json]转换异常: %v", err))
	}
	return string(str)
}

// json to structure
func Json2Struct(str string, obj any) {
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		panic(fmt.Sprintf("[Json2Struct]转换异常: %v", err))
	}
}

// json interface to structure
func JsonI2Struct(str interface{}, obj interface{}) {
	JsonStr := str.(string)
	Json2Struct(JsonStr, obj)
}

func RequestJson2Struct(url string, jsonStruct interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	buffer := bytes.NewBuffer(make([]byte, 8192))
	buffer.Reset()
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(buffer.Bytes(), jsonStruct)
	if err != nil {
		return err
	}
	return nil
}

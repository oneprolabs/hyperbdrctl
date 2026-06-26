package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

func requestBodyFromInput(file, inlineBody string) (interface{}, error) {
	if file != "" && inlineBody != "" {
		return nil, fmt.Errorf("file and body are mutually exclusive")
	}
	if file == "" && inlineBody == "" {
		return nil, nil
	}
	var raw []byte
	var err error
	if file != "" {
		raw, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	} else {
		raw = []byte(inlineBody)
	}
	var body interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body, nil
}

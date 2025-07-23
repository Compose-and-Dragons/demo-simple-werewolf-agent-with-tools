package helpers

import (
	"encoding/json"

	"github.com/openai/openai-go"
)

func MessageToMap(message openai.ChatCompletionMessageParamUnion) (map[string]string, error) {
	jsonData, err := message.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	stringMap := make(map[string]string)
	for key, value := range result {
		if str, ok := value.(string); ok {
			stringMap[key] = str
		}
	}

	return stringMap, nil
}
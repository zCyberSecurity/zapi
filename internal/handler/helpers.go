package handler

import "encoding/json"

func unmarshalJSON(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func errResp(msg string) map[string]any {
	return map[string]any{"error": msg}
}

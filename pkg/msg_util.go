package pkg

import (
	"ai-stream-bot/client/ai"
	"encoding/json"
)

func GetStrPoolTotalLength(strPool []ai.AiMessage) int {
	var total int
	for _, v := range strPool {
		bytes, _ := json.Marshal(v)
		total += len(string(bytes))
	}
	return total
}

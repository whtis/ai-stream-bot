package pkg

import (
	"encoding/json"
	"strings"
)

// ExtractTextFromJSON 从JSON格式中提取纯文本内容
// 支持处理飞书消息格式的JSON
func ExtractTextFromJSON(jsonStr string) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	return extractTextFromMap(data), nil
}

// extractTextFromMap 递归提取map中的文本内容
func extractTextFromMap(data map[string]interface{}) string {
	var result strings.Builder

	// 处理飞书消息格式
	if content, ok := data["content"]; ok {
		if contentArr, ok := content.([]interface{}); ok {
			for _, paragraph := range contentArr {
				if paragraphArr, ok := paragraph.([]interface{}); ok {
					for _, element := range paragraphArr {
						if elementMap, ok := element.(map[string]interface{}); ok {
							if text, ok := elementMap["text"].(string); ok {
								result.WriteString(text)
								result.WriteString(" ")
							}
						}
					}
					result.WriteString("\n")
				}
			}
			return strings.TrimSpace(result.String())
		}
	}

	// 通用处理：递归查找所有text字段
	for key, value := range data {
		if key == "text" {
			if text, ok := value.(string); ok {
				result.WriteString(text)
				result.WriteString(" ")
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			result.WriteString(extractTextFromMap(nestedMap))
		} else if nestedArr, ok := value.([]interface{}); ok {
			for _, item := range nestedArr {
				if itemMap, ok := item.(map[string]interface{}); ok {
					result.WriteString(extractTextFromMap(itemMap))
				}
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// ExtractTextFromFeishuMessage 专门处理飞书消息格式的JSON
func ExtractTextFromFeishuMessage(jsonStr string) (string, error) {
	var data struct {
		Title   string `json:"title"`
		Content [][]struct {
			Tag   string   `json:"tag"`
			Text  string   `json:"text"`
			Style []string `json:"style,omitempty"`
		} `json:"content"`
	}

	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	var result strings.Builder

	// 添加标题（如果有）
	if data.Title != "" {
		result.WriteString(data.Title)
		result.WriteString("\n\n")
	}

	// 处理内容
	for _, paragraph := range data.Content {
		for _, element := range paragraph {
			if element.Tag == "text" {
				result.WriteString(element.Text)
			}
		}
		result.WriteString("\n")
	}

	return strings.TrimSpace(result.String()), nil
}

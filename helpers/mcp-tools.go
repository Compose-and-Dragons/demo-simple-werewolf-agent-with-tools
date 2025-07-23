package helpers

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/openai/openai-go"
)

func ConvertMCPToolsToOpenAITools(tools *mcp.ListToolsResult) []openai.ChatCompletionToolParam {
	openAITools := make([]openai.ChatCompletionToolParam, len(tools.Tools))
	for i, tool := range tools.Tools {

		openAITools[i] = openai.ChatCompletionToolParam{
			Function: openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
				Parameters: openai.FunctionParameters{
					"type":       "object",
					"properties": tool.InputSchema.Properties,
					"required":   tool.InputSchema.Required,
				},
			},
		}
	}
	return openAITools
}

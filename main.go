package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/budgies-nest/budgie/helpers"
	"github.com/charmbracelet/huh"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"

	"werewolf-agent/ui"
)

type Werewolf struct {
	Health       float64
	Strength     float64
	Agility      float64
	Intelligence float64
}

func main() {

	werewolf := Werewolf{
		Health:       100,
		Strength:     80,
		Agility:      70,
		Intelligence: 60,
	}

	ui.Println(ui.Blue, strings.Repeat("=", 80))

	modelRunnerBaseUrl := os.Getenv("MODEL_RUNNER_BASE_URL")

	if modelRunnerBaseUrl == "" {
		panic("MODEL_RUNNER_BASE_URL environment variable is not set")
	}
	ui.Println(ui.Blue, "Model Runner Base URL:", modelRunnerBaseUrl)

	modelRunnerChatModel := os.Getenv("MODEL_RUNNER_CHAT_MODEL")

	if modelRunnerChatModel == "" {
		panic("MODEL_RUNNER_CHAT_MODEL environment variable is not set")
	}

	ui.Println(ui.Blue, "Model Runner Chat Model:", modelRunnerChatModel)

	modelRunnerToolsModel := os.Getenv("MODEL_RUNNER_TOOLS_MODEL")
	if modelRunnerToolsModel == "" {
		panic("MODEL_RUNNER_TOOLS_MODEL environment variable is not set")
	}

	ui.Println(ui.Blue, "Model Runner Tools Model:", modelRunnerToolsModel)

	ui.Println(ui.Blue, strings.Repeat("=", 80))

	systemInstructions, err := helpers.ReadTextFile("instructions.md")
	if err != nil {
		panic(err)
	}
	// NOTE: try without this
	systemToolsInstructions := ` 
	Your job is to understand the user prompt and decide if you need to use tools to run external commands.
	Ignore all things not related to the usage of a tool
	`

	systemToolsInstructionsForChat := ` 
	If you detect that the user prompt is related to a tool, 
	ignore this part and focus on the other parts.
	`

	characterSheet, err := helpers.ReadTextFile("character_sheet.md")
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	clientEngine := openai.NewClient(
		option.WithBaseURL(modelRunnerBaseUrl),
		option.WithAPIKey(""),
	)

	// Tools Completion parameters
	toolsCompletionParams := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemToolsInstructions),
		},
		ParallelToolCalls: openai.Bool(true),
		Tools:             toolsCatalog(),
		Model:             modelRunnerToolsModel,
		Temperature:       openai.Opt(0.0),
	}

	// Chat Completion parameters
	chatCompletionParams := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage("CONTEXT:\n" + characterSheet),
			openai.SystemMessage(systemInstructions),
			//openai.UserMessage(userQuestion), // NOTE: to be removed
		},
		Model:       modelRunnerChatModel,
		Temperature: openai.Opt(0.5),
	}

	type PromptConfig struct {
		StartingMessage            string
		ExplanationMessage         string
		PromptTitle                string
		ThinkingPrompt             string
		InterruptInstructions      string
		CompletionInterruptMessage string
		GoodbyeMessage             string
	}
	promptConfig := PromptConfig{
		StartingMessage:       "🐺 I'm an Werewolf",
		ExplanationMessage:    "Ask me anything about me. Type '/bye' to quit or Ctrl+C to interrupt responses.",
		PromptTitle:           "✋ Query",
		ThinkingPrompt:        "⏳",
		InterruptInstructions: "(Press Ctrl+C to interrupt)",
		//CompletionInterruptMessage: "⚠️ Response was interrupted\n",
		GoodbyeMessage: "🐺 Bye!",
	}

	//reader := bufio.NewScanner(os.Stdin)
	fmt.Println(promptConfig.StartingMessage)
	fmt.Println(promptConfig.ExplanationMessage)
	fmt.Println("\n🐺⛑️", werewolf.Health, "🧠", werewolf.Intelligence)

	for {
		fmt.Print(promptConfig.ThinkingPrompt)
		fmt.Println(promptConfig.InterruptInstructions)

		var userInput string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewText().
					Title(promptConfig.PromptTitle).
					Placeholder("Type your question here...").
					Value(&userInput).
					ExternalEditor(false),
			),
		)

		// Run the form
		err := form.Run()
		if err != nil {
			// TODO: handle error
		}

		// Trim whitespace
		userInput = strings.TrimSpace(userInput)

		// Check for empty input
		if userInput == "" {
			continue
		}

		// Check for /bye command
		if userInput == "/bye" {
			fmt.Println(promptConfig.GoodbyeMessage)
			break
		}

		// Completions here...
		// TOOLS DETECTION:
		fmt.Println("🚀 Starting tools detection...")

		// IMPORTANT: do not forget to set the user question in the params
		toolsCompletionParams.Messages = append(
			toolsCompletionParams.Messages,
			openai.UserMessage(userInput),
		)

		fmt.Println("⏳ Running tools completion...")
		// Make initial Tool completion request
		// TOOLS COMPLETION:
		completion, err := clientEngine.Chat.Completions.New(ctx, toolsCompletionParams)
		if err != nil {
			fmt.Printf("😡 Tools completion error: %v\n", err)
			continue
		}

		fmt.Println("🛠️ Tools completion received")
		toolCalls := completion.Choices[0].Message.ToolCalls

		firstCompletionResult := ""
		// Return early if there are no tool calls
		if len(toolCalls) == 0 {
			fmt.Println("✋ No function call")
			fmt.Println()
			//continue
		} else {
			// TOOL CALLS:
			firstCompletionResult = "RESULTS:\n"
			// Execute the tool calls

			toolsToCall := toolsImplementation(&werewolf)

			for _, toolCall := range toolCalls {
				var args map[string]any
				args, _ = JsonStringToMap(toolCall.Function.Arguments)

				result, err := toolsToCall[toolCall.Function.Name](args)
				if err != nil {
					fmt.Println("Unknown function call:", toolCall.Function.Name)
				}
				firstCompletionResult += result.(string) + "\n"
			}
			fmt.Println("🎉 Tools calls executed!")
		}

		fmt.Println("🤖 Starting chat completion...")
		fmt.Println(strings.Repeat("=", 80))

		// CHAT COMPLETION:
		chatCompletionParams.Messages = append(
			chatCompletionParams.Messages,
			openai.SystemMessage(firstCompletionResult), // NOTE: could be empty
			openai.SystemMessage(systemToolsInstructionsForChat),
			openai.UserMessage(userInput),
		)

		stream := clientEngine.Chat.Completions.NewStreaming(ctx, chatCompletionParams)

		for stream.Next() {
			chunk := stream.Current()
			// Stream each chunk as it arrives
			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				fmt.Print(chunk.Choices[0].Delta.Content)
			}
		}

		if err := stream.Err(); err != nil {
			fmt.Printf("😡 Stream error: %v\n", err)
		}

		fmt.Println(strings.Repeat("=", 80))
		fmt.Println("")

		fmt.Println("\n🐺⛑️", werewolf.Health, "🧠", werewolf.Intelligence)

		fmt.Println() // Add spacing between interactions
	}

}

func toolsCatalog() []openai.ChatCompletionToolParam {

	getHealth := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "get_health",
			Description: openai.String("Get the health of the Werewolf"),
			Parameters:  openai.FunctionParameters{},
		},
	}

	setHealth := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "set_health",
			Description: openai.String("Set the health of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"value": map[string]string{
						"type":        "number",
						"description": "The new health value for the Werewolf.",
					},
				},
				"required": []string{"value"},
			},
		},
	}

	increaseHealth := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "increase_health",
			Description: openai.String("Increase the health of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"amount": map[string]string{
						"type":        "number",
						"description": "The amount to increase the Werewolf's health by.",
					},
				},
				"required": []string{"amount"},
			},
		},
	}

	decreaseHealth := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "decrease_health",
			Description: openai.String("Decrease the health of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"amount": map[string]string{
						"type":        "number",
						"description": "The amount to decrease the Werewolf's health by.",
					},
				},
				"required": []string{"amount"},
			},
		},
	}

	getIntelligence := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "get_intelligence",
			Description: openai.String("Get the intelligence of the Werewolf"),
			Parameters:  openai.FunctionParameters{},
		},
	}

	setIntelligence := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "set_intelligence",
			Description: openai.String("Set the intelligence of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"value": map[string]string{
						"type":        "number",
						"description": "The new intelligence value for the Werewolf.",
					},
				},
				"required": []string{"value"},
			},
		},
	}

	increaseIntelligence := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "increase_intelligence",
			Description: openai.String("Increase the intelligence of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"amount": map[string]string{
						"type":        "number",
						"description": "The amount to increase the Werewolf's intelligence by.",
					},
				},
				"required": []string{"amount"},
			},
		},
	}

	decreaseIntelligence := openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        "decrease_intelligence",
			Description: openai.String("Decrease the intelligence of the Werewolf"),
			Parameters: openai.FunctionParameters{
				"type": "object",
				"properties": map[string]any{
					"amount": map[string]string{
						"type":        "number",
						"description": "The amount to decrease the Werewolf's intelligence by.",
					},
				},
				"required": []string{"amount"},
			},
		},
	}

	return []openai.ChatCompletionToolParam{
		getHealth, setHealth, increaseHealth, decreaseHealth,
		getIntelligence, setIntelligence, increaseIntelligence, decreaseIntelligence,
		// Add more tools as needed
	}
}

// TODO: check the arguments provided to the tool calls
func toolsImplementation(werewolf *Werewolf) map[string]func(map[string]any) (any, error) {
	return map[string]func(map[string]any) (any, error){
		"get_health": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: get_health with args:", arguments)
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's health is %f.", werewolf.Health), nil
		},
		"set_health": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: set_health with args:", arguments)
			newHealth := arguments["value"].(float64)
			werewolf.Health = newHealth
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's health has been set to %f.", werewolf.Health), nil
		},
		"increase_health": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: increase_health with args:", arguments)
			amount := arguments["amount"].(float64)
			werewolf.Health += amount
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's health has been increased by %f. New health is %f.", amount, werewolf.Health), nil
		},
		"decrease_health": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: decrease_health with args:", arguments)
			amount := arguments["amount"].(float64)
			werewolf.Health -= amount
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's health has been decreased by %f. New health is %f.", amount, werewolf.Health), nil
		},
		"get_intelligence": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: get_intelligence with args:", arguments)
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's intelligence is %f.", werewolf.Intelligence), nil
		},
		"set_intelligence": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: set_intelligence with args:", arguments)
			newIntelligence := arguments["value"].(float64)
			werewolf.Intelligence = newIntelligence
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's intelligence has been set to %f.", werewolf.Intelligence), nil
		},
		"increase_intelligence": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: increase_intelligence with args:", arguments)
			amount := arguments["amount"].(float64)
			werewolf.Intelligence += amount
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's intelligence has been increased by %f. New intelligence is %f.", amount, werewolf.Intelligence), nil
		},
		"decrease_intelligence": func(arguments map[string]any) (any, error) {
			fmt.Println("🔧 Executing tool call: decrease_intelligence with args:", arguments)
			amount := arguments["amount"].(float64)
			werewolf.Intelligence -= amount
			return fmt.Sprintf("TELL THIS TO THE USER: 🐺 The Werewolf's intelligence has been decreased by %f. New intelligence is %f.", amount, werewolf.Intelligence), nil
		},
	}
}

func JsonStringToMap(jsonString string) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

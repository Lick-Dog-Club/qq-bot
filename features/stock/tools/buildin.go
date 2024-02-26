package tools

import (
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

var BuildInPluginCurrentDatetime = Tool{
	Name: "CurrentDatetime",
	Define: openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "CurrentDatetime",
			Description: "get current datetime",
			Parameters: &jsonschema.Definition{
				Type: jsonschema.Object,
			},
		},
	},
}

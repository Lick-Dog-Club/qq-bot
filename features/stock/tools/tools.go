package tools

import (
	"github.com/sashabaranov/go-openai"
)

type Tool struct {
	Name   string
	Define openai.Tool
}

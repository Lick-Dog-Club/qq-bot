package tools

import (
	"errors"
	"fmt"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type ToolType string

const (
	ToolTypeBuildIn ToolType = "build-in"
	ToolTypePlugin  ToolType = "plugin"
)

var toolManagerInstance = &toolManager{list: map[string]Tool{}}

type Tool struct {
	Name   string
	Type   ToolType
	Define openai.Tool
}

type toolManager struct {
	mu   sync.RWMutex
	list map[string]Tool
}

func (tm *toolManager) Register(t Tool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if _, ok := tm.list[t.Name]; ok {
		panic(fmt.Sprintf("tool %s already defined", t.Name))
	}

	tm.list[t.Name] = t
}

func (tm *toolManager) GetPluginNameByFunctionName(name string) (Tool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, tool := range tm.list {
		if tool.Define.Function.Name == name {
			return tool, nil
		}
	}
	return Tool{}, errors.New("plugin not found")
}

func (tm *toolManager) GetByNames(withBuildIn bool, names ...string) (res []Tool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	for _, name := range names {
		if tool, ok := tm.list[name]; ok {
			res = append(res, tool)
		}
	}
	if withBuildIn {
		for idx, tool := range tm.list {
			if tool.Type == ToolTypeBuildIn {
				res = append(res, tm.list[idx])
			}
		}
	}
	return
}

func MustRegister(ts ...Tool) {
	for _, t := range ts {
		toolManagerInstance.Register(t)
	}
}

func GetByNames(withBuildIn bool, names ...string) (res []Tool) {
	return toolManagerInstance.GetByNames(withBuildIn, names...)
}
func GetPluginNameByFunctionName(name string) (Tool, error) {
	return toolManagerInstance.GetPluginNameByFunctionName(name)
}

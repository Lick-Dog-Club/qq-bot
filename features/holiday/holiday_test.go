package holiday

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	get := Get(2024)
	t.Log(get)
	temp.Execute(os.Stdout, map[string]any{
		"Days": get,
	})
}

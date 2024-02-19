package zuan

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"math/rand"
	"qq/bot"
	"qq/features"
	"sync"
)

//go:embed zuan.csv
var content []byte

var all []string
var once sync.Once

func init() {
	features.AddKeyword("zuan", "返回一句祖安/骂人的话", func(bot bot.Bot, content string) error {
		bot.Send(Get())
		return nil
	}, features.WithAIFunc(features.AIFuncDef{
		Properties: nil,
		Call: func(args string) (string, error) {
			return Get(), nil
		},
	}))
}

func Get() string {
	once.Do(func() {
		reader := csv.NewReader(bytes.NewReader(content))
		records, _ := reader.ReadAll()
		for _, record := range records {
			all = append(all, record[1])
		}
	})

	return all[rand.Int31n(int32(len(all)))]
}

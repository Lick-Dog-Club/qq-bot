package chart

import (
	"encoding/base64"
	"github.com/vicanso/go-charts/v2"
	"os"
	"qq/util/text2png"
)

type XY struct {
	X string
	Y float64
}

type LineChartInput struct {
	Base64    bool
	Path      string
	Width     int
	Height    int
	Title     string
	XLabel    string
	YLabel    string
	ShowLabel bool
	Lines     []struct {
		Name  string
		Items []XY
	}
}

func DrawLineChart(input LineChartInput) string {
	err := charts.InstallFont("font", text2png.FontBytes)
	if err != nil {
		panic(err)
	}
	font, _ := charts.GetFont("font")
	charts.SetDefaultFont(font)
	var (
		values       [][]float64
		legendLabels []string
		xAxisData    []string
	)
	for idx, line := range input.Lines {
		legendLabels = append(legendLabels, line.Name)
		var oneValue []float64
		for _, item := range line.Items {
			oneValue = append(oneValue, item.Y)
			if idx == 0 {
				xAxisData = append(xAxisData, item.X)
			}
		}
		values = append(values, oneValue)
	}

	var sl charts.SeriesList = make(charts.SeriesList, len(values))
	for idx, value := range values {
		sl[idx] = charts.Series{
			Type:  "line",
			Data:  charts.NewSeriesDataFromValues(value),
			Label: charts.SeriesLabel{Show: input.ShowLabel},
		}
	}
	render, err := charts.Render(charts.ChartOption{SeriesList: sl},
		charts.XAxisDataOptionFunc(xAxisData),
		charts.LegendLabelsOptionFunc(legendLabels, charts.PositionCenter),
		charts.TitleTextOptionFunc(input.Title),
	)
	buf, _ := render.Bytes()
	if input.Base64 {
		return base64.StdEncoding.EncodeToString(buf)
	}
	if input.Path != "" {
		writeFile(input.Path, buf)
	}
	return input.Path
}

func writeFile(path string, buf []byte) error {
	err := os.WriteFile(path, buf, 0600)
	if err != nil {
		return err
	}
	return nil
}

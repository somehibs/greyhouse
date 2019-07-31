package main

import (
	"log"
	"os"
	"encoding/csv"
	"time"
	"image/color"
	"strconv"
	//"bytes"

	"git.circuitco.de/self/greyhouse/presence"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	//"gonum.org/v1/plot/plotutil"
)

type AutoTicks struct {
}

func (a AutoTicks) Ticks(min, max float64) []plot.Tick {
	return []plot.Tick{plot.Tick{Value: 0}, plot.Tick{Value: 1}}
}

func main() {
	log.Print("Loading csv.")
	f, err := os.Open("motion.csv")
	pan(err)

	reader := csv.NewReader(f)
	pan(err)

	defer f.Close()
	log.Print("Loading header.")
	line, err := reader.Read()
	pan(err)

	log.Printf("Csv header: %s", line)
	log.Print("Loading all items.")
	lines, err := reader.ReadAll()
	pan(err)

	f.Close()

	plots := []string{"KITCHEN", "STUDY"}
	plot.SuggestedTicks = int(float64(len(lines))/float64(10))
	realPlot, err := plot.New()
	realPlot.Title.Text = "Motion plots"
	realPlot.X.Label.Text = "Time"
	realPlot.X.Tick.Marker = plot.TimeTicks{Format: "15:04:05"}
	realPlot.Y.Label.Text = "Triggers"
	realPlot.Y.Tick.Marker = AutoTicks{}

	//realPlot.Add(plotter.NewGrid())
	for _, plot := range plots {
		appendPlot(realPlot, lines, plot)
	}
	cent := 5000*vg.Centimeter
	err = realPlot.Save(cent, 3*vg.Centimeter, "motion.svg")
	pan(err)
	//generateHtml(plots)
}

func generateHtml(plots []string) {
	h := "<html><body>"
	for _, plot := range plots {
		h += "<img src=\""+plot+"_motion.svg\"/>"
	}
	h += "</body></html>"
	f, err := os.Create("motion.html")
	pan(err)
	defer f.Close()
	f.Write([]byte(h))
}

func appendPlot(p *plot.Plot, lines [][]string, room string) {
	log.Printf("Generating plot from %d lines.", len(lines))
	lmax := len(lines)//lmax := len(lines)/4
	lines = lines[:lmax]
	points := make(plotter.XYs, 0)
	for i := range lines {
		if i != 0 && i != len(lines) && lines[i][1] != room {
			continue
		}
		timeStr := lines[i][0]
		state := lines[i][3]
		stateInt, err  := strconv.Atoi(state)
		if lines[i][1] != room {
			stateInt = 0
		}
		if err != nil {
			log.Printf("Could not parse state: %s", state)
			continue
		}
		date, err := time.Parse(presence.MotionTimeFormat, timeStr)
		if err != nil {
			log.Printf("Could not parse time: %s", timeStr)
			continue
		}
		xy := plotter.XY{float64(date.Unix()), float64(stateInt)}
		points = append(points, xy)
	}

	plotLine, _, err := plotter.NewLinePoints(points)
	pan(err)
	if room == "KITCHEN" {
		log.Print("Changing line style to something more paletable")
		plotLine.LineStyle = draw.LineStyle{Width: vg.Points(1), Dashes: []vg.Length{}, DashOffs: 0, Color: color.RGBA{255,0,0,255}}
	}

	p.Add(plotLine)//, plotPoints)
	log.Printf("Rendering graph with %d points", len(points))
}

func pan(err error) {
	if err != nil {
		panic(err.Error())
	}
}

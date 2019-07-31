package main

import (
	"log"
	"os"
	"encoding/csv"
	"time"
	"strconv"
	//"bytes"

	"git.circuitco.de/self/greyhouse/presence"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	//"gonum.org/v1/plot/plotutil"
)

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

	log.Printf("Generating plot from %d lines.", len(lines))
	lmax := len(lines)/4
	lines = lines[:lmax]
	points := make(plotter.XYs, len(lines))
	for i := range lines {
		timeStr := lines[i][0]
		state := lines[i][3]
		stateInt, err  := strconv.Atoi(state)
		if err != nil {
			log.Printf("Could not parse state: %s", state)
			continue
		}
		date, err := time.Parse(presence.MotionTimeFormat, timeStr)
		if err != nil {
			log.Printf("Could not parse time: %s", timeStr)
			continue
		}
		points[i].X = float64(date.Unix())
		points[i].Y = float64(stateInt)
	}

	p, err := plot.New()
	p.Title.Text = "Motion plots"
	p.X.Label.Text = "Time"
	p.X.Tick.Marker = plot.TimeTicks{Format: presence.MotionTimeFormat}
	p.Y.Label.Text = "Triggers"

	p.Add(plotter.NewGrid())
	plotLine, _, err := plotter.NewLinePoints(points)
	pan(err)

	p.Add(plotLine)//, plotPoints)
	cent := 200*vg.Centimeter
	log.Printf("Rendering %+v graph with %d points", cent, len(points))
	err = p.Save(cent, 5*vg.Centimeter, "motion.png")
	pan(err)
}

func pan(err error) {
	if err != nil {
		panic(err.Error())
	}
}

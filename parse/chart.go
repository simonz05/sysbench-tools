package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/vdobler/chart"
	"github.com/vdobler/chart/imgg"
	"github.com/vdobler/chart/txtg"
)

type ChartWriter interface {
	Write(*Chart)
}

type Charter interface {
	Chart([]string) []*Chart
}

type Chart struct {
	c    chart.Chart
	name string
}

func (c *Chart) filename() string {
	return safeFilename(c.name)
}

type ImageWriter struct{}
type TermWriter struct{}

func (im *ImageWriter) Write(c *Chart) {
	os.MkdirAll("chart", os.ModePerm)

	fp, err := os.Create(path.Join("chart", c.filename()+".png"))

	if err != nil {
		panic(err)
	}

	defer fp.Close()

	img := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	bg := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	draw.Draw(img, img.Bounds(), bg, image.ZP, draw.Src)

	//row, col := d.Cnt/d.N, d.Cnt%d.N
	igr := imgg.AddTo(img, 0, 0, 1024, 768, color.RGBA{0xff, 0xff, 0xff, 0xff}, nil, nil)
	c.c.Plot(igr)
	png.Encode(fp, img)
}

func (tm *TermWriter) Write(c *Chart) {
	tgr := txtg.New(100, 40)
	c.c.Plot(tgr)
	os.Stdout.Write([]byte(tgr.String() + "\n\n\n"))
}

func TimeChart(xlabel, ylabel string, ylabels []string, data [][]chart.XYErrValue) *Chart {
	c := &chart.ScatterChart{Title: ylabel}
	c.XRange.Label = xlabel
	c.YRange.Label = ylabel
	//c.XRange.Time = true
	c.XRange.TicSetting.Mirror = 1
	c.YRange.ShowZero = true

	for i := 0; i < len(data); i++ {
		style := chart.AutoStyle(i, true)
		c.AddDataGeneric(ylabels[i], data[i], chart.PlotStyleLinesPoints, style)
	}

	return &Chart{
		c:    c,
		name: ylabel + "-" + xlabel + " time series",
	}
}

func resultsToChart(results []*Result, field string) []chart.XYErrValue {
	out := make([]chart.XYErrValue, 0, len(results))

	for _, v := range results {
		y := reflect.ValueOf(v).Elem().FieldByName(field)
		p := &chart.Point{
			Y: y.Float(),
			X: time.Duration(v.Duration).Seconds(),
		}
		out = append(out, p)
	}

	return out
}

type Point struct {
	X, Y float64
}

func (p Point) XVal() float64            { return p.X }
func (p Point) YVal() float64            { return p.Y }
func (p Point) XErr() (float64, float64) { return math.NaN(), math.NaN() }
func (p Point) YErr() (float64, float64) { return math.NaN(), math.NaN() }

func safeFilename(v string) string {
	v = strings.ToLower(v)

	re := regexp.MustCompile("[^a-z0-9]")
	v = re.ReplaceAllLiteralString(v, "-")

	re = regexp.MustCompile("[-]{2,}")
	v = re.ReplaceAllLiteralString(v, "-")
	return filepath.Clean(v)
}

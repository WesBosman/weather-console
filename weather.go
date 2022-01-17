package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
    "github.com/fogleman/gg"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/joho/godotenv"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

type WeatherFormat struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  int     `json:"pressure"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Visibility int `json:"visibility"`
	Wind       struct {
		Speed float64 `json:"speed"`
		Deg   int     `json:"deg"`
		Gust  float64 `json:"gust"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt  int `json:"dt"`
	Sys struct {
		Type    int    `json:"type"`
		ID      int    `json:"id"`
		Country string `json:"country"`
		Sunrise int64  `json:"sunrise"`
		Sunset  int64  `json:"sunset"`
	} `json:"sys"`
	Timezone int    `json:"timezone"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Cod      int    `json:"cod"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("Failed to load .env file")
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize tui: %v", err)
	}
	defer ui.Close()

	apiKey := os.Getenv("WEATHER_API_KEY")
	zipcode := os.Getenv("ZIPCODE")
	country := os.Getenv("COUNTRY")

	url := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?zip=%v,%v&appid=%v&units=imperial", zipcode, country, apiKey)
	resp, err := http.Get(url)

	if err != nil {
		log.Fatalf("Failed to get url: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatalf("Failed to read weather data: %v", err)
	}

	weatherJson := WeatherFormat{}
	jsonErr := json.Unmarshal(body, &weatherJson)

	if jsonErr != nil {
		log.Fatalf("Failed to parse json: %v", jsonErr)
	}

	weather := weatherJson.Weather[0]
	main := weather.Main
	icon := weather.Icon
	desc := weather.Description
	temp := weatherJson.Main.Temp
	lat := weatherJson.Coord.Lat
	lon := weatherJson.Coord.Lon
	windSpeed := weatherJson.Wind.Speed
	windDeg := weatherJson.Wind.Deg
	windGust := weatherJson.Wind.Gust
	sunrise := weatherJson.Sys.Sunrise
	sunset := weatherJson.Sys.Sunset

	iconUrl := fmt.Sprintf("http://openweathermap.org/img/wn/%v@2x.png", icon)
	resp, err = http.Get(iconUrl)

	if err != nil {
		log.Fatalf("Failed to get icon from url: %v", err)
	}

	image, _, err := image.Decode(resp.Body)

	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	img := widgets.NewImage(image)
	img.SetRect(0, 0, 20, 10)
	img.Title = main

	p := widgets.NewParagraph()
	p.Title = fmt.Sprintf("%v", weatherJson.Name)
	p.TitleStyle.Fg = ui.ColorYellow
	p.Text = fmt.Sprintf("%v, %v\n\nLat: %.4f, Lon: %.4f\nTemp: %.2f", main, desc, lat, lon, temp)
	p.SetRect(20, 0, 50, 5)

	p2 := widgets.NewParagraph()
	p2.Title = "Sunrise / Sunset"
	p2.TitleStyle.Fg = ui.ColorWhite
	p2.Text = fmt.Sprintf("Sunrise   %v\nSunset    %v", time.UnixMilli(sunrise).Format("3:04AM"), time.UnixMilli(sunset).Format("3:04PM"))
	p2.SetRect(20, 5, 50, 10)

	tempMin := weatherJson.Main.TempMin
	tempMax := weatherJson.Main.TempMax

	p3 := widgets.NewParagraph()
	p3.Title = "Temperature"
	p3.TitleStyle.Fg = ui.ColorWhite
	p3.Text = fmt.Sprintf("[    Max  %.2f°F](fg:red)\n   Temp  %.2f°F\n[    Min  %.2f°F](fg:blue)", tempMax, temp, tempMin)
	p3.SetRect(50, 0, 70, 5)

	table := widgets.NewTable()
	table.Title = "Wind"
	table.TitleStyle.Fg = ui.ColorWhite
	table.TextAlignment = ui.AlignCenter
	table.RowSeparator = false
	table.Rows = [][]string{
		{"[Speed](fg:cyan)", fmt.Sprintf("%.2f", windSpeed)},
		{"[Degrees](fg:cyan)", fmt.Sprintf("%d", windDeg)},
		{"[Gusts](fg:cyan)", fmt.Sprintf("%.2f", windGust)},
	}
	table.TextStyle = ui.NewStyle(ui.ColorBlue)
	table.SetRect(50, 5, 70, 10)

	p4 := widgets.NewParagraph()
	p4.Title = "Information"
	p4.Text = "Press Q (or ctrl+c) to quit."
	p4.SetRect(0, 10, 70, 13)

	windFloat := float64(windDeg)
	centerX := float64(50)
	centerY := float64(50)

	// Use gg to create a compass as an image and set it in the console
	arc := gg.NewContext(100, 100)
	arc.DrawCircle(centerX, centerY, 50)
	arc.SetRGB(128, 138, 159)
	arc.Fill()
	arc.DrawCircle(centerX, centerY, 35)
	arc.SetRGB(44, 73, 127)
	arc.Fill()
	arc.SetRGB(0, 0, 0)
	// Directions
	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic("")
	}
	face := truetype.NewFace(font, &truetype.Options{
		Size: 10,
	})
	arc.SetFontFace(face)
	// North
	north := "N"
	w, h := arc.MeasureString(north)
	arc.DrawRectangle(100, 180, w, h)
	arc.Stroke()
	arc.DrawStringAnchored(north, 50 - w/2, h, 0.0, 0.0)
	// East
	east := "E"
	w, h = arc.MeasureString(east)
	arc.DrawRectangle(100, 180, w, h)
	arc.Stroke()
	arc.DrawStringAnchored(east, 100 - w - 3, 50 + h/2, 0.0, 0.0)
	// South
	south := "S"
	w, h = arc.MeasureString(south)
	arc.DrawRectangle(100, 180, w, h)
	arc.Stroke()
	arc.DrawStringAnchored(south, 50 - w/2, 100 - 3, 0.0, 0.0)
	// Width
	west := "W"
	w, h = arc.MeasureString(west)
	arc.DrawRectangle(100, 180, w, h)
	arc.Stroke()
	arc.DrawStringAnchored(west, 3, 50 + h/2, 0.0, 0.0)
	// Draw a line for the compass direction
	// arc.Translate(centerX, centerY)
	// arc.Rotate(-gg.Radians(windFloat))
	// arc.MoveTo(0, 0)
	// arc.LineTo(0, 30)
	arc.Stroke()
	arc.SavePNG("compass.png")


	ui.Render(p, p2, p3, p4, img, table)

	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents

		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}

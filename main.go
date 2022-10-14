package main

import (
	"encoding/json"
	"fmt"
	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	clipB "github.com/atotto/clipboard"
	"github.com/changharry/iqdbAPI"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	go func() {
		w := app.NewWindow(app.Title("DanbooruTag"))
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

type (
	C = layout.Context
	D = layout.Dimensions
)

type ImageResult struct {
	Error  error
	Format string
	Image  image.Image
}

func loop(w *app.Window) error {
	expl := explorer.NewExplorer(w)
	var openBtn widget.Clickable
	var boilDurationInput widget.Editor
	var outputText widget.Editor
	var startButton widget.Clickable
	var startButton2 widget.Clickable
	var clipboardButton widget.Clickable
	var extension string
	var output string
	inputString := boilDurationInput.Text()
	inputString = strings.TrimSpace(inputString)
	th := material.NewTheme(gofont.Collection())
	imgChan := make(chan ImageResult)
	var img ImageResult
	var ops op.Ops
	var req string
	for {
		select {
		case img = <-imgChan:
			w.Invalidate()
		case e := <-w.Events():
			expl.ListenEvents(e)
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				if openBtn.Clicked() {
					go func() {
						file, err := expl.ChooseFile("png", "jpeg", "jpg")
						if err != nil {
							err = fmt.Errorf("failed opening image file: %w", err)
							imgChan <- ImageResult{Error: err}
							return
						}
						defer file.Close()
						imgData, format, err := image.Decode(file)
						extension = format
						imgChan <- ImageResult{Image: imgData, Format: format}
						home, _ := os.UserHomeDir()
						filePath, _ := filepath.Abs(home + "/req." + format)
						f, err := os.Create(filePath)
						if err != nil {
							panic(e)
						}
						defer f.Close()
						switch format {
						case "jpeg":
							if err := jpeg.Encode(f, imgData, nil); err != nil {
								return
							}
						case "png":
							if err := png.Encode(f, imgData); err != nil {
								return
							}
						}
					}()
				}
				if clipboardButton.Clicked() {
					err := clipB.WriteAll(output)
					if err != nil {
						return err
					}
				}
				if startButton.Clicked() {
					home, _ := os.UserHomeDir()
					req = iqdbAPI.API(iqdbAPI.Options{
						FilePath: home + "/req." + extension})
					var response iqdbAPI.Response
					err := json.Unmarshal([]byte(req), &response)
					if err != nil {
						return nil
					}
					body := response.Results
					for i := 0; i < len(body); i++ {
						if body[i].Head == "Best match" || body[i].Head == "Additional match" {
							if strings.Contains(body[i].Url, "danbooru") {
								regexp1, _ := regexp.Compile(`\bTags.*\b`)
								r := regexp1.FindString(body[i].Titles)
								r = strings.ReplaceAll(r, " ", ",")
								r = strings.ReplaceAll(r, "_", " ")
								output = r[6:]
							}
						}
					}

				}
				if startButton2.Clicked() {
					req = iqdbAPI.API(iqdbAPI.Options{
						Url: boilDurationInput.Text()})
					var response iqdbAPI.Response
					err := json.Unmarshal([]byte(req), &response)
					if err != nil {
						return nil
					}
					body := response.Results
					for i := 0; i < len(body); i++ {
						if body[i].Head == "Best match" || body[i].Head == "Additional match" {
							if strings.Contains(body[i].Url, "danbooru") {
								regexp1, _ := regexp.Compile(`\bTags.*\b`)
								r := regexp1.FindString(body[i].Titles)
								r = strings.ReplaceAll(r, " ", ",")
								r = strings.ReplaceAll(r, "_", " ")
								output = r[6:]
							}
						}
					}
				}
				layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(material.Button(th, &openBtn, "Open Image").Layout),
					layout.Flexed(1, func(gtx C) D {
						if img.Error == nil && img.Image == nil {
							return D{}
						} else if img.Error != nil {
							return material.H6(th, img.Error.Error()).Layout(gtx)
						}
						return widget.Image{
							Src: paint.NewImageOp(img.Image),
							Fit: widget.Contain,
						}.Layout(gtx)
					}),
					layout.Rigid(material.H6(th, "Image URL:").Layout),
					layout.Rigid(
						func(gtx C) D {
							ed := material.Editor(th, &boilDurationInput, "https://www.abc.com/xxx/123.jpg")
							boilDurationInput.SingleLine = true
							boilDurationInput.Alignment = text.Middle

							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(40),
							}
							border := widget.Border{
								Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								CornerRadius: unit.Dp(3),
								Width:        unit.Dp(2),
							}

							return margins.Layout(gtx,
								func(gtx C) D {
									return border.Layout(gtx, ed.Layout)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx C) D {
							ed := material.Editor(th, &outputText, "Output will show here")
							outputText.SetText(output)
							outputText.Alignment = text.Middle
							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(0),
							}
							border := widget.Border{
								Color:        color.NRGBA{R: 204, G: 204, B: 204, A: 255},
								CornerRadius: unit.Dp(3),
								Width:        unit.Dp(2),
							}

							return margins.Layout(gtx,
								func(gtx C) D {
									return border.Layout(gtx, ed.Layout)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx C) D {
							// margins
							margins := layout.Inset{
								Top:    unit.Dp(10),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(300),
								Left:   unit.Dp(300),
							}

							return margins.Layout(gtx,
								func(gtx C) D {
									var text string
									text = "copy"
									btn := material.Button(th, &clipboardButton, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx C) D {
							// margins
							margins := layout.Inset{
								Top:    unit.Dp(25),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(35),
								Left:   unit.Dp(35),
							}

							return margins.Layout(gtx,
								func(gtx C) D {
									var text string
									text = "Search using local image"
									btn := material.Button(th, &startButton, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
					layout.Rigid(
						func(gtx C) D {
							// margins
							margins := layout.Inset{
								Top:    unit.Dp(25),
								Bottom: unit.Dp(25),
								Right:  unit.Dp(35),
								Left:   unit.Dp(35),
							}
							return margins.Layout(gtx,
								func(gtx C) D {
									var text string
									text = "Search using url"
									btn := material.Button(th, &startButton2, text)
									return btn.Layout(gtx)
								},
							)
						},
					),
				)
				e.Frame(gtx.Ops)
			}
		}
	}
}

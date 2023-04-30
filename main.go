package main

import (
	"fmt"
	"gioui.org/io/system"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"image"
	"image/color"
	"log"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/hablullah/go-hijri"
	"github.com/mnadev/adhango/pkg/calc"
	"github.com/mnadev/adhango/pkg/data"
	"github.com/mnadev/adhango/pkg/util"
)

type PrayerTimes struct {
	Fajr, Sunrise, Dhuhr, Asr, Maghrib, Isha string
	TimeUntilNextPrayer                      time.Duration
}

type DateHolder struct {
	Day, Year int
	Month     string
}

var monthToName = map[int]string{
	1:  "Muḥarram",
	2:  "Safar",
	3:  "Rabi Al-Awwal",
	4:  "Rabi Al-Thani",
	5:  "Jumada Al-Ula",
	6:  "Jumada Al-Thaniyah",
	7:  "Rajab",
	8:  "Shaban",
	9:  "Ramadan",
	10: "Shawwal",
	11: "Du Al-Qadah",
	12: "Du Al-Hijjah",
}

func getCurrentDate() DateHolder {
	date, _ := hijri.CreateUmmAlQuraDate(time.Now())
	return DateHolder{
		Day:   int(date.Day),
		Month: monthToName[int(date.Month)],
		Year:  int(date.Year),
	}
}

func main() {
	go func() {
		width := unit.Dp(300)
		height := unit.Dp(410)
		options := []app.Option{
			app.Size(width, height),
			app.Title("Athan App"),
			app.Decorated(false),
			app.MinSize(width, height),
			app.MaxSize(width, height),
		}
		w := app.NewWindow(options...)
		if err := loop(w); err != nil {
			log.Fatal(err)
		}
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	th.TextSize = unit.Sp(14)
	th.Fg = color.NRGBA{R: 0x44, G: 0x44, B: 0x44, A: 0xFF}

	var ops op.Ops
	prayerTimes := getPrayerTimes()
	hijriDate := getCurrentDate()

	// Add a ticker to update the countdown every second
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)
				layoutUI(gtx, th, prayerTimes, hijriDate)
				e.Frame(gtx.Ops)
			case system.DestroyEvent:
				return e.Err
			}
		case <-ticker.C:
			// Refresh the UI when the ticker ticks
			prayerTimes = getPrayerTimes()
			w.Invalidate()
		}
	}
}

func getPrayerTimes() PrayerTimes {
	// Replace these coordinates with your location's coordinates
	latitude := 29.3117
	longitude := 47.4818

	// Set the method and parameters
	date := data.NewDateComponents(time.Now().In(time.UTC))
	params := calc.GetMethodParameters(calc.KUWAIT)
	params.Madhab = calc.SHAFI_HANBALI_MALIKI

	coords, err := util.NewCoordinates(latitude, longitude)
	if err != nil {
		fmt.Printf("Error %+v", err)
		panic(err)
	}

	prayerTimes, err := calc.NewPrayerTimes(coords, date, params)
	if err != nil {
		fmt.Printf("got error %+v", err)
		panic(err)
	}

	err = prayerTimes.SetTimeZone("Asia/Kuwait")
	if err != nil {
		fmt.Printf("got error %+v", err)
		panic(err)
	}

	prayer := prayerTimes.NextPrayerNow()
	var nextPrayer time.Time
	if prayer == calc.NO_PRAYER {
		temp, _ := calc.NewPrayerTimes(coords, data.NewDateComponents(time.Now().In(time.UTC).AddDate(0, 0, 1)), params)
		err := temp.SetTimeZone("Asia/Kuwait")
		if err != nil {
			fmt.Printf("got error %+v", err)
			panic(err)
		}
		nextPrayer = temp.Fajr
	} else {
		nextPrayer = prayerTimes.TimeForPrayer(prayer)
	}
	fmt.Printf("Next prayer is %v\n", prayer)
	fmt.Printf("Next prayer at %s\n", nextPrayer.Format("03:04 PM"))

	durationUntilNextPrayer := nextPrayer.Sub(time.Now()).Round(time.Second)

	return PrayerTimes{
		Fajr:                prayerTimes.Fajr.Format("03:04 PM"),
		Sunrise:             prayerTimes.Sunrise.Format("03:04 PM"),
		Dhuhr:               prayerTimes.Dhuhr.Format("03:04 PM"),
		Asr:                 prayerTimes.Asr.Format("03:04 PM"),
		Maghrib:             prayerTimes.Maghrib.Format("03:04 PM"),
		Isha:                prayerTimes.Isha.Format("03:04 PM"),
		TimeUntilNextPrayer: durationUntilNextPrayer,
	}
}

func layoutUI(gtx layout.Context, th *material.Theme, prayerTimes PrayerTimes, date DateHolder) layout.Dimensions {
	inset := layout.Inset{
		Top:    unit.Dp(10),
		Bottom: unit.Dp(10),
		Left:   unit.Dp(10),
		Right:  unit.Dp(10),
	}

	th.TextSize = unit.Sp(18)
	//th.Fg = color.NRGBA{R: 0x00, G: 0x00, B: 0x80, A: 0xFF} // New text color

	// Set the background color for the entire window
	bgColor := color.NRGBA{R: 0x42, G: 0xA5, B: 0xF5, A: 0xFF}
	paint.FillShape(gtx.Ops, bgColor, clip.Rect{Max: gtx.Constraints.Max}.Op())

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// First section
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutDateSection(gtx, th, date)
			})
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Separator
			return layoutSeparator(gtx, th)
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Second section
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutPrayerTimesSection(gtx, th, prayerTimes)
			})
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Separator
			return layoutSeparator(gtx, th)
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			// Third section
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layoutCounterSection(gtx, th, prayerTimes)
			})
		}))
	})

}

func layoutCounterSection(gtx layout.Context, th *material.Theme, prayerTimes PrayerTimes) layout.Dimensions {
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		th.TextSize = unit.Sp(20)
		th.Fg = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
		lbl := material.H6(th, fmt.Sprintf("%02d:%02d:%02d", int(prayerTimes.TimeUntilNextPrayer.Hours()), int(prayerTimes.TimeUntilNextPrayer.Minutes())%60, int(prayerTimes.TimeUntilNextPrayer.Seconds())%60))
		return layout.Stack{}.Layout(gtx, layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			rr := int(unit.Dp(4))
			clip.RRect{
				Rect: image.Rectangle{
					Max: image.Point{
						X: lbl.Layout(gtx).Size.X,
						Y: lbl.Layout(gtx).Size.Y,
					},
				},
				SE: rr, SW: rr, NW: rr, NE: rr,
			}.Op(gtx.Ops).Push(gtx.Ops)
			paint.Fill(gtx.Ops, color.NRGBA{R: 0x42, G: 0xA5, B: 0xF5, A: 0xFF})
			return layout.Dimensions{}
		}), layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			return lbl.Layout(gtx)
		}))
	})
}

func layoutDateSection(gtx layout.Context, th *material.Theme, date DateHolder) layout.Dimensions {
	th.TextSize = unit.Sp(18)
	th.Fg = color.NRGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF}

	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.H3(th, fmt.Sprintf("%02d", date.Day)).Layout(gtx)
		}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H4(th, date.Month).Layout(gtx)
				}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return material.H5(th, fmt.Sprintf("%04dh", date.Year)).Layout(gtx)
				}))
			})
		}))
	})
}

func layoutSeparator(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		line := image.Rect(0, 0, gtx.Constraints.Max.X, 1)
		borderColor := color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0x40}
		paint.FillShape(gtx.Ops, borderColor, clip.Rect(line).Op())
		return layout.Dimensions{
			Size: image.Point{X: gtx.Constraints.Max.X, Y: 1},
		}
	})
}

func layoutPrayerTimesSection(gtx layout.Context, th *material.Theme, prayerTimes PrayerTimes) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Fajr", prayerTimes.Fajr)
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Sunrise", prayerTimes.Sunrise)
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Dhuhr", prayerTimes.Dhuhr)
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Asr", prayerTimes.Asr)
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Maghrib", prayerTimes.Maghrib)
	}), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return layoutPrayerTimeRow(gtx, th, "Isha", prayerTimes.Isha)
	}))
}

func layoutPrayerTimeRow(gtx layout.Context, th *material.Theme, prayerName, prayerTime string) layout.Dimensions {
	marginY := unit.Dp(6)
	marginX := unit.Dp(0)
	inset := layout.Inset{
		Top:    marginY,
		Bottom: marginY,
		Left:   marginX,
		Right:  marginX,
	}
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			th.TextSize = unit.Sp(18)
			return material.Body1(th, prayerName).Layout(gtx)
		}), layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.E.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				th.TextSize = unit.Sp(18)
				return material.Body1(th, prayerTime).Layout(gtx)

			})
		}))
	})
}

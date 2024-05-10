package log

import (
	"github.com/anqiansong/ketty/console"
	"github.com/anqiansong/ketty/text"
)

var (
	DotConsole        *console.Console
	DoubleLineConsole *console.Console
	StarConsole       *console.Console
	PlusConsole       *console.Console
)

func init() {
	doubleLine := text.WithDoubleLine()
	DoubleLineConsole = console.NewConsole(console.WithTextOption(doubleLine))
	star := text.WithStarStyle()
	StarConsole = console.NewConsole(console.WithTextOption(star))
	plus := text.WithPlusStyle()
	PlusConsole = console.NewConsole(console.WithTextOption(plus))
	dot := text.WithDotStyle()
	DotConsole = console.NewConsole(console.WithTextOption(dot))
}

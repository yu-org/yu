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

func InitPrettyLog(output string) {
	doubleLine := text.WithDoubleLine()
	DoubleLineConsole = console.NewConsole(console.WithOutputDir(output), console.WithTextOption(doubleLine))
	doLog(DoubleLineConsole)

	star := text.WithStarStyle()
	StarConsole = console.NewConsole(console.WithOutputDir(output), console.WithTextOption(star))
	doLog(StarConsole)

	plus := text.WithPlusStyle()
	PlusConsole = console.NewConsole(console.WithOutputDir(output), console.WithTextOption(plus))
	doLog(PlusConsole)

	dot := text.WithDotStyle()
	DotConsole = console.NewConsole(console.WithOutputDir(output), console.WithTextOption(dot))
	doLog(DotConsole)
}

func doLog(c *console.Console) {
	c.DisableColor()
	c.DisableBorder()
	c.Close()
}

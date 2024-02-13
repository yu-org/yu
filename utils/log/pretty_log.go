package log

import (
	"github.com/anqiansong/ketty/console"
	"github.com/anqiansong/ketty/text"
)

func DoubleLineLog() *console.Console {
	style := text.WithDoubleLine()
	return console.NewConsole(console.WithTextOption(style))
}

func StarLog() *console.Console {
	style := text.WithStarStyle()
	return console.NewConsole(console.WithTextOption(style))
}

func PlusLog() *console.Console {
	style := text.WithPlusStyle()
	return console.NewConsole(console.WithTextOption(style))
}

func DotLog() *console.Console {
	style := text.WithDotStyle()
	return console.NewConsole(console.WithTextOption(style))
}

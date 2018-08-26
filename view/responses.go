package view

import (
	"fmt"
	"github.com/fatih/color"
)

func Success(emote string, msg string) {
	fmt.Fprintf(color.Output, "\n  %s %s", Green(emote), msg)
}

func Failure(emote string, msg string) {
	fmt.Fprintf(color.Output, "\n  %s %s", Red(emote), msg)
}

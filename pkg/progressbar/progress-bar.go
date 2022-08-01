package progressbar

import (
	"fmt"
	"github.com/gookit/color"
)

type Bar struct {
	cur        int64  // current progress
	rate       string // the actual progress bar to be printed
	graph      string // the fill value for progress bar
	totalMatch int64  // total of match
}

func (b *Bar) NewOption(graph string) {
	if graph == "" {
		b.graph = "="
	} else {
		b.graph = graph
	}
}

func (b *Bar) Play(cur int64) {
	if cur%50 == 0 {
		if b.cur <= 38 {
			b.cur += 1
			b.rate += b.graph
		} else {
			b.cur = 0
			b.rate = b.graph
		}
	}

	color.Printf("\r[<cyan>%-41s</>] ANALISADO [ <cyan>%d</> ]  MATCH [ <green>%d</> ]",
		b.rate+">", cur, b.totalMatch)
}

func (b *Bar) Match() {
	b.totalMatch++
}

func (b *Bar) Finish() {
	fmt.Println()
}

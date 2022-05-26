package progressbar

import (
	"fmt"
	"github.com/tsmweb/chasam/pkg/textcolor"
)

type Bar struct {
	percent    int64  // progress percentage
	cur        int64  // current progress
	total      int64  // total value for progress
	rate       string // the actual progress bar to be printed
	graph      string // the fill value for progress bar
	totalMatch int64  // total of match
}

func (b *Bar) NewOption(start, total int64, graph string) {
	b.cur = start
	b.total = total

	if graph == "" {
		b.graph = "="
	} else {
		b.graph = graph
	}

	b.percent = b.getPercent()

	for i := 0; i < int(b.percent); i += 2 {
		b.rate += b.graph // initial progress position
	}
}

func (b *Bar) getPercent() int64 {
	return int64((float32(b.cur) / float32(b.total)) * 100)
}

func (b *Bar) Play(cur int64) {
	b.cur = cur
	last := b.percent
	b.percent = b.getPercent()

	if b.percent != last && b.percent%2 == 0 {
		b.rate += b.graph
	}

	fmt.Printf("\r[%-61s]%3d%% %8d/%d %20s [ %d ]",
		textcolor.Cyan(b.rate),
		b.percent,
		b.cur,
		b.total,
		textcolor.Green("Match"),
		b.totalMatch)
}

func (b *Bar) Match() {
	b.totalMatch++
}

func (b *Bar) Finish() {
	fmt.Println()
}

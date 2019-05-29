package progressui

import (
	"log"
	"sync"

	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

type progressUI struct {
	progress *mpb.Progress
	bars     map[string]*mpb.Bar
}

func NewProgressUI(wg *sync.WaitGroup) *progressUI {
	progress := mpb.New(mpb.WithWaitGroup(wg))
	return &progressUI{
		progress: progress,
	}
}

func (p *progressUI) ShowError(err error) {
	log.Println(err)
}

func (p *progressUI) SetBarTotal(name string, total int) {
	if p.bars == nil {
		p.bars = make(map[string]*mpb.Bar)
	}
	var bar *mpb.Bar
	var ok bool
	if bar, ok = p.bars[name]; !ok {
		bar = p.progress.AddBar(int64(total),
			mpb.PrependDecorators(
				decor.Name(name+" ", decor.WCSyncWidth),
				decor.Percentage(decor.WCSyncWidth),
			),
		)
		p.bars[name] = bar
	}

	bar.SetTotal(int64(total), false)
}

func (p *progressUI) IncrementBar(name string, value int) {
	if bar, ok := p.bars[name]; ok {
		bar.IncrBy(value)
	}
}

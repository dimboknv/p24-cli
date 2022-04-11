package pb

import (
	"context"
	"io"

	"github.com/vbauerster/mpb/v6"
)

// Progress represents a wrapper of github.com/vbauerster/mpb Progress
// with extra features
type Progress struct {
	p    *mpb.Progress
	bars []Bar
}

// NewProgress creates new Progress instance. It's not possible to
// reuse instance after *Progress.Wait() method has been called
func NewProgress(w io.Writer) *Progress {
	return NewProgressWithContext(context.Background(), w)
}

// NewProgressWithContext creates new Progress instance. It's not possible to
// reuse instance after *Progress.Wait() method has been called
func NewProgressWithContext(ctx context.Context, w io.Writer) *Progress {
	return &Progress{
		p: mpb.NewWithContext(
			ctx,
			mpb.WithWidth(2),
			mpb.WithOutput(w),
		),
		bars: make([]Bar, 0),
	}
}

// AddBar to Progress
func (p *Progress) AddBar(bar Bar) {
	bar.AddToProgress(p.p)
	p.bars = append(p.bars, bar)
}

// Stop all bars with nil error
func (p *Progress) Stop() {
	for _, b := range p.bars {
		b.Stop()
	}
}

// Wait waits for all bars to complete
func (p *Progress) Wait() {
	p.p.Wait()
}

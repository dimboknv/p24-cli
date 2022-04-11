package pb

import (
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
)

type state int

const (
	running state = iota
	canceled
	done
)

// SpinBar represents a wrapper of github.com/vbauerster/mpb Bar
// with spinning progress bar that can be canceled, stopped with an error or nil
type SpinBar struct {
	err    error
	mpbBar *mpb.Bar
	title  string
	total  int64
	state  state
	mu     sync.RWMutex
}

// NewSpinBar returns new instance of SpinBar with title
func NewSpinBar(title string) *SpinBar {
	return &SpinBar{
		total: 1,
		title: title,
	}
}

// AddToProgress the spin bar
func (sb *SpinBar) AddToProgress(p *mpb.Progress) {
	mpbBar := p.AddSpinner(sb.total, mpb.SpinnerOnLeft,
		mpb.AppendDecorators(
			decor.Any(func(s decor.Statistics) string {
				err, title, canceled := sb.Error(), sb.Title(), sb.Canceled()
				switch {
				case canceled:
					return color.New(color.FgYellow).Sprintf("%s canceled", title)
				case err != nil:
					return color.New(color.FgRed).Sprintf("%s error: %s", title, err)
				case s.Completed:
					return color.New(color.FgGreen).Sprintf("%s done!", title)
				default:
					return color.New(color.FgYellow).Sprint(title)
				}
			}),
		),
		mpb.PrependDecorators(
			decor.Any(func(s decor.Statistics) string {
				err, canceled := sb.Error(), sb.Canceled()
				switch {
				case canceled:
					return color.New(color.FgYellow).Sprint(" ✕")
				case err != nil:
					return color.New(color.FgRed).Sprint(" ✕")
				case s.Completed:
					return color.New(color.FgGreen).Sprint(" ✓")
				default:
					return ""
				}
			}),
		),
		mpb.BarFillerClearOnComplete(),
	)

	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.mpbBar = mpbBar
}

// Cancel spinning. It will be printed into progress
func (sb *SpinBar) Cancel() {
	sb.mu.Lock()
	sb.setState(canceled)
	sb.mu.Unlock()

	sb.setMaxCurrent()
}

// StopWithErrMsg stop spinning and print errMsg into progress
func (sb *SpinBar) StopWithErrMsg(errMsg string) {
	sb.mu.Lock()
	sb.setState(done)
	sb.err = errors.New(strings.TrimRight(errMsg, "\r\n"))
	sb.mu.Unlock()

	sb.setMaxCurrent()
}

// Stop spinning
func (sb *SpinBar) Stop() {
	sb.mu.Lock()
	sb.setState(done)
	sb.mu.Unlock()

	sb.setMaxCurrent()
}

// Title returns bar title
func (sb *SpinBar) Title() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.title
}

// Error returns bar error
func (sb *SpinBar) Error() error {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.err
}

// Canceled returns if spin is canceled
func (sb *SpinBar) Canceled() bool {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return sb.state == canceled
}

func (sb *SpinBar) setState(s state) {
	if sb.state != running {
		return
	}
	sb.state = s
}

// setMaxCurrent sets progress' sb.mpbBar current to max value.
// It produces Shutdown event on sb.mpbBar.
// It is concurrency safe. Don`t use it with sb.mu possible get deadlock
// because Shutdown event can call Decorators(see sb.AddToProgress)
// that already has sb.mu calls
func (sb *SpinBar) setMaxCurrent() {
	if sb.mpbBar != nil {
		sb.mpbBar.SetCurrent(sb.total)
	}
}

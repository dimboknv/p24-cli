package pb

import "github.com/vbauerster/mpb/v6"

// Bar interface represents progress bar
type Bar interface {
	AddToProgress(p *mpb.Progress)
	StopWithErrMsg(errMsg string)
	Stop()
}

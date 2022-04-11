package cmd

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dimboknv/p24"
	p24http "github.com/dimboknv/p24-cli/http"
	"github.com/dimboknv/p24-cli/pb"
	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-retryablehttp"
)

// CommonOptionsCommander extends flags.Commander with SetCommon
// All commands should implement this interfaces
type CommonOptionsCommander interface {
	SetCommon(commonOpts CommonOpts)
	Execute(args []string) error
}

// BuildInfo about the executable
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// CommonOpts sets externally from main, shared across all commands
type CommonOpts struct {
	BuildInfo BuildInfo
	Debug     bool
}

// SetCommon satisfies CommonOptionsCommander interface and sets common option fields
// The method called by main for each command
func (opts *CommonOpts) SetCommon(commonOpts CommonOpts) {
	opts.BuildInfo = commonOpts.BuildInfo
	opts.Debug = commonOpts.Debug
}

func (opts *CommonOpts) waitSigterm(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer close(sigCh)
	defer signal.Stop(sigCh)
	select {
	case <-sigCh:
		log.Print("[WARN] interrupt signal")
	case <-ctx.Done():
	}
}

// CommonP24Opts struct with common options and funcs for p24 api commands
// nolint:govet // need to save command arguments order
type CommonP24Opts struct {
	ID          string        `long:"id" required:"true" description:"Merchant id"`
	Password    string        `long:"pass" required:"true" description:"Merchant password"`
	Card        string        `long:"card" required:"true" description:"Merchant card number"`
	HTTPTimeout time.Duration `long:"timeout" default:"90s" description:"http request timeout"`
	CommonOpts
}

func (opts *CommonP24Opts) makeP24Client() *p24.Client {
	retryHTTP := retryablehttp.NewClient()
	retryHTTP.HTTPClient.Timeout = opts.HTTPTimeout
	return p24.NewClient(p24.ClientOpts{
		Merchant: p24.Merchant{
			ID:   opts.ID,
			Pass: opts.Password,
		},
		HTTP: p24http.NewClient(
			p24http.WithRateLimiter(p24.NewRateLimiter()),
			p24http.WithRetryableHTTP(retryHTTP),
			p24http.WithLogger(log.Default()),
		),
		Log: log.Default(),
	})
}

func (opts *CommonP24Opts) makeProgressBar() *pb.Progress {
	var w io.Writer = os.Stderr
	if opts.Debug {
		w = nil
	}
	return pb.NewProgress(w)
}

package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dimboknv/p24-cli/cmd"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
)

// Opts with all cli commands and flags
// nolint:govet // need to save commands order
type Opts struct {
	BalanceCmd    cmd.BalanceCmd    `command:"balance" description:"Get card balance of specified merchant"`
	StatementsCmd cmd.StatementsCmd `command:"statements" description:"Load statements list for specified merchant and export it to a file/stdout"` // nolint
	VersionCmd    cmd.VersionCmd    `command:"version" description:"Show the 'p24' version information"`
	Debug         bool              `long:"debug" description:"Is debug mode?"`
}

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var opts Opts
	p := flags.NewParser(&opts, flags.HelpFlag|flags.PassDoubleDash)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		setupLogging(opts.Debug)

		c := command.(cmd.CommonOptionsCommander)
		c.SetCommon(cmd.CommonOpts{
			Debug: opts.Debug,
			BuildInfo: cmd.BuildInfo{
				Version: version,
				Commit:  commit,
				Date:    date,
			},
		})

		err := c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] command %q failed with: %+v", p.Active.Name, err)
		}
		return err
	}

	if _, err := p.Parse(); err != nil {
		// internal flags.Error error like 'option `-o1, --option1' uses the same long name as option `-o2, --option1'
		// wouldn't be printed by flags.Default
		w, code := os.Stderr, 1
		if flagsErr := (&flags.Error{}); errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			w, code = os.Stdout, 0
		}
		_, _ = fmt.Fprintln(w, err)
		os.Exit(code)
	}
}

func setupLogging(debug bool) {
	opts := []log.Option{log.Out(io.Discard), log.Err(io.Discard)}
	if debug {
		opts = []log.Option{log.Out(os.Stderr), log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces}
	}
	log.Setup(opts...)
	log.SetupStdLogger(opts...)
}

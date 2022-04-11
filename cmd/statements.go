package cmd

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path"
	"reflect"
	"sync"
	"time"

	"github.com/dimboknv/p24"
	"github.com/dimboknv/p24-cli/export"
	"github.com/dimboknv/p24-cli/pb"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	inputTimeLayout = "02.01.2006"
)

// StatementsCmd set of flags for getting p24 merchant statements list
// nolint:govet // need to save command arguments order
type StatementsCmd struct {
	CommonP24Opts
	StartDateStr    string         `long:"sd" required:"true" description:"Start date of statements date range with \"dd.mm.yyyy\" layout"`                                 // nolint
	EndDateStr      string         `long:"ed" required:"true" description:"End date of statements date range with \"dd.mm.yyyy\" layout"`                                   // nolint
	ExportFormatStr string         `short:"f" long:"format" default:"Card|Appcode|TranDate|Amount|CardAmount|Rest|Terminal|Description|," description:"Export format todo"` // nolint
	ExportEncoding  string         `short:"e" long:"encoding" default:"xml" choice:"xml" choice:"xlsx" description:"Export encoding"`
	OutputFilename  flags.Filename `short:"o" long:"out" description:"Export statements list to a file with specified extname encoding. If empty export to stdout with '-e' encoding"` // nolint

	startDate    time.Time
	endDate      time.Time
	exportFormat export.Format
}

// Execute gets statements list for specified merchant, entry point for "statements" command
func (cmd *StatementsCmd) Execute(_ []string) error {
	log.Printf("[INFO] \"statements\" command is started id=%s card=%s sd=%s ed=%s", cmd.ID, cmd.Card, cmd.StartDateStr, cmd.EndDateStr)

	if err := cmd.setup(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer cancel()
		cmd.waitSigterm(ctx)
	}()

	statements, err := cmd.getStatementsWithProgressBar(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get statements list")
	}

	if err = cmd.export(statements); err != nil {
		return errors.Wrapf(err, "failed to export")
	}

	log.Printf("[INFO] \"statements\" command succeeded terminated")
	return nil
}

func (cmd *StatementsCmd) export(statements p24.Statements) error {
	var w io.Writer = os.Stdout
	if cmd.OutputFilename != "" {
		f, err := os.OpenFile(string(cmd.OutputFilename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
		if err != nil {
			return errors.Wrapf(err, "failed to open file %q", cmd.OutputFilename)
		}
		log.Printf("[DEBUG] generation %q file", f.Name())
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("[WARN] failed to close file %s: %s", f.Name(), err)
			}
		}()
		w = f
	}

	// can skip error. it handled in Execute -> setup -> makeMarshaller
	exporter, _ := cmd.makeExporter(statements)
	log.Printf("[DEBUG] use %q marhsaller", reflect.TypeOf(exporter).String())
	log.Printf("[DEBUG] exporting statements to %q", w.(*os.File).Name())
	return exporter.Export(w, cmd.exportFormat)
}

func (cmd *StatementsCmd) getStatementsWithProgressBar(ctx context.Context) (p24.Statements, error) {
	eg, egCtx := errgroup.WithContext(ctx)
	statementsOpts := SplitStatementsDateRange(cmd.startDate, cmd.endDate, cmd.Card)
	client, prg, mu, res := cmd.makeP24Client(), cmd.makeProgressBar(), &sync.Mutex{}, p24.Statements{}

	for _, opts := range statementsOpts {
		opts := opts
		title := fmt.Sprintf("load: %s - %s", opts.StartDate.Format(inputTimeLayout), opts.EndDate.Format(inputTimeLayout))
		bar := pb.NewSpinBar(title)
		prg.AddBar(bar)

		eg.Go(func() error {
			log.Printf("[DEBUG] getting statements for %+v ...", opts)
			statements, err := client.GetStatements(egCtx, opts)

			// check context cancellation error and cancel bar if needed
			if errors.Is(err, context.Canceled) {
				log.Printf("[DEBUG] getting statements was canceled for %+v", opts)
				bar.Cancel()
				return err
			}

			// check http request timeout error and stop bar with timeout error if needed
			if urlErr := (&url.Error{}); errors.As(err, &urlErr) && urlErr.Timeout() {
				log.Printf("[DEBUG] getting statements was timeout for %+v", opts)
				bar.StopWithErrMsg("timeout")
				return err
			}

			if err != nil {
				bar.StopWithErrMsg(errors.Cause(err).Error())
				if p24Err := (&p24.Error{}); errors.As(err, &p24Err) {
					log.Printf("[DEBUG] getting statements failed for %+v: req: %s, resp: %s", opts, p24Err.Req, p24Err.Resp)
				} else {
					log.Printf("[DEBUG] getting statements failed for %+v: ", opts)
				}
				return err
			}

			log.Printf("[DEBUG] getting statements was succeeded for %+v", opts)
			bar.Stop()

			mu.Lock()
			res = mergeStatements(res, statements)
			mu.Unlock()
			return nil
		})
	}

	prg.Wait()
	if err := eg.Wait(); err != nil {
		return p24.Statements{}, err
	}
	return res, nil
}

func (cmd *StatementsCmd) setup() (err error) {
	cmd.exportFormat, err = export.MakeFormat(cmd.ExportFormatStr, export.DefaultFormatParser(p24.Statement{}))
	if err != nil {
		return errors.Wrapf(err, "invalid export format")
	}

	cmd.startDate, err = time.Parse(inputTimeLayout, cmd.StartDateStr)
	if err != nil {
		return errors.Wrapf(err, "invalid start date")
	}

	cmd.endDate, err = time.Parse(inputTimeLayout, cmd.EndDateStr)
	if err != nil {
		return errors.Wrapf(err, "invalid end date")
	}

	if err := p24.CheckCardNumber(cmd.Card); err != nil {
		return errors.Wrapf(err, "invalid card number")
	}

	if _, err := cmd.makeExporter(p24.Statements{}); err != nil {
		return errors.Wrapf(err, "invalid encoding")
	}

	return nil
}

func (cmd *StatementsCmd) makeExporter(statements p24.Statements) (export.Exporter, error) {
	encoding := cmd.ExportEncoding
	if ext := path.Ext(string(cmd.OutputFilename)); ext != "" {
		encoding = ext[1:]
	}

	switch encoding {
	case "xml":
		return export.NewXML(statements), nil
	case "xlsx":
		return export.NewXLSX(statements), nil
	default:
		return nil, errors.Errorf("%q is unsupported", encoding)
	}
}

// SplitStatementsDateRange splits given date range into 90 intervals
// and make StatementsOpts for each interval. Returns slice of StatementsOpts
func SplitStatementsDateRange(startDate, endDate time.Time, card string) []p24.StatementsOpts {
	days90 := 90 * 24 * time.Hour
	n := 1
	if dateRange := endDate.Sub(startDate); dateRange > days90 {
		n = int(math.Ceil(float64(dateRange) / float64(days90)))
	}
	sd, ed, opts := startDate, startDate.Add(days90), make([]p24.StatementsOpts, n)

	for i := 0; i < n; i++ {
		if ed.After(endDate) {
			ed = endDate
		}
		opts[i] = p24.StatementsOpts{
			StartDate:  sd,
			EndDate:    ed,
			CardNumber: card,
		}
		sd = sd.Add(days90 + 24*time.Hour)
		ed = ed.Add(days90 + 24*time.Hour)
	}
	return opts
}

func mergeStatements(statements ...p24.Statements) (res p24.Statements) {
	for _, s := range statements {
		res.Status = s.Status
		res.Debet += s.Debet
		res.Credit += s.Credit
		res.Statements = append(res.Statements, s.Statements...)
	}
	return res
}

package cmd

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"

	p24 "github.com/dimboknv/p24"
	"github.com/dimboknv/p24-cli/pb"
	log "github.com/go-pkgz/lgr"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
)

// BalanceCmd set of flags for getting p24 merchant card balance
// nolint:govet // need to save command arguments order
type BalanceCmd struct {
	CommonP24Opts
	Country        string         `short:"k" long:"country" required:"true" description:"Merchant card number"`
	ExportEncoding string         `short:"e" long:"encoding" default:"xml" choice:"xml" choice:"json" description:"Export encoding"`
	OutputFilename flags.Filename `short:"o" long:"out" description:"Export statements list to a file with specified extname encoding. If empty export to stdout with '-e' encoding"` // nolint
}

// Execute prints p24 merchant card balance, entry point for "balance" command
func (cmd *BalanceCmd) Execute(_ []string) error {
	log.Printf("[INFO] \"balance\" command started id=%s card=%s country=%s", cmd.ID, cmd.Card, cmd.Country)

	if err := cmd.setup(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer cancel()
		cmd.waitSigterm(ctx)
	}()

	cardBalance, err := cmd.getCardBalanceWithProgressBar(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get card balance")
	}

	if err := cmd.export(cardBalance); err != nil {
		return errors.Wrapf(err, "failed to %s export", cmd.ExportEncoding)
	}

	log.Printf("[INFO] \"balance\" command succeeded terminated")
	return nil
}

func (cmd *BalanceCmd) export(cardBalance p24.CardBalance) error {
	marhsal, err := cmd.makeMarshaller()
	if err != nil {
		return errors.Wrapf(err, "failed to %s export", cmd.ExportEncoding)
	}
	data, err := marhsal(cardBalance)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (cmd *BalanceCmd) getCardBalanceWithProgressBar(ctx context.Context) (p24.CardBalance, error) {
	client := cmd.makeP24Client()
	opts := p24.BalanceOpts{
		CardNumber: cmd.Card,
		Country:    cmd.Country,
	}

	prg := cmd.makeProgressBar()
	title := fmt.Sprintf("load: %s %s", opts.CardNumber, opts.Country)
	bar := pb.NewSpinBar(title)
	prg.AddBar(bar)

	log.Printf("[DEBUG] getting card balance for %+v", opts)
	cardBalance, err := client.GetCardBalance(ctx, opts)

	if errors.Is(err, context.Canceled) {
		log.Printf("[DEBUG] getting card balance was canceled for %+v", opts)
		bar.Cancel()
	}

	if urlErr := (&url.Error{}); errors.As(err, &urlErr) && urlErr.Timeout() {
		log.Printf("[DEBUG] getting card balance was timeout for %+v", opts)
		bar.StopWithErrMsg("timeout")
		return p24.CardBalance{}, err
	}

	if err != nil {
		bar.StopWithErrMsg(errors.Cause(err).Error())
		if p24Err := (&p24.Error{}); errors.As(err, &p24Err) {
			log.Printf("[DEBUG] getting card balance failed for %+v: req: %s, resp: %s", opts, p24Err.Req, p24Err.Resp)
		} else {
			log.Printf("[DEBUG] getting card balance failed for %+v ", opts)
		}
		return p24.CardBalance{}, err
	}

	log.Printf("[DEBUG] getting card balance was succeeded for %+v", opts)
	bar.Stop()

	prg.Wait()
	return cardBalance, nil
}

type marhsaller func(v interface{}) ([]byte, error)

func (cmd *BalanceCmd) makeMarshaller() (marhsaller, error) {
	var m marhsaller

	switch cmd.ExportEncoding {
	case "xml":
		m = xml.Marshal
	case "json":
		m = json.Marshal
	default:
		return nil, errors.Errorf("%q is unsupported", cmd.ExportEncoding)
	}
	return m, nil
}

func (cmd *BalanceCmd) setup() (err error) {
	if err := p24.CheckCardNumber(cmd.Card); err != nil {
		return errors.Wrapf(err, "invalid card number")
	}

	if _, err := cmd.makeMarshaller(); err != nil {
		return errors.Wrapf(err, "invalid encoding")
	}

	return nil
}

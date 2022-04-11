package cmd

import (
	"fmt"
	"os"
	"path"

	log "github.com/go-pkgz/lgr"
)

// VersionCmd set of flags for showing p24info version
type VersionCmd struct {
	CommonOpts
}

// Execute prints p24info cmd version, entry point for "version" command
func (cmd *VersionCmd) Execute(_ []string) error {
	log.Printf("[INFO] \"version\" started")

	fmt.Printf(
		"%s version: %s, commit: %s, date: %s\n",
		path.Base(os.Args[0]), cmd.BuildInfo.Version, cmd.BuildInfo.Commit, cmd.BuildInfo.Date,
	)

	log.Printf("[INFO] \"version\" command succeeded terminated")
	return nil
}

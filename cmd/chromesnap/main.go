package main

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

// set by -ldflags at build time
var (
	version   = "dev"
	buildDate = "unknown"
	goVersion = "unknown"
)

// globalFlags are the persistent flags shared across all commands.
type globalFlags struct {
	chrome           string
	remote           string
	noHeadless       bool
	proxy            string
	ignoreCertErrors bool
	timeout          time.Duration
	quiet            bool
}

func main() {
	gf := &globalFlags{}

	root := &cobra.Command{
		Use:   "chromesnap [URL]",
		Short: "Headless Chrome screenshot tool",
		Long:  "chromesnap takes screenshots of web pages using headless Chrome.\nRun without a subcommand to capture a single URL.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			sf := snapFlagsFrom(cmd)
			return runSnap(args[0], gf, sf)
		},
		SilenceUsage: true,
	}

	addGlobalFlags(root, gf)
	addSnapFlags(root)

	root.AddCommand(
		newSnapCmd(gf),
		newBatchCmd(gf),
		newVersionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func addGlobalFlags(cmd *cobra.Command, gf *globalFlags) {
	pf := cmd.PersistentFlags()
	pf.StringVar(&gf.chrome, "chrome", "", "Chrome/Chromium executable path")
	pf.StringVar(&gf.remote, "remote", "", "remote Chrome CDP URL (ws://host:port)")
	pf.BoolVar(&gf.noHeadless, "no-headless", false, "disable headless mode")
	pf.StringVar(&gf.proxy, "proxy", "", "proxy server URL")
	pf.BoolVar(&gf.ignoreCertErrors, "ignore-cert-errors", false, "ignore TLS certificate errors")
	pf.DurationVarP(&gf.timeout, "timeout", "t", 30*time.Second, "per-page timeout")
	pf.BoolVarP(&gf.quiet, "quiet", "q", false, "suppress log output")
}

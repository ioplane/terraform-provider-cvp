// terraform-provider-cvp — Arista CloudVision Portal provider.
//
// Prototype scoped to the Tier 1 resources from
// docs/notes/2026-07-16-i1-declarable-state.md.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/ioplane/terraform-provider-cvp/internal/provider"
)

// Injected at build time via -ldflags (see Taskfile.yml / .goreleaser.yml).
var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	var (
		debug       bool
		showVersion bool
	)
	flag.BoolVar(&debug, "debug", false, "run the provider in debug mode for provider development")
	flag.BoolVar(&showVersion, "version", false, "print version information and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("terraform-provider-cvp %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	err := providerserver.Serve(context.Background(), provider.New(version),
		providerserver.ServeOpts{
			Address: "registry.terraform.io/ioplane/cvp",
			Debug:   debug,
		})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

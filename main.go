// terraform-provider-cvp — Arista CloudVision Portal provider.
//
// Prototype scoped к Tier 1 resources из
// docs/notes/2026-07-16-i1-declarable-state.md.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/ioplane/terraform-provider-cvp/internal/provider"
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider in debug mode for provider development")
	flag.Parse()

	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/ioplane/cvp",
		Debug:   debug,
	})
	if err != nil {
		log.Fatal(err)
	}
}

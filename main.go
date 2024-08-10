package main

import (
	"context"
	_ "embed"
	"flag"
	"log"

	"github.com/folio-sec/terraform-provider-zoom/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Generate OpenAPI Clients
//go:generate go run github.com/ogen-go/ogen/cmd/ogen --target generated/api/zoomphone -package zoomphone --clean spec/ZoomPhoneAPISpec.json

// Run "go generate" to format example terraform files and generate the docs for the registry/website
//go:generate terraform fmt -recursive ./examples/

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name zoom

//go:embed version
var version string

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/folio-sec/zoom",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}

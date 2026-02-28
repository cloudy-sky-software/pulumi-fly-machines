// Copyright 2022, Cloudy Sky Software LLC.

package main

import (
	"context"
	_ "embed"

	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	providerSchemaGen "github.com/cloudy-sky-software/pulumi-fly-machines/provider/pkg/gen"
	providerVersion "github.com/cloudy-sky-software/pulumi-fly-machines/provider/pkg/version"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/pkg/errors"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

//go:embed openapi.yml
var openapiDocBytes []byte

// TemplateDir is the path to the base directory for code generator templates.
var TemplateDir string

// BaseDir is the path to the base pulumi-provider-template directory.
var BaseDir string

// Language is the SDK language.
type Language string

const (
	DotNet Language = "dotnet"
	Go     Language = "go"
	NodeJS Language = "nodejs"
	Python Language = "python"
	Schema Language = "schema"
)

func getOpenAPISpec(ctx context.Context, data []byte) *openapi3.T {
	doc, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		contract.Failf("Failed to load openapi.yml: %v", err)
	}

	err = providerSchemaGen.FixOpenAPIDoc(doc)
	if err != nil {
		panic(err)
	}

	// For the purposes of building a Pulumi schema, we don't care about
	// examples that may have been added to the spec by the cloud provider,
	// ignore those as those tend to have errors.
	if err := doc.Validate(ctx, openapi3.DisableExamplesValidation()); err != nil {
		contract.Failf("OpenAPI spec failed validation: %v", err)
	}

	return doc
}

func main() {
	ctx := context.Background()

	flag.Usage = func() {
		const usageFormat = "Usage: %s <language>"
		_, err := fmt.Fprintf(flag.CommandLine.Output(), usageFormat, os.Args[0])
		contract.IgnoreError(err)
		flag.PrintDefaults()
	}

	var version string
	flag.StringVar(&version, "version", providerVersion.Version, "the provider version to record in the generated code")

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		return
	}

	language := Language(args[0])

	switch language {
	case Schema:
		openAPIDoc := getOpenAPISpec(ctx, openapiDocBytes)

		schemaSpec, metadata, updatedOpenAPIDoc := providerSchemaGen.PulumiSchema(*openAPIDoc)
		providerDir := filepath.Join(".", "provider", "cmd", "pulumi-resource-fly-machines")
		mustWritePulumiSchema(schemaSpec, providerDir)

		// Write the metadata.json file as well.
		metadataBytes, _ := json.Marshal(metadata)
		mustWriteFile(providerDir, "metadata.json", metadataBytes)

		updatedOpenAPIDocBytes, _ := yaml.Marshal(updatedOpenAPIDoc)
		// Also copy the raw OpenAPI spec file to the provider dir.
		mustWriteFile(providerDir, "openapi_generated.yml", updatedOpenAPIDocBytes)
	default:
		panic(fmt.Sprintf("Unrecognized language '%s'", language))
	}
}

func mustWritePulumiSchema(pkgSpec schema.PackageSpec, outdir string) {
	pkgSpec.Version = ""
	schemaJSON, err := json.MarshalIndent(pkgSpec, "", "    ")
	if err != nil {
		panic(errors.Wrap(err, "marshaling Pulumi schema"))
	}
	mustWriteFile(outdir, "schema.json", schemaJSON)
}

func mustWriteFile(rootDir, filename string, contents []byte) {
	outPath := filepath.Join(rootDir, filename)

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		panic(err)
	}
	// nolint: gosec
	err := os.WriteFile(outPath, contents, 0644)
	if err != nil {
		panic(err)
	}
}

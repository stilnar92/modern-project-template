// Package openapi holds the server code that oapi-codegen generates from
// api/openapi.yaml. Do not edit openapi.gen.go: change the contract and run
// "mise run gen". CI fails when the generated code is not up to date.
package openapi

//go:generate go tool oapi-codegen -config config.yaml ../../../../api/openapi.yaml

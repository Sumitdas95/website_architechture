//go:build tools
// +build tools

// Package tools records build-time dependencies that aren't used by the
// library itself, but are tracked by go mod and required to lint and
// build the project.
package tools

import (
	_ "github.com/deliveroo/determinator-go/cmd/feature-codegen"
)

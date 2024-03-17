//go:build tools
// +build tools

package tools

import (
	// Use the following pattern for installing golang tools
	// https://marcofranssen.nl/manage-go-tools-via-go-modules/

	// For automatic reloading of go code
	_ "github.com/cespare/reflex"

	// For generating HTML
	_ "github.com/a-h/templ/cmd/templ"
)

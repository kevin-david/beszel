//go:build !linux

// Package bootid provides boot identity collection where supported.
package bootid

import "errors"

// Get reports that boot identity collection is unavailable on unsupported operating systems.
func Get() (string, error) {
	return "", errors.New("boot ID is unsupported on this operating system")
}

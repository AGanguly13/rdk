//go:build !no_cgo

// Package register registers all relevant vision models and also API specific functions
package register

import (
	// for vision models.
	_ "go.viam.com/rdk/services/vision/obstaclesdepth"
	_ "go.viam.com/rdk/services/vision/obstaclespointcloud"
)

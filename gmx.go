package main

import (
	"github.com/davecheney/gmx"
)

var (
	active   = gmx.NewGauge("socksie.connections.active")
	accepted = gmx.NewCounter("socksie.connections.accepted")
)

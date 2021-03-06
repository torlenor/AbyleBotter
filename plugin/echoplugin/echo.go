package echoplugin

import (
	"github.com/torlenor/redseligg/plugin"
)

// EchoPlugin is a plugin implementing a echo command
// which sends back all text received by that command to the
// User/Channel where the command was initiated.
//
// It serves as a working example how a simple plugin is implemented.
type EchoPlugin struct {
	plugin.RedseliggPlugin
	onlyOnWhisper bool
}

// SetOnlyOnWhisper tells the EchoPlugin that it should only
// echo on WHISPER type messages
func (p *EchoPlugin) SetOnlyOnWhisper(onlyOnWhisper bool) {
	p.onlyOnWhisper = onlyOnWhisper
}

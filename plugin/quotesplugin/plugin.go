package quotesplugin

import (
	"errors"
	"math/rand"
	"time"

	"github.com/torlenor/redseligg/botconfig"

	"github.com/torlenor/redseligg/plugin"
	"github.com/torlenor/redseligg/storagemodels"
)

const (
	PLUGIN_TYPE = "quotes"
	command     = "quote"
)

// ErrNoValidStorage is set when the provided storage does not implement the correct functions
var ErrNoValidStorage = errors.New("No valid storage set")

type randomizer interface {
	Intn(max int) int
}

// TODO (#36): The separate QuotesList storage is actually dumb, we should get rid of it
type quotesPluginWriter interface {
	StoreQuotesPluginQuote(botID, pluginID, identifier string, data storagemodels.QuotesPluginQuote) error
	StoreQuotesPluginQuotesList(botID, pluginID, identifier string, data storagemodels.QuotesPluginQuotesList) error
}

type quotesPluginReader interface {
	GetQuotesPluginQuote(botID, pluginID, identifier string) (storagemodels.QuotesPluginQuote, error)
	GetQuotesPluginQuotesList(botID, pluginID, identifier string) (storagemodels.QuotesPluginQuotesList, error)
}
type quotesPluginDeleter interface {
	DeleteQuotesPluginQuote(botID, pluginID, identifier string) error
}

type quotesPluginReaderWriter interface {
	quotesPluginReader
	quotesPluginWriter
	quotesPluginDeleter
}

// QuotesPlugin is a plugin that allows viewers or mods to add quotes and randomly fetch one.
type QuotesPlugin struct {
	plugin.RedseliggPlugin

	cfg config

	randomizer randomizer

	storage quotesPluginReaderWriter
}

// New returns a new QuotesPlugin
func New(pluginConfig botconfig.PluginConfig) (*QuotesPlugin, error) {
	cfg, err := parseConfig(pluginConfig)
	if err != nil {
		return nil, err
	}

	ep := QuotesPlugin{
		cfg:        cfg,
		randomizer: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	return &ep, nil
}

// getStorage returns the correct storage if it supports the necessary
// functions.
func (p *QuotesPlugin) getStorage() quotesPluginReaderWriter {
	if s, ok := p.API.GetStorage().(quotesPluginReaderWriter); ok {
		return s
	}
	return nil
}

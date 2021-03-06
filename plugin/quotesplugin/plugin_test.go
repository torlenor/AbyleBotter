package quotesplugin

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/torlenor/redseligg/botconfig"
	"github.com/torlenor/redseligg/model"
	"github.com/torlenor/redseligg/plugin"
	"github.com/torlenor/redseligg/storagemodels"
)

var commandQuote = command
var commandAdd = command + " add"
var commandRemove = command + " remove"
var commandHelp = command + "help"

func TestCreateQuotesPlugin(t *testing.T) {
	assert := assert.New(t)

	p, err := New(botconfig.PluginConfig{Type: "something"})
	assert.Error(err)
	assert.Nil(p)

	p, err = New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	p.SetAPI(&api)
}

func TestQuotesPlugin_OnRun(t *testing.T) {
	assert := assert.New(t)

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: nil}
	p.SetAPI(&api)

	assert.Equal("", api.LastLoggedError)
	p.OnRun()

	assert.Equal(ErrNoValidStorage.Error(), api.LastLoggedError)

	api.Reset()
	api.Storage = storage
	p.OnRun()
	assert.Equal("", api.LastLoggedError)
}

func TestQuotesPlugin_HelpTextAndInvalidCommands(t *testing.T) {
	assert := assert.New(t)

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "MESSAGE CONTENT",
		IsPrivate: false,
	}

	postToPlugin.Content = "!" + commandAdd
	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   fmt.Sprintf(helpText, api.GetCallPrefix()),
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "add", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!" + commandHelp
	p.OnCommand(commandQuote, "help", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!" + commandRemove
	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Content:   fmt.Sprintf(helpTextRemove, api.GetCallPrefix(), api.GetCallPrefix()),
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "remove", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!" + commandRemove + " something"
	p.OnCommand(commandQuote, "remove something", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)
}

func TestQuotesPlugin_AddQuote(t *testing.T) {
	assert := assert.New(t)

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"
	quote := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "MESSAGE CONTENT",
		IsPrivate: true,
	}

	postToPlugin.Content = "!" + commandAdd + " " + quote.Text
	p.OnCommand(commandAdd, quote.Text, postToPlugin)
	assert.Equal(false, api.WasCreatePostCalled)

	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "Successfully added quote #0", // because we are getting 0 entries for QuotesList from mockstorage
		IsPrivate: false,
	}
	postToPlugin.IsPrivate = false
	p.OnCommand(commandQuote, "add "+quote.Text, postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	if !assert.Equal(1, len(storage.StoredQuotes)) {
		t.FailNow()
	}
	if !assert.Equal(1, len(storage.StoredQuotesList)) {
		t.FailNow()
	}

	actualData := storage.StoredQuotes[0]
	assert.Equal(pluginID, actualData.PluginID)
	assert.Greater(len(actualData.Identifier), 0)
	assert.Equal(quote, actualData.Data)

	actualList := storage.StoredQuotesList[0]
	assert.Equal(pluginID, actualList.PluginID)
	assert.Equal(identFieldList, actualList.Identifier)
	assert.Equal(1, len(actualList.Data.UUIDs))
}

func TestQuotesPlugin_AddQuoteFail(t *testing.T) {
	assert := assert.New(t)

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"
	quote := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	storage.ErrorToReturn = fmt.Errorf("Some error")
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "MESSAGE CONTENT",
		IsPrivate: false,
	}

	postToPlugin.Content = "!" + commandAdd + " " + quote.Text
	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "Error storing quote. Try again later!",
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "add "+quote.Text, postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)
}

func TestQuotesPlugin_GetQuote(t *testing.T) {
	assert := assert.New(t)

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"
	quote := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}

	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "!" + commandQuote,
		IsPrivate: false,
	}

	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "No quotes found. Use the command `!" + command + " add <your quote>` to add a new one.",
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	storage.QuoteDataToReturn = quote
	storage.QuotesListDataToReturn = storagemodels.QuotesPluginQuotesList{
		UUIDs: []string{"some identifier"},
	}

	year, month, day := now().Date()
	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Content:   fmt.Sprintf(`1. "%s" - %d-%d-%d, added by %s`, quote.Text, year, month, day, postToPlugin.User.Name),
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)
}

func TestQuotesPlugin_GetQuote_Number(t *testing.T) {
	assert := assert.New(t)

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"

	quote2 := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some other quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	storage.QuoteDataToReturn = quote2
	storage.QuotesListDataToReturn = storagemodels.QuotesPluginQuotesList{
		UUIDs: []string{"some identifier", "some other identifier"},
	}

	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "!" + commandQuote + " 2",
		IsPrivate: false,
	}

	year, month, day := now().Date()
	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   fmt.Sprintf(`2. "%s" - %d-%d-%d, added by %s`, quote2.Text, year, month, day, postToPlugin.User.Name),
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "2", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)
}

func TestQuotesPlugin_RemoveQuote_OnlyMods(t *testing.T) {
	assert := assert.New(t)

	userName := "SOME USER NAME"

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"
	botID := "SOME BOT ID"
	quote := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID
	p.BotID = botID

	p.cfg.OnlyMods = true
	p.cfg.Mods = append(p.cfg.Mods, userName)

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	storage.QuoteDataToReturn = quote
	storage.QuotesListDataToReturn = storagemodels.QuotesPluginQuotesList{
		UUIDs: []string{"some identifier", "some other identifier"},
	}
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "SOME OTHER USER NAME"},
		Content:   "!" + commandRemove + " 2",
		IsPrivate: false,
	}

	p.OnCommand(commandQuote, "remove 2", postToPlugin)
	assert.Equal(false, api.WasCreatePostCalled)

	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "Successfully removed quote #2",
		IsPrivate: false,
	}
	postToPlugin.User.Name = userName
	p.OnCommand(commandQuote, "remove 2", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	if !assert.Equal(1, len(storage.StoredQuotesList)) {
		t.FailNow()
	}

	assert.Equal(botID, storage.LastDeleted.BotID)
	assert.Equal(pluginID, storage.LastDeleted.PluginID)
	assert.Equal("some other identifier", storage.LastDeleted.Identifier)
}

func TestQuotesPlugin_RemoveQuote(t *testing.T) {
	assert := assert.New(t)

	// Inject a new time.Now()
	now = func() time.Time {
		layout := "2006-01-02T15:04:05.000Z"
		str := "2018-12-22T13:00:00.000Z"
		t, _ := time.Parse(layout, str)
		return t
	}

	pluginID := "SOME_PLUGIN_ID"
	botID := "SOME BOT ID"
	quote := storagemodels.QuotesPluginQuote{
		Author:    "USER 1",
		Added:     now(),
		AuthorID:  "SOME USER ID",
		ChannelID: "CHANNEL ID",
		Text:      "some quote",
	}

	p, err := New(botconfig.PluginConfig{Type: PLUGIN_TYPE})
	assert.NoError(err)
	assert.Equal(nil, p.API)

	p.PluginID = pluginID
	p.BotID = botID

	storage := &MockStorage{}
	api := plugin.MockAPI{Storage: storage}
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME OTHER USER ID", Name: "USER 1"},
		Content:   "!" + commandRemove + " 2",
		IsPrivate: false,
	}

	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "Quote #2 not found",
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "remove 2", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	storage.QuoteDataToReturn = quote
	storage.QuotesListDataToReturn = storagemodels.QuotesPluginQuotesList{
		UUIDs: []string{"some identifier", "some other identifier"},
	}

	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Content:   "Successfully removed quote #2",
		IsPrivate: false,
	}
	p.OnCommand(commandQuote, "remove 2", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	if !assert.Equal(1, len(storage.StoredQuotesList)) {
		t.FailNow()
	}

	assert.Equal(botID, storage.LastDeleted.BotID)
	assert.Equal(pluginID, storage.LastDeleted.PluginID)
	assert.Equal("some other identifier", storage.LastDeleted.Identifier)
}

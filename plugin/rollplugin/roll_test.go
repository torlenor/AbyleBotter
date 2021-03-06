package rollplugin

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/torlenor/redseligg/model"
	"github.com/torlenor/redseligg/plugin"
)

type mockRandomizer struct{}

func (r mockRandomizer) random(int) int {
	return 123
}

func TestCreateRollPlugin(t *testing.T) {
	assert := assert.New(t)

	p, err := New()
	assert.NoError(err)
	assert.Equal(nil, p.API)

	api := plugin.MockAPI{}
	p.SetAPI(&api)
}

func TestRollPlugin_OnCommand(t *testing.T) {
	assert := assert.New(t)

	p := RollPlugin{
		randomizer: mockRandomizer{},
	}
	assert.Equal(nil, p.API)

	api := plugin.MockAPI{}
	p.SetAPI(&api)

	postToPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "MESSAGE CONTENT",
		IsPrivate: false,
	}

	api.Reset()
	postToPlugin.Content = "!roll"
	expectedPostFromPlugin := model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "<@" + postToPlugin.User.ID + "> rolled *" + strconv.Itoa(123) + "* in [0,100]",
		IsPrivate: false,
	}
	p.OnCommand("roll", "", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!roll 1000"
	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "<@" + postToPlugin.User.ID + "> rolled *" + strconv.Itoa(123) + "* in [0,1000]",
		IsPrivate: false,
	}
	p.OnCommand("roll", "1000", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!roll -1"
	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "Number must be > 0",
		IsPrivate: false,
	}
	p.OnCommand("roll", "-1", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)

	api.Reset()
	postToPlugin.Content = "!roll sdsadsad"
	expectedPostFromPlugin = model.Post{
		ChannelID: "CHANNEL ID",
		Channel:   "SOME CHANNEL",
		User:      model.User{ID: "SOME USER ID", Name: "USER 1"},
		Content:   "Not a number",
		IsPrivate: false,
	}
	p.OnCommand("roll", "sdsadsad", postToPlugin)
	assert.Equal(true, api.WasCreatePostCalled)
	assert.Equal(expectedPostFromPlugin, api.LastCreatePostPost)
}

package twitch

import (
	"fmt"

	"github.com/gorilla/websocket"
	"gopkg.in/irc.v3"

	"github.com/torlenor/redseligg/model"
	"github.com/torlenor/redseligg/utils"
)

var version string

// GetUsers a list of users based on search options.
func (b *Bot) GetUsers() ([]model.User, error) { return nil, nil }

// GetUser gets a user.
func (b *Bot) GetUser(userID string) (model.User, error) { return model.User{}, nil }

// GetUserByUsername gets a user by their username.
func (b *Bot) GetUserByUsername(name string) (model.User, error) { return model.User{}, nil }

// GetChannel gets a channel.
func (b *Bot) GetChannel(channelID string) (model.Channel, error) { return model.Channel{}, nil }

// GetChannelByName gets a channel by its name, given a team id.
func (b *Bot) GetChannelByName(name string) (model.Channel, error) { return model.Channel{}, nil }

// CreatePost creates a post.
func (b *Bot) CreatePost(post model.Post) (model.PostResponse, error) {
	ircMessage := irc.Message{
		Command: "PRIVMSG",
		Params:  []string{post.ChannelID, convertMessageFromRedseligg(post.Content)},
	}
	err := b.ws.SendMessage(websocket.TextMessage, []byte(ircMessage.String()))
	if err != nil {
		return model.PostResponse{}, fmt.Errorf("Could not send message: %s", err)
	}

	return model.PostResponse{}, nil
}

// UpdatePost updates a post.
func (b *Bot) UpdatePost(messageID model.MessageIdentifier, newPost model.Post) (model.PostResponse, error) {
	return model.PostResponse{}, fmt.Errorf("Not supported")
}

// DeletePost deletes a post.
func (b *Bot) DeletePost(messageID model.MessageIdentifier) (model.PostResponse, error) {
	return model.PostResponse{}, fmt.Errorf("Not supported")
}

// GetReaction gives back the platform specific string for a reaction, e.g., one -> :one:
func (b *Bot) GetReaction(reactionName string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

// LogTrace writes a log message to the server log file.
func (b *Bot) LogTrace(msg string) {
	log.Tracef("Error from plugin: %s", msg)
}

// LogDebug writes a log message to the server log file.
func (b *Bot) LogDebug(msg string) {
	log.Debugf("From plugin: %s", msg)
}

// LogInfo writes a log message to the server log file.
func (b *Bot) LogInfo(msg string) {
	log.Infof("From plugin: %s", msg)
}

// LogWarn writes a log message to the server log file.
func (b *Bot) LogWarn(msg string) {
	log.Warnf("From plugin: %s", msg)
}

// LogError writes a log message to the server log file.
func (b *Bot) LogError(msg string) {
	log.Errorf("From plugin: %s", msg)
}

// GetVersion returns the version of the server.
func (b *Bot) GetVersion() string {
	return utils.Version().Get() + " (" + utils.Version().GetCompTime() + ")"
}

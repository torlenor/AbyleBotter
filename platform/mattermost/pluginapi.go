package mattermost

import (
	"fmt"

	"github.com/torlenor/abylebotter/model"
)

// RegisterCommand registers a custom slash or ! command, depending on what the bot supports.
func (b *Bot) RegisterCommand(command string) error { return nil }

// UnregisterCommand unregisters a command previously registered via RegisterCommand.
func (b *Bot) UnregisterCommand(command string) error { return nil }

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
func (b *Bot) CreatePost(post model.Post) error {

	if post.IsPrivate {
		err := b.sendWhisper(post.UserID, post.Content)
		if err != nil {
			return fmt.Errorf("Error sending whisper: %s", err)
		}
	} else {
		err := b.sendMessage(post.ChannelID, post.Content)
		if err != nil {
			return fmt.Errorf("Error sending message: %s", err)
		}
	}

	return nil
}

// LogTrace writes a log message to the server log file.
func (b *Bot) LogTrace(msg string) {}

// LogDebug writes a log message to the server log file.
func (b *Bot) LogDebug(msg string) {}

// LogInfo writes a log message to the server log file.
func (b *Bot) LogInfo(msg string) {}

// LogWarn writes a log message to the server log file.
func (b *Bot) LogWarn(msg string) {}

// LogError writes a log message to the server log file.
func (b *Bot) LogError(msg string) {}
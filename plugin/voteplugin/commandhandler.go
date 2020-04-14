package voteplugin

import (
	"regexp"
	"strings"

	"github.com/torlenor/abylebotter/model"
)

const (
	helpText = "Type '!vote What is the best color? [Red, Green, Blue]' to start a new giveaway.\nYou can omit the custom options in the [...] to initiate a simple Yes/No vote."
)

func (p *VotePlugin) returnHelp(channelID string) {
	p.returnMessage(channelID, helpText)
}

func (p *VotePlugin) returnMessage(channelID, msg string) {
	post := model.Post{
		ChannelID: channelID,
		Content:   msg,
	}
	p.API.CreatePost(post)
}

func (p *VotePlugin) postAndStartVote(vote *vote) {
	post := vote.getCurrentPost()

	msgID, err := p.API.CreatePost(post)
	if err != nil {
		p.API.LogError("Something went wrong in creating the Vote message: " + err.Error())
		p.returnMessage(post.ChannelID, "Sorry to inform you, but we failed to create the Vote! Please try again later.")
		return
	}

	vote.messageIdent = msgID.PostedMessageIdent
}

func (p *VotePlugin) updatePost(vote *vote) {
	post := vote.getCurrentPost()

	_, err := p.API.UpdatePost(vote.messageIdent, post)
	if err != nil {
		p.API.LogError("Something went wrong in updating the Vote message: " + err.Error())
		return
	}
}

func (p *VotePlugin) extractDescriptionAndOptions(fullText string) (string, []string) {
	re := regexp.MustCompile(`!vote ([^\[\]]*)\s?(\[([^\[\]]*)])?`)
	const captureGroupDescription = 1
	const captureGroupOptions = 3

	matches := re.FindAllStringSubmatch(fullText, -1)

	if matches == nil || len(matches) < 1 {
		return "", []string{}
	} else if len(matches) > 1 {
		p.API.LogWarn("VotePlugin: extractDescriptionAndOptions matched more than one occurrence")
	}

	var options []string
	if len(matches[0]) > 3 && len(matches[0][captureGroupOptions]) > 0 {
		options = strings.Split(matches[0][captureGroupOptions], ",")
		for i := range options {
			options[i] = strings.Trim(options[i], " ")
			options[i] = strings.Trim(options[i], ",")
		}
	}

	return strings.Trim(matches[0][captureGroupDescription], " "), options
}

// onCommandVoteStart starts a new vote with the settings extracted
// from the received !vote command.
// Note: The command requires a valid !vote command. This check
// shall be performed at post retrieval.
func (p *VotePlugin) onCommandVoteStart(post model.Post) {
	description, options := p.extractDescriptionAndOptions(post.Content)
	if len(options) == 0 {
		// if empty, add Yes/No defaults
		options = []string{"Yes", "No"}
	}

	p.votesMutex.Lock()
	defer p.votesMutex.Unlock()
	nVote := newVote(voteSettings{
		ChannelID: post.ChannelID,
		Text:      description,
		Options:   options,
	})

	p.postAndStartVote(&nVote)
	p.runningVotes[nVote.Settings.Text] = &nVote
}

func (p *VotePlugin) onCommandVoteEnd(post model.Post) {
	cont := strings.Split(post.Content, " ")
	args := cont[1:]

	p.votesMutex.Lock()
	defer p.votesMutex.Unlock()
	description := strings.Join(args, " ")
	if v, ok := p.runningVotes[description]; ok {
		if v.messageIdent.Channel == post.ChannelID {
			v.end()
			p.updatePost(v)
			delete(p.runningVotes, description)
			return
		}
	}

	p.returnMessage(post.ChannelID, "No vote running with that description in this channel. Use the !vote command to start a new one.")
}

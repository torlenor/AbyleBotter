# AbyleBotter

[![Build Status](https://travis-ci.org/torlenor/abylebotter.svg?branch=master)](https://travis-ci.org/torlenor/abylebotter)
[![Coverage Status](https://coveralls.io/repos/github/torlenor/AbyleBotter/badge.svg?branch=master)](https://coveralls.io/github/torlenor/AbyleBotter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/torlenor/AbyleBotter)](https://goreportcard.com/report/github.com/torlenor/AbyleBotter)
[![Docker](https://img.shields.io/docker/pulls/hpsch/abylebotter.svg)](https://hub.docker.com/r/hpsch/abylebotter/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](/LICENSE)

## Description

This is AbyleBotter, an extensible Chat Bot for various platforms.

At the moment the Bot is in a proof of concept/API/interface development phase with very limited functional use.

## Supported plattforms

These plattforms are current supported (at least with the functionality to send and receive messages):

### Discord

### Matrix

### Mattermost

### Slack

## Releases

For releases binaries for Linux, Windows and Mac are provided. Check out the respective section on GitHub.

## How to build from source

### Requirements

- Go >= 1.12

### Building

Clone the sources from GitHub and compile AbyleBotter with

```
make deps
make
```

and optionally

```
make test
```

to run tests.

## How to run it

Independent of the way you obtain it, you have to configure the bot first and it is necessary to have a registered bot account for the service you want to use. 

- Discord: Please take a look at https://discordapp.com/developers/docs/intro on how to set up a bot user and generate the required authentication token.
- Matrix: For Matrix it is simpler, just create a user for the bot on your preferred Matrix server.
- Mattermost: For Mattermost a username and password with the necessary rights on the specified server is enough.
- Slack: The bot as to be added to the workspace and a token has to be generated.

Then please take a look at the provided example configuration in _config/config.toml_ and adapt it to match your settings.

To start AbyleBotter using the self-built or downloaded binary enter

```
./path/to/abylebotter -c /path/to/config/file.toml
```

The Bot should now connect automatically to the service and should be ready to use.

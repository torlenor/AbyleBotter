package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.abyle.org/redseligg/botorchestrator/botconfig"

	"github.com/gorilla/websocket"

	"github.com/torlenor/abylebotter/logging"
	"github.com/torlenor/abylebotter/platform"
	"github.com/torlenor/abylebotter/plugin"
	"github.com/torlenor/abylebotter/utils"
	"github.com/torlenor/abylebotter/webclient"
)

var (
	log = logging.Get("DiscordBot")
)

// Used for injection in unit tests
var newHeartbeatSender = func(ws webSocketClient) *discordHeartBeatSender {
	return &discordHeartBeatSender{ws: ws}
}

type webSocketClient interface {
	Dial(wsURL string) error
	Stop()

	Close() error

	ReadMessage() (int, []byte, error)

	SendMessage(messageType int, data []byte) error
	SendJSONMessage(v interface{}) error
}

type api interface {
	Call(path string, method string, body string) (webclient.APIResponse, error)
}

// The Bot struct holds parameters related to the bot
type Bot struct {
	api api

	gatewayURL string
	ws         webSocketClient

	knownChannels     map[string]channelCreate
	token             string
	ownSnowflakeID    string
	currentSeqNumber  int
	heartBeatStopChan chan bool
	seqNumberChan     chan int

	plugins []plugin.Hooks

	wg sync.WaitGroup

	watchdog *utils.Watchdog

	guilds        map[string]guildCreate // map[ID]
	guildNameToID map[string]string
}

// CreateDiscordBotWithAPI creates a new instance of a DiscordBot with the
// provided api
func CreateDiscordBotWithAPI(api api, cfg botconfig.DiscordConfig, ws webSocketClient) (*Bot, error) {
	log.Info("DiscordBot is CREATING itself")

	b := Bot{
		api: api,

		token: cfg.Token,
		ws:    ws,

		watchdog: &utils.Watchdog{},
	}

	url, err := b.getGateway()
	if err != nil {
		return nil, fmt.Errorf("Error connecting to Discord servers: %s", err)
	}
	b.gatewayURL = url

	b.knownChannels = make(map[string]channelCreate)

	b.seqNumberChan = make(chan int)

	b.guilds = make(map[string]guildCreate)
	b.guildNameToID = make(map[string]string)

	return &b, nil
}

// CreateDiscordBot creates a new instance of a DiscordBot
func CreateDiscordBot(cfg botconfig.DiscordConfig, ws webSocketClient) (*Bot, error) {
	api := webclient.New("https://discordapp.com/api", "Bot "+cfg.Token, "application/json")

	return CreateDiscordBotWithAPI(api, cfg, ws)
}

func (b *Bot) startHeartbeatSender(heartbeatInterval int) {
	b.heartBeatStopChan = make(chan bool)

	interval := time.Duration(heartbeatInterval) * time.Millisecond
	go func() {
		b.wg.Add(1)
		heartBeat(interval, newHeartbeatSender(b.ws), b.heartBeatStopChan, b.seqNumberChan, b.onFail)
		defer b.wg.Done()
	}()
	b.watchdog.SetFailCallback(b.onFail).Start(2 * interval)
}

func (b *Bot) reconnectWebSocket() error {
	err := b.ws.Close()
	if err != nil {
		log.Warnf("Error closing the WebSocket: %s", err)
	}
	return b.ws.Dial(b.gatewayURL)
}

func (b *Bot) run() {
	var fail bool
	for {
		_, message, err := b.ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Debugln("Connection closed normally: ", err)
				break
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway) {
				log.Infof("Received GoingAway from WebSocket, reconnecting")
				err := b.reconnectWebSocket()
				if err != nil {
					log.Errorf("Could not dial Discord WebSocket, Discord Bot not operational: %s", err.Error())
					break
				}
				continue
			} else {
				log.Errorln("UNHANDLED ERROR: ", err)
				fail = true
				break
			}
		}

		var data map[string]interface{}

		if err := json.Unmarshal(message, &data); err != nil {
			log.Errorln("UNHANDLED ERROR: ", err)
			continue
		}

		if data["op"].(float64) == 10 { // Hello from Discord Gateway
			log.Debugln("Received HELLO from gateway")
			sendIdent(b.token, b.ws)
			heartbeatInterval := int(data["d"].(map[string]interface{})["heartbeat_interval"].(float64))
			b.startHeartbeatSender(heartbeatInterval)
			log.Infoln("DiscordBot is READY")
		} else if data["op"].(float64) == 0 { // Dispatch to event handlers
			switch data["t"] {
			case "MESSAGE_CREATE":
				b.handleMessageCreate(data)
			case "READY":
				b.handleReady(data)
			case "GUILD_CREATE":
				b.handleGuildCreate(data)
			case "PRESENCE_UPDATE":
				b.handlePresenceUpdate(data)
			case "PRESENCE_REPLACE":
				b.handlePresenceReplace(data)
			case "TYPING_START":
				b.handleTypingStart(data)
			case "CHANNEL_CREATE":
				b.handleChannelCreate(data)
			case "MESSAGE_REACTION_ADD":
				b.handleMessageReactionAdd(data)
			case "MESSAGE_REACTION_REMOVE":
				b.handleMessageReactionRemove(data)
			case "MESSAGE_DELETE":
				b.handleMessageDelete(data)
			case "MESSAGE_UPDATE":
				b.handleMessageUpdate(data)
			case "CHANNEL_PINS_UPDATE":
				b.handleChannelPinsUpdate(data)
			case "GUILD_MEMBER_UPDATE":
				b.handleGuildMemberUpdate(data)
			case "PRESENCES_REPLACE":
				b.handlePresencesReplace(data)
			default:
				log.Errorln("Unhandled message:", string(message))
				b.handleUnknown(data)
			}
			b.currentSeqNumber = int(data["s"].(float64))
			b.seqNumberChan <- b.currentSeqNumber
		} else if data["op"].(float64) == 9 { // Invalid Session
			log.Errorln("Invalid Session received. Please try again...")
			return
		} else if data["op"].(float64) == 11 { // Heartbeat ACK
			b.watchdog.Feed()
		} else { // opcode which is not handled, yet
			log.Errorf("data: %s", data)
		}
	}

	if fail {
		b.onFail()
	}
}

// Start the Discord Bot
func (b *Bot) start() error {
	log.Infof("DiscordBot is STARTING (have %d plugin(s))", len(b.plugins))

	err := b.ws.Dial(b.gatewayURL)
	if err != nil {
		return fmt.Errorf("Could not dial Discord WebSocket, Discord Bot not operational: %s", err.Error())
	}

	go func() {
		b.wg.Add(1)
		b.run()
		defer b.wg.Done()
	}()

	for _, plugin := range b.plugins {
		plugin.OnRun()
	}
	log.Infoln("DiscordBot is RUNNING")

	return nil
}

// Run the Discord Bot (blocking)
func (b *Bot) Run(ctx context.Context) error {
	err := b.start()
	if err != nil {
		return err
	}

	<-ctx.Done()

	for _, plugin := range b.plugins {
		plugin.OnStop()
	}

	b.stop()

	return nil
}

func (b *Bot) stopHeartBeatWatchdog() {
	b.watchdog.Stop()
	b.heartBeatStopChan <- true
}

// Stop the Discord Bot
func (b *Bot) stop() {
	log.Infoln("DiscordBot is SHUTING DOWN")

	b.stopHeartBeatWatchdog()

	err := b.closeWebSocket()
	if err != nil {
		log.Errorln("Error when writing close message to ws:", err)
	}

	b.wg.Wait()
	b.ws.Stop()

	log.Infoln("DiscordBot is SHUT DOWN")
}

// AddPlugin takes as argument a plugin and
// adds it to the bot providing it with the API
func (b *Bot) AddPlugin(plugin platform.BotPlugin) {
	plugin.SetAPI(b)
	b.plugins = append(b.plugins, plugin)
}

// GetInfo returns information about the Bot
func (b *Bot) GetInfo() platform.BotInfo {
	return platform.BotInfo{
		BotID:    "",
		Platform: "Discord",
		Healthy:  true,
		Plugins:  []platform.PluginInfo{},
	}
}

func (b *Bot) closeWebSocket() error {
	return b.ws.SendMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func (b *Bot) onFail() {
	log.Warnf("Encountered an error, trying to restart the bot...")

	b.stopHeartBeatWatchdog()

	err := b.closeWebSocket()
	if err != nil {
		log.Errorf("Error when writing close message to ws: %s, still trying to recover", err)
	}

	b.wg.Wait()
	b.ws.Stop()

	url, err := b.getGateway()
	if err != nil {
		log.Errorf("Error connecting to Discord servers: %s", err)
		return
	}
	b.gatewayURL = url

	err = b.ws.Dial(b.gatewayURL)
	if err != nil {
		log.Errorln("Could not dial Discord WebSocket, Discord Bot not operational:", err)
		return
	}

	go func() {
		b.wg.Add(1)
		b.run()
		defer b.wg.Done()
	}()

	log.Info("Recovery attempt finished")
}

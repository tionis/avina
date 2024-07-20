package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "roll",
			Description: "Rolls a dice",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "amount",
					Description: "Amount of dice to roll, n to roll n CoD dice, 'ndx' to roll n x-sided dice",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "modifier",
					Description: "Modifier to add to the roll",
					Required:    false,
				},
			},
		},
		{
			Name:        "shadowroll",
			Description: "Rolls dice using Shadowrun rules",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "amount",
					Description: "Amount of dice to roll",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionBoolean,
					Name:        "exploding",
					Description: "Whether to explode 6s"},
			},
		},
	}
	commandsRegistered = map[string]bool{}
	commandHandlers    = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"roll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			amount := options[0].StringValue()
			var modifier string
			if len(options) > 1 {
				modifier = options[1].StringValue()
			} else {
				modifier = ""
			}
			response, err := roll(amount, modifier)
			if err != nil {
				log.Printf("Error rolling dice: %v", err)
				response = err.Error()
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
			if err != nil {
				log.Println(err)
			}
		},
		"shadowroll": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			amountStr := i.ApplicationCommandData().Options[0].StringValue()
			exploding := i.ApplicationCommandData().Options[1].BoolValue()
			amount, err := strconv.Atoi(amountStr)
			if err != nil {
				log.Println(err)
				return
			}
			response := shadowroll(amount, exploding)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
				},
			})
			if err != nil {
				log.Println(err)
			}
		},
	}
)

type Bot struct {
	token string
	log   *slog.Logger
	AppID string
}

func NewBot(logger *slog.Logger, token string, appID string) *Bot {
	return &Bot{
		token: token,
		log:   logger,
		AppID: appID,
	}
}

func (b *Bot) Start() error {
	dg, err := discordgo.New("Bot " + b.token)
	if err != nil {
		return err
	}
	dg.AddHandler(b.messageCreate)
	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
				// TODO convert types etc here for backwards compatibility?
				h(s, i)
			}
		}
	})
	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages
	dg.Identify.Intents |= discordgo.IntentsDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		return err
	}

	// Wait here until CTRL-C or other term signal is received.
	b.log.Info("Avina bot was started")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	return dg.Close()
}

func (b *Bot) messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if !commandsRegistered[m.GuildID] {
		b.log.Info("Registering commands", "guild", m.GuildID)
		_, err := s.ApplicationCommandBulkOverwrite(b.AppID, m.GuildID, commands)
		if err != nil {
			b.log.Error("Error creating application command", "error", err)
			return
		}
		commandsRegistered[m.GuildID] = true
		b.log.Info("Commands registered", "guild", m.GuildID)
	}
	// get first word
	words := strings.Split(m.Content, " ")
	if len(words) > 0 {
		switch words[0] {
		case "roll", "r":
			if len(words) > 1 {
				var amount, modifier string
				amount = words[1]
				if len(words) > 2 {
					modifier = words[2]
				}
				response, err := roll(amount, modifier)
				if err != nil {
					log.Printf("Error rolling dice: %v", err)
					response = err.Error()
				}
				response = m.Author.Mention() + "\n" + response
				send, err := s.ChannelMessageSend(m.ChannelID, response)
				if err != nil {
					b.log.Error("Error sending message", "error", err, "message", send)
				}
			} else {
				send, err := s.ChannelMessageSend(m.ChannelID, "Please specify amount and optional modifier")
				if err != nil {
					b.log.Error("Error sending message", "error", err, "message", send)
				}
			}
		case "shadowroll", "sr":
			if len(words) > 1 {
				var amountStr string
				amountStr = words[1]
				exploding := false
				if len(words) > 2 {
					switch words[2] {
					case "e", "exploding", "edge", "!", "!!":
						exploding = true
					default:
						send, err := s.ChannelMessageSend(m.ChannelID, "Please specify amount and optional exploding flag")
						if err != nil {
							b.log.Error("Error sending message", "error", err, "message", send)
						}
						return
					}
				}
				amount, err := strconv.Atoi(amountStr)
				if err != nil {
					send, err := s.ChannelMessageSend(m.ChannelID, "Please specify amount as a number")
					if err != nil {
						b.log.Error("Error sending message", "error", err, "message", send)
					}
					return
				}
				response := shadowroll(amount, exploding)
				response = m.Author.Mention() + "\n" + response
				send, err := s.ChannelMessageSend(m.ChannelID, response)
				if err != nil {
					b.log.Error("Error sending message", "error", err, "message", send)
				}
			} else {
				send, err := s.ChannelMessageSend(m.ChannelID, "Please specify amount")
				if err != nil {
					b.log.Error("Error sending message", "error", err, "message", send)
				}
			}
		}
	}
}

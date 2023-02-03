/*
 * File: bot.go
 * Project: slack
 * File Created: Tuesday, 24th January 2023 5:25:17 pm
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Monday, 30th January 2023 8:55:08 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package slackbot

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/patrickmn/go-cache"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"github.com/markmester/fanyi-slackbot/pkg/clients"
	"github.com/markmester/fanyi-slackbot/pkg/common"
)

const (
	// Cache expiration time of 5 minutes
	cacheExpireDuration = 5 * time.Minute
	// Cache purges expired items every 10 minutes
	cacheCleanupDuration = 10 * time.Minute
	// Datastore key
	datastoreKey = "store.json"
)

var (
	ErrMsgInternalServerError = "An unexpected error occurred; please try again later!"
	ErrMsgUnsupportedFlag     = "Sorry, the flag '%s' is not supported!"
	ErrMsgUnknownLanguage     = "Sorry, we are unable to detect the language of provided text"

	//go:embed templates/*
	templates embed.FS
)

// Bot provides a control plane to both slack and gpts,
// responding to messages in channels and providing translations
type Bot struct {
	slack     *clients.SlackClient
	gpt       *clients.Gpt3Client
	detector  *clients.Detector
	datastore clients.DataStore

	cache *cache.Cache

	logger *zap.SugaredLogger
}

// New creates a new bot, and subscribes to slack events for Process
// to start processing
func New(slackClient *clients.SlackClient, gpt3Client *clients.Gpt3Client, detector *clients.Detector, datastore clients.DataStore) (*Bot, error) {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// Load our config if it exists
	if config, err := datastore.Get(datastoreKey); err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "error retrieving config from datastore")
		}
	} else {
		if err := detector.FromJSON(config); err != nil {
			return nil, errors.Wrapf(err, "error loading config into detector")
		}
	}

	bot := Bot{
		slack:     slackClient,
		gpt:       gpt3Client,
		cache:     cache.New(cacheExpireDuration, cacheCleanupDuration),
		datastore: datastore,
		logger:    logger.Sugar(),
		detector:  detector,
	}

	return &bot, nil
}

func (b *Bot) Shutdown() {
	b.logger.Info("Bot shutting down; cleaning up")
	jsonBytes, err := b.detector.ToJSON()
	if err != nil {
		b.logger.Error("error persisting configuration to datastore!")
		return
	}

	if err := b.datastore.Set(datastoreKey, jsonBytes); err != nil {
		b.logger.Error("error persisting configuration to datastore!")
		return
	}
}

// Process will:
//  1. Listen to slack events
//  2. On channel message, detect character encoding
//  3. Perform appropriate translation
//  4. Respond as thread reply
func (b *Bot) Process() error {
	b.logger.Info("Starting bot receive routine...")
	for evt := range b.slack.Sock().Events {
		switch evt.Type {
		case socketmode.EventTypeEventsAPI:
			eventsAPIEvent, _ := evt.Data.(slackevents.EventsAPIEvent)
			switch eventsAPIEvent.Type {
			case slackevents.CallbackEvent:
				innerEvent := eventsAPIEvent.InnerEvent

				switch ev := innerEvent.Data.(type) {
				case *slackevents.ReactionAddedEvent:
					var timestamp = ev.Item.Timestamp

					err := func() error {
						if user, found := b.cache.Get(ev.Item.Timestamp); found {
							if user.(string) == ev.User {
								b.logger.Infof("received duplicate message; skipping")
								return nil
							}
						}
						b.cache.Set(ev.Item.Timestamp, ev.User, cache.DefaultExpiration)

						// Get associated slack message
						msg, err := b.slack.GetMessage(ev.Item.Channel, ev.Item.Timestamp)
						if err != nil {
							b.logger.Errorf("unable to get associated msg for reaction; err=%s", err.Error())
							return fmt.Errorf(ErrMsgInternalServerError)
						}
						timestamp = msg.Timestamp // update to thread timestamp

						// Map emoji to language
						targetLanguage, ok := GetLanguageCode(ev.Reaction)
						if !ok {
							if strings.HasPrefix("flag-", ev.Reaction) {
								b.logger.Errorf("unable to get language corresponding to emoji reaction: %s", ev.Reaction)
								return fmt.Errorf(ErrMsgUnsupportedFlag, ev.Reaction)
							}
							// Not a flag emoji
							return nil
						}

						// Detect language
						sourceLanguage, exists := b.detector.Detect(msg.Text)
						if !exists {
							b.logger.Errorf("unable to determine language of message")
							return fmt.Errorf(ErrMsgUnknownLanguage)
						}

						// Translate
						body, err := b.gpt.Translate(sourceLanguage, "", targetLanguage, "", msg.Text)
						if err != nil {
							b.logger.Errorf("unable to provide translation for msg=%s from %s->%s; err=%s", msg.Text, sourceLanguage, targetLanguage, err.Error())
							return fmt.Errorf(ErrMsgInternalServerError)
						}

						if err := b.slack.PostMessage(ev.Item.Channel,
							slack.MsgOptionText(body, false),
							slack.MsgOptionTS(msg.Timestamp),
						); err != nil {
							b.logger.Errorf("unable to post translation for msg=%s from %s->%s; err=%s", msg.Text, sourceLanguage, targetLanguage, err.Error())
							return fmt.Errorf(ErrMsgInternalServerError)
						}

						return nil
					}()

					if err != nil {
						if err := b.slack.PostEphemeralMessage(
							ev.Item.Channel,
							ev.User,
							slack.MsgOptionText("Sorry! Something went wrong. Please try again later!", false),
							slack.MsgOptionTS(timestamp)); err != nil {
							b.logger.Errorf("unable to post message; err=%s", err.Error())
						}
					}

				case *slackevents.MessageEvent:
					timestamp := ev.TimeStamp
					if ev.ThreadTimeStamp != "" {
						timestamp = ev.ThreadTimeStamp
					}

					err := func() error {
						// Event is a bot event or does not contain any text
						if ev.BotID != "" || ev.Text == "" {
							return nil
						}

						// Retrieve select detector for this channel
						selectDetector, err := b.detector.GetSelectedDetector(ev.Channel)
						if err != nil {
							// Auto translation hasn't been configured
							return nil
						}

						b.logger.Infof("Retrieved select detector for channel=%s: %s <-> %s", ev.Channel, selectDetector.Selected.L1.String(), selectDetector.Selected.L2.String())

						// Ignore Cache hits
						if user, found := b.cache.Get(ev.TimeStamp); found {
							if user.(string) == ev.User {
								b.logger.Infof("received duplicate message; skipping")
								return nil
							}
						}
						b.cache.Set(ev.TimeStamp, ev.User, cache.DefaultExpiration)

						// Detect language
						sourceLanguage, err := selectDetector.Select(ev.Channel, ev.Text)
						if err != nil {
							b.logger.Errorf("error detecting language in channel=%s; err=%s", ev.Channel, err.Error())
							return nil
						}

						var targetLanguage string
						switch {
						case sourceLanguage == selectDetector.Selected.L1:
							targetLanguage = selectDetector.Selected.L2.String()
						case sourceLanguage == selectDetector.Selected.L2:
							targetLanguage = selectDetector.Selected.L1.String()
						default:
							b.logger.Infof("source language not in configured auto-translation pair; skipping")
							return nil
						}

						b.logger.Infof("Translating the following between: %s<->%s: %s", sourceLanguage.String(), targetLanguage, ev.Text)

						// Translate
						body, err := b.gpt.Translate(sourceLanguage.String(), "", targetLanguage, "", ev.Text)
						if err != nil {
							b.logger.Errorf("unable to provide translation for msg=%s from %s->%s; err=%s", ev.Text, sourceLanguage.String(), targetLanguage, err.Error())
							return fmt.Errorf(ErrMsgInternalServerError)
						}

						// Reply in thread
						if err := b.slack.PostMessage(ev.Channel,
							slack.MsgOptionText(body, false),
							slack.MsgOptionTS(timestamp),
						); err != nil {
							b.logger.Errorf("unable to post translation for msg=%s; err=%s", ev.Text, err.Error())
							return fmt.Errorf(ErrMsgInternalServerError)
						}

						return nil
					}()

					if err != nil {
						if err := b.slack.PostEphemeralMessage(
							ev.Channel,
							ev.User,
							slack.MsgOptionText("Sorry! Something went wrong. Please try again later!", false),
							slack.MsgOptionTS(timestamp)); err != nil {
							b.logger.Errorf("unable to post message; err=%s", err.Error())
						}
					}
				}
			}

			b.slack.Ack(*evt.Request)

		case socketmode.EventTypeSlashCommand:
			// Just like before, type cast to the correct event type, this time a SlashEvent
			command, ok := evt.Data.(slack.SlashCommand)
			if !ok {
				b.logger.Infof("Could not type cast the message to a SlashCommand: %v", command)
				continue
			}

			// Acknowledge the request
			b.slack.Ack(*evt.Request)

			// handleSlashCommand will take care of the command
			if err := b.handleSlashCommand(command); err != nil {
				b.logger.Infof("Could not process slash command: err=%s", err.Error())
				continue
			}

		case socketmode.EventTypeInteractive:
			interaction, ok := evt.Data.(slack.InteractionCallback)
			if !ok {
				b.logger.Errorf("Could not type cast the message to a Interaction callback: err=%v", interaction)
				continue
			}

			err := b.handleInteractionEvent(interaction)
			if err != nil {
				b.logger.Infof("Could not process slash command: err=%s", err.Error())
				continue
			}
			b.slack.Ack(*evt.Request)

		} //end of switch
	}

	return nil
}

// handleSlashCommand will take a slash command and route to the appropriate function
func (b *Bot) handleSlashCommand(command slack.SlashCommand) error {
	// We need to switch depending on the command
	switch command.Command {
	case "/help":
		return b.handleHelpCommand(command)
	case "/translate":
		return b.handleTranslateCommand(command)
	}

	return nil
}

func (b *Bot) handleInteractionEvent(interaction slack.InteractionCallback) error {
	// Handle interaction
	switch interaction.Type {
	case slack.InteractionTypeBlockActions:
		log.Printf("Received interaction event on channel: %s (%s)", interaction.Channel.Name, interaction.Channel.ID)
		// This is a block action, so we need to handle it
		for _, action := range interaction.ActionCallback.BlockActions {
			if action.ActionID == "multi_static_select_action-language-select" {
				selectedOptions := []string{}
				for _, opt := range action.SelectedOptions {
					selectedOptions = append(selectedOptions, opt.Text.Text)
				}
				if len(selectedOptions) == 2 {
					if ok, err := b.detector.UpdateSelected(interaction.Channel.ID, selectedOptions[0], selectedOptions[1]); err != nil {
						return err
					} else if ok {
						b.logger.Infof("updated auto-translation selection to %s:%s", selectedOptions[0], selectedOptions[1])

						if err := b.slack.PostMessage(
							interaction.Channel.ID,
							slack.MsgOptionText(fmt.Sprintf("Auto-translation activated: %s  ↔  %s ", selectedOptions[0], selectedOptions[1]), false),
							slack.MsgOptionTS(interaction.Message.Timestamp)); err != nil {
							return fmt.Errorf("failed to post message: %s", err.Error())
						}
					}
				} else {
					if err := b.slack.PostMessage(
						interaction.Channel.ID,
						slack.MsgOptionText("Please select 2 languages.", false),
						slack.MsgOptionTS(interaction.Message.Timestamp)); err != nil {
						return fmt.Errorf("failed to post message: %s", err.Error())
					}
				}
			}
		}
	default:
		// NooP
	}

	return nil
}

// handleTranslateCommand will trigger a prompt to select between a common list of translation languages
func (b *Bot) handleTranslateCommand(command slack.SlashCommand) error {
	if strings.Contains(command.Text, "stop") {
		b.logger.Info("stopping auto-translation")
		b.detector.ClearSelected(command.ChannelID)
		if err := b.slack.PostMessage(command.ChannelID,
			slack.MsgOptionText("Stopping auto-translation!", false)); err != nil {
			return err
		}
		return nil
	}

	data, err := templates.ReadFile("templates/language_select.json")
	if err != nil {
		return err
	}

	var blocks common.Blocks
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		return err
	}

	var options []slack.MsgOption
	options = append(options, slack.MsgOptionBlocks(blocks.Blocks...))
	if err := b.slack.PostMessage(command.ChannelID, options...); err != nil {
		return err
	}

	return nil
}

// handleHelpCommand will provide a help dialog
func (b *Bot) handleHelpCommand(command slack.SlashCommand) error {
	// The Input is found in the text field so
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{
		Color: "#4af030",
		Title: "Fanyi Translator Help",
		Text: `
Welcome to Fanyi, your personal Translator!
Please see below for a list of available commands:

• /help → Display this help message.

• /translate → Automatically detect and translate between the specified languages.

• /translate stop → Stop auto-translation.

• flag emoji → React to any message with a flag emoji (🇺🇸) and Fanyi will respond with the translation of that flags language.
`,
	}

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
	if err := b.slack.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment)); err != nil {
		return fmt.Errorf("failed to post message: %s", err.Error())
	}
	return nil
}

/*
 * File: slack.go
 * Project: bot
 * File Created: Wednesday, 25th January 2023 2:56:57 pm
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Wednesday, 25th January 2023 8:15:51 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package clients

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SlackClient struct {
	client *slack.Client
	socket *socketmode.Client
}

type slackMsg struct {
	Text      string
	Timestamp string
}

func NewSlackClient(slackBotToken, slackAppToken string) *SlackClient {
	client := slack.New(slackBotToken, slack.OptionAppLevelToken(slackAppToken))
	socket := socketmode.New(client)

	// TODO: graceful shutdown of this routine
	go func() {
		log.Println("Starting slack runloop")
		for {
			if err := socket.Run(); err != nil {
				log.Println("error connecting to slack; retrying in 10 seconds...")
				time.Sleep(10 * time.Second)
			}
		}
	}()

	return &SlackClient{
		client: client,
		socket: socket,
	}
}

func (s *SlackClient) Sock() *socketmode.Client {
	return s.socket
}

func (s *SlackClient) PostMessage(channelId string, options ...slack.MsgOption) error {
	if _, _, err := s.socket.PostMessage(channelId, options...); err != nil {
		return err
	}
	return nil
}

func (s *SlackClient) PostEphemeralMessage(channelId, userId string, options ...slack.MsgOption) error {
	if _, err := s.socket.PostEphemeral(channelId, userId, options...); err != nil {
		return err
	}
	return nil
}

func (s *SlackClient) Ack(ack socketmode.Request, payload ...interface{}) {
	s.socket.Ack(ack, payload...)
}

// https://api.slack.com/methods/conversations.replies
func (s *SlackClient) GetMessage(id string, timestamp string) (*slackMsg, error) {
	params := &slack.GetConversationRepliesParameters{}
	params.ChannelID = id
	params.Timestamp = timestamp
	params.Inclusive = true
	params.Limit = 1

	// get slack messages
	msg, _, _, err := s.client.GetConversationReplies(params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get slack messages")
	}

	// get message text
	slMsg := &slackMsg{}
	for _, i := range msg {
		slMsg.Timestamp = i.Timestamp
		if slMsg.Timestamp == "" {
			slMsg.Timestamp = i.ThreadTimestamp
		}

		slMsg.Text = i.Text
		if slMsg.Text == "" {
			for _, j := range i.Attachments {
				slMsg.Text = j.Text
				break
			}
		}
	}

	return slMsg, nil
}

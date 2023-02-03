/*
 * File: gpt3.go
 * Project: clients
 * File Created: Wednesday, 25th January 2023 3:02:02 pm
 * Author: Mark Mester (mmester6016@gmail.com)
 * -----
 * Last Modified: Thursday, 2nd February 2023 10:19:42 pm
 * Modified By: Mark Mester (mmester6016@gmail.com>)
 */
package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/PullRequestInc/go-gpt3"
	"github.com/pkg/errors"
)

type Gpt3Client struct {
	gpt3.Client
}

func NewGpt3Client(chatGptApiKey, chatGptEngine string) *Gpt3Client {
	return &Gpt3Client{
		gpt3.NewClient(chatGptApiKey, gpt3.WithDefaultEngine(chatGptEngine)),
	}
}

func (g *Gpt3Client) Translate(fromLanguage, fromDialect, toLanguage, toDialect, msg string) (string, error) {
	ask := ""

	switch {
	case fromDialect != "" && toDialect != "":
		ask = fmt.Sprintf("Translate this from %s (%s) to %s (%s): %s", fromLanguage, fromDialect, toLanguage, toDialect, msg)
	case fromDialect != "" && toDialect == "":
		ask = fmt.Sprintf("Translate this from %s (%s) to %s: %s", fromLanguage, fromDialect, toLanguage, msg)
	case fromDialect == "" && toDialect != "":
		ask = fmt.Sprintf("Translate this from %s to %s(%s): %s", fromLanguage, toLanguage, toDialect, msg)
	case fromDialect == "" && toDialect == "":
		ask = fmt.Sprintf("Translate this from %s to %s: %s", fromLanguage, toLanguage, msg)
	default:
		return "", fmt.Errorf("from and to language must be defined")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancel()

	resp, err := g.Completion(ctx, gpt3.CompletionRequest{
		Prompt:           []string{ask},
		MaxTokens:        gpt3.IntPtr(512),
		Temperature:      gpt3.Float32Ptr(0.3),
		TopP:             gpt3.Float32Ptr(1),
		FrequencyPenalty: 0,
		PresencePenalty:  0,
		Echo:             false,
	})
	if err != nil {
		return "", errors.Wrap(err, "ChatGPT completion API error")
	}

	return resp.Choices[0].Text, nil
}

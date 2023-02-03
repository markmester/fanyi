# Fanyi

Fanyi is a translation slack bot that utilizes Chat-GPT's `davinci-003` model for real-time translation.

## Motivation

Living in a bi-lingual household and not able to find any suitable auto-translation option, Fanyi was developed to be able to easily and collaboratively chat with each other. Fanyi utilizes Chat-GPT's `davinci-003` for translation, which provides context and dialect aware translations. This translation engine seems to provide better translation than other current options (based solely on empirical data).

## Current Feature-Set

- Provides auto-translation for a channel between 2 languages.
- Respond to message with a flag emoji(e.g. 🇨🇳 🇺🇸 🇬🇧) and Fanyi will translate to the language of that flag's country.
- Saves user configuration (e.g. channels configured for auto-translation) to local storage.

## Future Feature-Set

- Feature to allow users to setup a channel that will auto translate from the source channel and mirror the conversation into the new, language specific channel.

- Enable dialect selection. This was tested in an earlier version of this app, and it seemed to work great (for our use case - Wuhan, Chinese dialect), but I did not have time to integrate it into this version.

- Feedback loop for mistranslated messages. ChatGPT is smart enough to take feedback and re-craft a response (possibly with a rewording of the text). A user could respond, indicating the translation doesn't make sense and it could try agin with a new translation and/or prompt the OP to reword the message.

- Incorporate chat history to ChatGPT engine. This would allow it to learn the message style and help translate better.

- S3 config store

## Configuration

Rename the `.env.sample` to `.env` and fill the env vars with your credentials. See [Set Up Your Slack App](#set-up-your-slack-app) and  [Get an OpenAPI API Key](#get-an-openapi-api-key) for details on how to obtain env vars info.

> Note: The environmental variable prefixed with `AWS` are only relevant for deployment of the app to AWS. For local deployments, these can be ignored.

### Set Up Your Slack App

1. Create an app at your Slack App Settings page at [api.slack.com/apps](https://api.slack.com/apps)
2. Choose "From an app manifest", select the workspace you want to use, then paste the contents of [`manifest.yml`](./manifest.yml) into the dialog marked "Enter app manifest below".
3. On the **OAuth & Permissions** page, install the app and get a **Bot User OAuth Token** - it begins with `xoxb-`. Copy this new token to your `.env` file as `SLACK_BOT_TOKEN`
4. On the **Basic Information** page, scroll down to **App-Level Tokens** and click **Generate Token and Scopes**. Add the following scopes `connections:write` scope, give your token a name, and click **Generate**. Copy this new token to your `.env` file as `SLACK_APP_TOKEN`

### Get an OpenAPI API Key

Register for a paid OpenAPI account and obtain a [ChatGPT API key](https://beta.openai.com/account/api-keys)

## Run

### Local

After setting up the [App Configuration](#configuration), run:

```sh
    ENV=.env make build
   ENV=.env make up
```

> By default, this will start a bot in 'ephemeral' mode. This means any user configurations (e.g. channel auto-translation preferences) will not be saved when the bot restarts. The configuration can be persisted by following the instructions in the [.env.example](.env.example)

### ECS

The app can be deployed to ECS using the docker-compose ECS context. To deploy, configure a `.env` file according to the `.env.example` file and run: `ENV=.env make deploy".

> Note: ECS deployment not yet fully tested/documented

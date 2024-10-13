# GPT Helper

## Features

- Chat with AI using text or voice messages
- Keeps conversation context
- Simple command system

## Usage

Chat with the bot on Telegram and use these commands:
- `/start`: Begin chatting
- `/prompt`: Enter a custom prompt
- `/clear`: Reset conversation
- `/commands`: See available commands

## Build

You'll need:
- Go 1.22+
- OpenAI API key
- Telegram Bot Token

Steps:
1. Clone the repo
2. Run `go mod download`
3. Set your API keys as env variables
4. Build with `go build -o gpt-helper .`
5. Run it: `./gpt-helper`

## Docker Setup

1. Build: `docker build -t gpt-helper .`
2. Run:
   ```
   docker run -d --name gpt-helper \
     -e OPENAI_API_KEY=your_key \
     -e TELEGRAM_BOT_TOKEN=your_token \
     gpt-helper
   ```

That's it! The bot should be up and running.

## Settings

You can adjust these in `telegram/bot.go`:
- `maxSessionCtxLength`: Max messages in context
- `requestTimeout`: API request timeout

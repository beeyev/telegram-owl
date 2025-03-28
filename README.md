# Telegram Owl 🦉

> 📣 Use it to send deployment notifications, alerts, logs, reports, or media — from scripts, cron jobs, CI/CD pipelines, monitoring tools, or any shell environment.

Whether you're a DevOps engineer automating infrastructure, a developer managing CI/CD pipelines, or just want to notify your Telegram group from a terminal script — **Telegram Owl** gives you a simple and script-friendly way to do it.

---

## ✨ Features

- 📨 Send text messages
- 📎 Attach multiple files
- 🔕 Silent messages (no notification sound)
- 🛡️ Protect messages (disable forwarding/saving)
- 📸 Automatic media type detection (or force as document)
- 🧵 Send to forum thread topics
- 📤 Read input from `stdin`
- 📌 Set environment variables for easy usage
- 🐧 Cross-platform support (Windows, Mac, Linux)
- 🚀 Fast and lightweight (written in Go)

## 📦 Installation

### Binary Releases
For Windows, Mac OS(10.12+) or Linux, you can download a binary release [here](https://github.com/beeyev/telegram-owl/releases/latest).

## 🚀 Usage

To start using **Telegram Owl**, you need to obtain a Telegram bot token and chat ID.
You can learn how to get it [here](/docs/HowToTelegramBot.md).

```bash
telegram-owl \
  --token <bot-token> \
  --chat <chat-id or @channel> \
  [--message "your message"] \
  [--attach file1,file2,...] \
  [options]
```

### 🔐 Required Flags

| Flag            | Description                     | Environment Variable        |
|----------------|---------------------------------|-----------------------------|
| `--token`, `-t`  | Telegram bot token             | `TELEGRAM_SEND_TOKEN`       |
| `--chat`, `-c`   | Chat ID or `@username`        | `TELEGRAM_SEND_CHAT`        |

### ⚙️ Common Flags

| Flag                  | Description                                                    |
|-----------------------|----------------------------------------------------------------|
| `--message`, `-m`      | Text message to send                                          |
| `--format`, `-f`         | Message format options, possible values: `markdown`, `html` |
| `--stdin`              | Read message content from `stdin`                             |
| `--attach`, `-a`       | Attach files (comma-separated or multiple flags)              |
| `--as-document`, `-d`  | Force all files to be sent as documents                       |
| `--silent`, `-s`       | Send silently (no notification sound)                         |
| `--spoiler`            | Hide media with spoiler animation                             |
| `--protect`            | Prevent forwarding and saving of content                      |
| `--no-link-preview`    | Disable automatic link previews in messages                   |
| `--thread`             | Thread ID for forum supergroup topics                         |

## 📌 Examples

### ✅ Send a Simple Message

```bash
telegram-owl -t $BOT_TOKEN -c @mychannel -m "Server status: OK ✅"
```

### 📝 Send a Message with Markdown formatting
```bash
telegram-owl -t $BOT_TOKEN -c 123456 --format=markdown -m "*Bold text* via Markdown"
```

### 📝 Send a Message with HTML formatting
```bash
telegram-owl -t $BOT_TOKEN -c 123456 --format=html -m '<b>Bold text</b> via HTML and <a href="http://www.example.com/">inline URL</a>'
```

> Message formatting is supported for both `markdown` and `html` formats. But it does not work when text and files are sent together.

### 📎 Send Files with a Message

```bash
telegram-owl -t $BOT_TOKEN -c 123456 \
  -m "Daily report attached" \
  -a report.pdf,screenshot.png
```

### 🔕 Send a Protected, Silent Message

```bash
telegram-owl -t $BOT_TOKEN -c 123456 \
  -m "Confidential: Project roadmap" \
  --silent --protect
```

### 📤 Pipe Message from File or Command

```bash
cat message.txt | telegram-owl -t $BOT_TOKEN -c @devs --stdin
```

### 🧵 Post in a Forum Thread

```bash
telegram-owl -t $BOT_TOKEN -c @forumgroup --thread 67890 -m "New bug report 🐞"
```

## ⚙️ Configuration

Set environment variables to simplify usage:

```bash
export TELEGRAM_OWL_TOKEN="123:abc"
export TELEGRAM_OWL_CHAT="112451"
export TELEGRAM_OWL_THREAD="67890"
```

## 📏 Attachment Limits

| Limit Type              | Value         |
|-------------------------|---------------|
| Max attachments         | 10 files      |
| Max photo size          | 10 MB         |
| Max file size           | 50 MB         |
| Max total size per send | 50 MB total   |

## 🐞 Found a Bug or Want a Feature?

Feel free to open an issue on [GitHub](https://github.com/beeyev/telegram-owl/issues).

## © License

The MIT License (MIT). Please see [License File](https://github.com/beeyev/telegram-owl/blob/master/LICENSE) for more information.

---

If you like this project, please consider giving me a ⭐

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
- 🌐 Configure HTTP or SOCKS5 proxy
- 🐧 Cross-platform support (Windows, Mac, Linux)
- 🚀 Fast and lightweight (written in Go)

## 📦 Installation

### macOS and Linux

Install with [`Homebrew`](https://brew.sh/)
```console
brew tap beeyev/pkg https://github.com/beeyev/pkg
brew install --cask telegram-owl
```

**⚠️ macOS Security Note:**

If macOS blocks the app with "cannot be opened because the developer cannot be verified":

```bash
# Option 1: Remove quarantine attribute
xattr -d com.apple.quarantine /opt/homebrew/bin/telegram-owl

# Option 2: Allow in System Settings
# Go to: System Settings > Privacy & Security > Allow "telegram-owl"
```

### Windows

Install with [`Scoop`](https://scoop.sh/)
```console
scoop bucket add beeyev https://github.com/beeyev/pkg
scoop install telegram-owl
```

Install with `winget`
```console
winget install telegram-owl
```

### Binary Releases
For Windows, Mac OS(10.12+) or Linux, you can download a binary release [here](https://github.com/beeyev/telegram-owl/releases/latest).

### Docker
Official multi-architecture images live on [Docker Hub](https://hub.docker.com/r/beeyev/telegram-owl), [GHCR](https://github.com/beeyev/telegram-owl/pkgs/container/telegram-owl), and [Quay](https://quay.io/repository/beeyev/telegram-owl). Pull the tag you need and run the CLI directly:

```console
docker run --rm ghcr.io/beeyev/telegram-owl:latest --help
```

Need the binary inside your own image? Copy it from the published image via a multistage Dockerfile:

```Dockerfile
COPY --from=beeyev/telegram-owl:latest /usr/bin/telegram-owl /usr/local/bin/telegram-owl
```

This reuses the official build without compiling from source.

## 🚀 Usage

To start using **Telegram Owl**, you need to obtain a Telegram bot token and chat ID.
You can learn how to get it [here](/docs/HowToTelegramBot.md).

```console
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
| `--token`, `-t`  | Telegram bot token             | `TELEGRAM_OWL_TOKEN`       |
| `--chat`, `-c`   | Chat ID or `@username`        | `TELEGRAM_OWL_CHAT`        |

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
| `--proxy`              | Proxy URL (HTTP/HTTPS/SOCKS5) for outbound requests           |

## 📌 Examples

### ✅ Send a Simple Message

```console
telegram-owl -t $BOT_TOKEN -c @mychannel -m "Server status: OK ✅"
```

### 📝 Send a Message with Markdown formatting
```console
telegram-owl -t $BOT_TOKEN -c 123456 --format=markdown -m "*Bold text* via Markdown"
```

### 📝 Send a Message with HTML formatting
```console
telegram-owl -t $BOT_TOKEN -c 123456 --format=html -m '<b>Bold text</b> via HTML and <a href="http://www.example.com/">inline URL</a>'
```

> Message formatting is supported for both `markdown` and `html` formats. But it does not work when text and files are sent together.

### 📎 Send Files with a Message

```console
telegram-owl -t $BOT_TOKEN -c 123456 \
  -m "Daily report attached" \
  -a report.pdf,screenshot.png
```

### 🔕 Send a Protected, Silent Message

```console
telegram-owl -t $BOT_TOKEN -c 123456 \
  -m "Confidential: Project roadmap" \
  --silent --protect
```

### 📤 Pipe Message from File or Command

```console
cat message.txt | telegram-owl -t $BOT_TOKEN -c @devs --stdin
```

### 🧵 Post in a Forum Thread

```console
telegram-owl -t $BOT_TOKEN -c @forumgroup --thread 67890 -m "New bug report 🐞"
```

## ⚙️ Configuration

Set environment variables to simplify usage:

```console
export TELEGRAM_OWL_TOKEN="123:abc"
export TELEGRAM_OWL_CHAT="112451"
export TELEGRAM_OWL_THREAD="67890"
export TELEGRAM_OWL_PROXY="http://proxy.example.com:8080"
```

### 🌐 Proxy Configuration

> `telegram-owl` can route requests through a proxy. Supply the proxy via `--proxy` or the `TELEGRAM_OWL_PROXY` environment variable. Proxy handling is powered by [Resty](https://github.com/go-resty/resty) under the hood, so any scheme supported by Resty (`http`, `https`, `socks5`) works here.

- HTTP(S) proxy:
```console
telegram-owl --proxy http://proxy.local:3128 -t $BOT_TOKEN -c @channel -m "Hello via proxy"`
```

- SOCKS5 proxy:
```console
telegram-owl --proxy socks5://127.0.0.1:1080 -t $BOT_TOKEN -c @channel -m "Hello via SOCKS5"
```

- Authenticated proxy:
```console
telegram-owl --proxy http://user:pass@proxy.local:8080 -t $BOT_TOKEN -c @channel -m "Hello with auth proxy"
```

Authentication is supported by embedding credentials in the URL, e.g. `http://user:pass@proxy.local:3128`.

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

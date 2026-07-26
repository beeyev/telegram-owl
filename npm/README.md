# Telegram Owl

This npm package installs the native
[Telegram Owl](https://github.com/beeyev/telegram-owl) CLI for the current
operating system and CPU architecture.

```console
npm install --global telegram-owl
telegram-owl --help
```

Run without a global install:

```console
npx telegram-owl --help
```

The package does not use lifecycle scripts or download code at runtime. npm
selects a platform-specific optional package containing the native executable.

Supported targets:

- macOS: x64, arm64
- Linux: x64, arm64
- Windows: x64

For other targets, install from the
[GitHub Releases](https://github.com/beeyev/telegram-owl/releases) page.

See the main
[Telegram Owl documentation](https://github.com/beeyev/telegram-owl#readme)
for configuration and usage.

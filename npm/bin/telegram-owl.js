#!/usr/bin/env node
"use strict";

const path = require("node:path");
const { spawn } = require("node:child_process");
const { findPlatform } = require("../platforms");

const forwardedSignals = Object.freeze(["SIGINT", "SIGTERM", "SIGHUP"]);
const releasesURL = "https://github.com/beeyev/telegram-owl/releases";

function resolveBinary({
  platform = process.platform,
  arch = process.arch,
  resolveImplementation = require.resolve,
} = {}) {
  const definition = findPlatform(platform, arch);
  if (!definition) {
    throw new Error(
      `no npm binary is available for ${platform}-${arch}. ` +
        `Install from ${releasesURL}`,
    );
  }

  let manifestPath;
  try {
    manifestPath = resolveImplementation(
      `${definition.packageName}/package.json`,
      { paths: [path.resolve(__dirname, "..")] },
    );
  } catch (error) {
    if (error?.code !== "MODULE_NOT_FOUND") {
      throw error;
    }

    throw new Error(
      `${definition.packageName} is missing. ` +
        "Reinstall without --omit=optional",
      { cause: error },
    );
  }

  return path.join(path.dirname(manifestPath), "bin", definition.binaryName);
}

function forwardSignals(
  child,
  signalEmitter = process,
  signals = forwardedSignals,
) {
  const handlers = new Map();
  for (const signal of signals) {
    const handler = () => {
      try {
        child.kill(signal);
      } catch {
        // The child can exit between signal delivery and forwarding.
      }
    };
    handlers.set(signal, handler);
    signalEmitter.on(signal, handler);
  }

  return () => {
    for (const [signal, handler] of handlers) {
      signalEmitter.removeListener(signal, handler);
    }
  };
}

function spawnAndWait(
  executable,
  args,
  spawnImplementation = spawn,
  signalEmitter = process,
) {
  return new Promise((resolve, reject) => {
    const child = spawnImplementation(executable, args, {
      env: process.env,
      stdio: "inherit",
    });
    const stopForwarding = forwardSignals(child, signalEmitter);

    child.once("error", (error) => {
      stopForwarding();
      reject(error);
    });
    child.once("exit", (code, signal) => {
      stopForwarding();
      resolve({ code, signal });
    });
  });
}

async function run({
  args = process.argv.slice(2),
  resolveImplementation = resolveBinary,
  spawnImplementation = spawn,
  signalEmitter = process,
} = {}) {
  const executable = resolveImplementation();
  return spawnAndWait(executable, args, spawnImplementation, signalEmitter);
}

async function main() {
  try {
    const { code, signal } = await run();
    if (signal) {
      process.kill(process.pid, signal);
      return;
    }

    process.exitCode = code ?? 1;
  } catch (error) {
    process.stderr.write(`Failed to start telegram-owl: ${error.message}\n`);
    process.exitCode = 1;
  }
}

if (require.main === module) {
  void main();
}

module.exports = {
  forwardSignals,
  forwardedSignals,
  resolveBinary,
  run,
  spawnAndWait,
};

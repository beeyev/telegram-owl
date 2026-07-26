"use strict";

const assert = require("node:assert/strict");
const { EventEmitter } = require("node:events");
const test = require("node:test");
const {
  forwardSignals,
  resolveBinary,
  run,
  spawnAndWait,
} = require("../bin/telegram-owl");

test("resolves the platform package binary", () => {
  const calls = [];
  const result = resolveBinary({
    platform: "linux",
    arch: "x64",
    resolveImplementation(specifier, options) {
      calls.push({ specifier, options });
      return "/workspace/node_modules/@beeyev/telegram-owl-linux-x64/package.json";
    },
  });

  assert.equal(
    result,
    "/workspace/node_modules/@beeyev/telegram-owl-linux-x64/bin/telegram-owl",
  );
  assert.equal(
    calls[0].specifier,
    "@beeyev/telegram-owl-linux-x64/package.json",
  );
});

test("rejects unsupported platforms", () => {
  assert.throws(
    () => resolveBinary({ platform: "freebsd", arch: "x64" }),
    /no npm binary is available for freebsd-x64/u,
  );
});

test("reports a missing optional dependency", () => {
  assert.throws(
    () =>
      resolveBinary({
        platform: "linux",
        arch: "x64",
        resolveImplementation() {
          const error = new Error("missing");
          error.code = "MODULE_NOT_FOUND";
          throw error;
        },
      }),
    /Reinstall without --omit=optional/u,
  );
});

test("propagates unexpected package resolution errors", () => {
  const resolutionError = new Error("invalid package metadata");
  resolutionError.code = "ERR_INVALID_PACKAGE_CONFIG";

  assert.throws(
    () =>
      resolveBinary({
        platform: "linux",
        arch: "x64",
        resolveImplementation() {
          throw resolutionError;
        },
      }),
    (error) => error === resolutionError,
  );
});

test("launches the resolved native binary", async () => {
  const calls = [];
  const child = new EventEmitter();
  child.kill = () => true;
  const signalEmitter = new EventEmitter();

  const resultPromise = run({
    args: ["--version"],
    resolveImplementation: () => {
      calls.push("resolve");
      return "/cache/telegram-owl";
    },
    spawnImplementation: (executable, args, options) => {
      calls.push({ executable, args, options });
      process.nextTick(() => child.emit("exit", 0, null));
      return child;
    },
    signalEmitter,
  });

  assert.deepEqual(await resultPromise, { code: 0, signal: null });
  assert.equal(calls[0], "resolve");
  assert.equal(calls[1].executable, "/cache/telegram-owl");
  assert.deepEqual(calls[1].args, ["--version"]);
  assert.equal(calls[1].options.stdio, "inherit");
  assert.equal(calls[1].options.env, process.env);
});

test("propagates native exit status", async () => {
  const child = new EventEmitter();
  child.kill = () => true;
  const resultPromise = spawnAndWait(
    "/cache/telegram-owl",
    [],
    () => child,
    new EventEmitter(),
  );
  process.nextTick(() => child.emit("exit", 23, null));
  assert.deepEqual(await resultPromise, { code: 23, signal: null });
});

test("propagates spawn errors", async () => {
  const child = new EventEmitter();
  child.kill = () => true;
  const resultPromise = spawnAndWait(
    "/cache/telegram-owl",
    [],
    () => child,
    new EventEmitter(),
  );
  process.nextTick(() => child.emit("error", new Error("spawn failed")));
  await assert.rejects(resultPromise, /spawn failed/u);
});

test("forwards termination signals and removes handlers", () => {
  const child = {
    signals: [],
    kill(signal) {
      this.signals.push(signal);
    },
  };
  const signalEmitter = new EventEmitter();
  const stop = forwardSignals(child, signalEmitter, ["SIGTERM"]);

  signalEmitter.emit("SIGTERM");
  assert.deepEqual(child.signals, ["SIGTERM"]);
  assert.equal(signalEmitter.listenerCount("SIGTERM"), 1);

  stop();
  assert.equal(signalEmitter.listenerCount("SIGTERM"), 0);
});

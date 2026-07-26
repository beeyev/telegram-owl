"use strict";

const platforms = Object.freeze([
  Object.freeze({
    platform: "linux",
    arch: "x64",
    packageName: "@beeyev/telegram-owl-linux-x64",
    archiveSuffix: "Linux_64bit.tar.gz",
    binaryName: "telegram-owl",
  }),
  Object.freeze({
    platform: "linux",
    arch: "arm64",
    packageName: "@beeyev/telegram-owl-linux-arm64",
    archiveSuffix: "Linux_ARM64.tar.gz",
    binaryName: "telegram-owl",
  }),
  Object.freeze({
    platform: "darwin",
    arch: "x64",
    packageName: "@beeyev/telegram-owl-darwin-x64",
    archiveSuffix: "macOS_64bit.tar.gz",
    binaryName: "telegram-owl",
  }),
  Object.freeze({
    platform: "darwin",
    arch: "arm64",
    packageName: "@beeyev/telegram-owl-darwin-arm64",
    archiveSuffix: "macOS_ARM64.tar.gz",
    binaryName: "telegram-owl",
  }),
  Object.freeze({
    platform: "win32",
    arch: "x64",
    packageName: "@beeyev/telegram-owl-win32-x64",
    archiveSuffix: "Windows_64bit.zip",
    binaryName: "telegram-owl.exe",
  }),
]);

function findPlatform(platform = process.platform, arch = process.arch) {
  return platforms.find(
    (definition) =>
      definition.platform === platform && definition.arch === arch,
  );
}

function archiveName(definition, version) {
  return `telegram-owl_v${version}_${definition.archiveSuffix}`;
}

module.exports = {
  archiveName,
  findPlatform,
  platforms,
};

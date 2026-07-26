"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const { platforms } = require("../platforms");

const mainPackageName = "telegram-owl";
const versionPattern = /^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$/u;
const tagPattern = /^[0-9A-Za-z][0-9A-Za-z._-]*$/u;

function npm(args) {
  const result = spawnSync("npm", args, {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "inherit"],
  });
  if (result.status !== 0) {
    throw new Error(`npm ${args.join(" ")} failed with ${result.status}`);
  }
  return result.stdout;
}

function packageMetadata(tarball) {
  const output = npm([
    "publish",
    tarball,
    "--access",
    "public",
    "--tag",
    "dry-run",
    "--dry-run",
    "--json",
  ]);
  const metadata = JSON.parse(output);
  const result = Array.isArray(metadata)
    ? metadata[0]
    : metadata.name
      ? metadata
      : Object.values(metadata)[0];
  return {
    identity: `${result.name}@${result.version}`,
    integrity: result.integrity,
  };
}

function registryIntegrity(identity, execute = spawnSync) {
  const result = execute("npm", ["view", identity, "dist.integrity"], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "pipe"],
  });
  if (result.status === 0) {
    return result.stdout.trim();
  }
  if (/\bE404\b|404 Not Found/u.test(result.stderr)) {
    return null;
  }
  throw new Error(
    `Failed to query ${identity}: ${result.stderr.trim() || result.status}`,
  );
}

function discoverPackages(tarballRoot, version) {
  if (!versionPattern.test(version)) {
    throw new Error(`Invalid release version: ${version}`);
  }

  const expectedNames = [
    ...platforms.map((definition) => definition.packageName),
    mainPackageName,
  ];
  const expectedIdentities = new Set(
    expectedNames.map((name) => `${name}@${version}`),
  );
  const packages = new Map();

  for (const filename of fs.readdirSync(tarballRoot)) {
    if (!filename.endsWith(".tgz")) {
      continue;
    }
    const tarball = path.join(tarballRoot, filename);
    const metadata = packageMetadata(tarball);
    const { identity } = metadata;
    if (!expectedIdentities.has(identity)) {
      throw new Error(`Unexpected npm package: ${identity}`);
    }
    if (packages.has(identity)) {
      throw new Error(`Duplicate npm package: ${identity}`);
    }
    packages.set(identity, { ...metadata, tarball });
  }

  for (const identity of expectedIdentities) {
    if (!packages.has(identity)) {
      throw new Error(`Missing npm package: ${identity}`);
    }
  }

  return expectedNames.map((name) => packages.get(`${name}@${version}`));
}

function publishPackages(
  tarballRoot,
  version,
  tagOverride,
  {
    discover = discoverPackages,
    getRegistryIntegrity = registryIntegrity,
    runNpm = npm,
    output = process.stdout,
  } = {},
) {
  const packages = discover(tarballRoot, version);
  const tag = tagOverride ?? (version.includes("-") ? "next" : "latest");
  if (!tagPattern.test(tag)) {
    throw new Error(`Invalid npm dist-tag: ${tag}`);
  }

  for (const packageDefinition of packages) {
    const publishedIntegrity = getRegistryIntegrity(packageDefinition.identity);
    if (publishedIntegrity !== null) {
      if (publishedIntegrity !== packageDefinition.integrity) {
        throw new Error(
          `${packageDefinition.identity} already exists with different ` +
            "tarball integrity",
        );
      }
      output.write(`${packageDefinition.identity} is already published\n`);
      continue;
    }

    output.write(`Publishing ${packageDefinition.identity}\n`);
    runNpm([
      "publish",
      packageDefinition.tarball,
      "--access",
      "public",
      "--tag",
      tag,
    ]);
  }
}

function main() {
  const [tarballRoot, version, tag] = process.argv.slice(2);
  if (!tarballRoot || !version) {
    throw new Error(
      "Usage: publish-packages.js <tarball-root> <version> [tag]",
    );
  }
  publishPackages(path.resolve(tarballRoot), version, tag);
}

if (require.main === module) {
  try {
    main();
  } catch (error) {
    process.stderr.write(`Failed to publish npm packages: ${error.message}\n`);
    process.exitCode = 1;
  }
}

module.exports = {
  discoverPackages,
  packageMetadata,
  publishPackages,
  registryIntegrity,
};

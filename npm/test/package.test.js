"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");
const { buildPackages } = require("../scripts/build-packages");
const { discoverPackages } = require("../scripts/publish-packages");
const { findPlatform, platforms } = require("../platforms");
const { createFixtureArchives } = require("./helpers");

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    encoding: "utf8",
    ...options,
  });
  assert.equal(
    result.status,
    0,
    `${command} ${args.join(" ")} failed:\n${result.stdout}${result.stderr}`,
  );
  return result;
}

function pack(packageRoot, destination) {
  const result = run("npm", [
    "pack",
    packageRoot,
    "--pack-destination",
    destination,
    "--json",
  ]);
  const metadata = JSON.parse(result.stdout);
  if (Array.isArray(metadata)) {
    return metadata[0];
  }
  return metadata.filename ? metadata : Object.values(metadata)[0];
}

test(
  "packed CLI runs with npm 12 default script policy",
  { skip: !findPlatform() },
  async (t) => {
    const npmVersion = run("npm", ["--version"]).stdout.trim();
    assert.equal(
      Number.parseInt(npmVersion, 10),
      12,
      `package test requires npm 12, got ${npmVersion}`,
    );

    const temporaryRoot = await fs.promises.mkdtemp(
      path.join(os.tmpdir(), "telegram-owl-package-test-"),
    );
    t.after(() =>
      fs.promises.rm(temporaryRoot, { force: true, recursive: true }),
    );

    const version = "0.0.1-package-test";
    const archiveRoot = path.join(temporaryRoot, "archives");
    const packageRoot = path.join(temporaryRoot, "packages");
    const tarballRoot = path.join(temporaryRoot, "tarballs");
    await createFixtureArchives(archiveRoot, version);
    await buildPackages(archiveRoot, version, packageRoot);
    await fs.promises.mkdir(tarballRoot);

    const packageMetadata = [];
    for (const directory of await fs.promises.readdir(packageRoot)) {
      packageMetadata.push(
        pack(path.join(packageRoot, directory), tarballRoot),
      );
    }
    assert.equal(packageMetadata.length, platforms.length + 1);
    const publishOrder = discoverPackages(tarballRoot, version);
    assert.deepEqual(
      publishOrder.map((packageDefinition) => packageDefinition.identity),
      [
        ...platforms.map(
          (definition) => `${definition.packageName}@${version}`,
        ),
        `telegram-owl@${version}`,
      ],
    );
    for (const packageDefinition of publishOrder) {
      assert.match(packageDefinition.integrity, /^sha512-/u);
    }

    const mainMetadata = packageMetadata.find(
      (metadata) => metadata.name === "telegram-owl",
    );
    assert.ok(mainMetadata);
    const launcher = mainMetadata.files.find(
      (file) => file.path === "bin/telegram-owl.js",
    );
    assert.notEqual(launcher.mode & 0o111, 0);

    const currentPlatform = findPlatform();
    const platformMetadata = packageMetadata.find(
      (metadata) => metadata.name === currentPlatform.packageName,
    );
    assert.ok(platformMetadata);
    const binary = platformMetadata.files.find(
      (file) => file.path === `bin/${currentPlatform.binaryName}`,
    );
    assert.notEqual(binary.mode & 0o111, 0);

    const testMainRoot = path.join(temporaryRoot, "test-main");
    await fs.promises.cp(path.join(packageRoot, "telegram-owl"), testMainRoot, {
      recursive: true,
    });
    const testMainManifestPath = path.join(testMainRoot, "package.json");
    const testMainManifest = JSON.parse(
      await fs.promises.readFile(testMainManifestPath, "utf8"),
    );
    testMainManifest.optionalDependencies = {
      [currentPlatform.packageName]: `file:${path.join(tarballRoot, platformMetadata.filename)}`,
    };
    await fs.promises.writeFile(
      testMainManifestPath,
      `${JSON.stringify(testMainManifest, null, 2)}\n`,
    );
    const testMainMetadata = pack(testMainRoot, tarballRoot);

    const installationRoot = path.join(temporaryRoot, "installation");
    await fs.promises.mkdir(installationRoot);
    await fs.promises.writeFile(
      path.join(installationRoot, "package.json"),
      '{"name":"package-test","private":true}\n',
    );
    const installation = run(
      "npm",
      [
        "install",
        path.join(tarballRoot, testMainMetadata.filename),
        "--no-audit",
        "--no-fund",
      ],
      {
        cwd: installationRoot,
        env: {
          ...process.env,
          npm_config_cache: path.join(temporaryRoot, "npm-cache"),
          npm_config_update_notifier: "false",
        },
      },
    );
    assert.doesNotMatch(installation.stderr, /Blocked build scripts/u);

    const command =
      process.platform === "win32"
        ? path.join(
            installationRoot,
            "node_modules",
            ".bin",
            "telegram-owl.cmd",
          )
        : path.join(installationRoot, "node_modules", ".bin", "telegram-owl");
    const execution = run(command, ["--version"], {
      cwd: installationRoot,
    });
    assert.equal(execution.stdout, `${currentPlatform.packageName} fixture\n`);
  },
);

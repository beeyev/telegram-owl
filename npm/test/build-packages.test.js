"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { pipeline } = require("node:stream/promises");
const test = require("node:test");
const tar = require("tar");
const yazl = require("yazl");
const {
  buildPackages,
  extractTarBinary,
  extractZipBinary,
  mainManifest,
  parseChecksums,
  platformManifest,
  validateBinarySize,
} = require("../scripts/build-packages");
const { archiveName, platforms } = require("../platforms");
const { createFixtureArchives } = require("./helpers");

function createDuplicateZip(filename, binaryName) {
  const zipFile = new yazl.ZipFile();
  zipFile.addBuffer(Buffer.from("first"), binaryName);
  zipFile.addBuffer(Buffer.from("second"), binaryName);
  zipFile.end();
  return pipeline(zipFile.outputStream, fs.createWriteStream(filename));
}

test("defines the five supported npm platforms without duplicates", () => {
  assert.equal(platforms.length, 5);
  assert.equal(
    new Set(
      platforms.map(
        (definition) => `${definition.platform}-${definition.arch}`,
      ),
    ).size,
    platforms.length,
  );
  assert.equal(
    new Set(platforms.map((definition) => definition.packageName)).size,
    platforms.length,
  );
});

test("generates exact platform and main package metadata", () => {
  const baseManifest = {
    name: "telegram-owl",
    description: "description",
    license: "MIT",
    author: "author",
    homepage: "https://example.com",
    bugs: { url: "https://example.com/issues" },
    repository: { type: "git", url: "https://example.com/repo.git" },
    private: true,
    scripts: { test: "node --test" },
    devDependencies: { tar: "1.0.0" },
    files: ["bin", "platforms.js"],
  };
  const version = "1.2.3-rc.1";
  const platform = platformManifest(baseManifest, platforms[0], version);
  const main = mainManifest(baseManifest, version);

  assert.deepEqual(platform.os, ["linux"]);
  assert.deepEqual(platform.cpu, ["x64"]);
  assert.equal(platform.version, version);
  assert.equal(platform.scripts, undefined);
  assert.equal(main.version, version);
  assert.equal(main.scripts, undefined);
  assert.equal(main.devDependencies, undefined);
  assert.equal(main.private, undefined);
  assert.deepEqual(
    main.optionalDependencies,
    Object.fromEntries(
      platforms.map((definition) => [definition.packageName, version]),
    ),
  );
});

test("parses checksums and rejects malformed or duplicate entries", () => {
  const checksum = "a".repeat(64);
  assert.equal(
    parseChecksums(`${checksum}  archive.tar.gz\n`).get("archive.tar.gz"),
    checksum,
  );
  assert.throws(() => parseChecksums("invalid\n"), /Invalid checksum line/u);
  assert.throws(
    () =>
      parseChecksums(
        `${checksum}  archive.tar.gz\n${checksum}  archive.tar.gz\n`,
      ),
    /Duplicate checksum/u,
  );
});

test("rejects empty, invalid, and oversized archive entries", () => {
  assert.throws(
    () => validateBinarySize(0, "telegram-owl", "archive.tar.gz"),
    /Invalid telegram-owl size/u,
  );
  assert.throws(
    () => validateBinarySize(Number.NaN, "telegram-owl", "archive.tar.gz"),
    /Invalid telegram-owl size/u,
  );
  assert.throws(
    () =>
      validateBinarySize(
        128 * 1024 * 1024 + 1,
        "telegram-owl",
        "archive.tar.gz",
      ),
    /Invalid telegram-owl size/u,
  );
  assert.doesNotThrow(() =>
    validateBinarySize(128 * 1024 * 1024, "telegram-owl", "archive.tar.gz"),
  );
});

test("rejects malformed tar and zip archives without uncaught errors", async (t) => {
  const temporaryRoot = await fs.promises.mkdtemp(
    path.join(os.tmpdir(), "telegram-owl-malformed-archives-"),
  );
  t.after(() =>
    fs.promises.rm(temporaryRoot, { force: true, recursive: true }),
  );

  const tarSource = path.join(temporaryRoot, "tar-source");
  const tarArchive = path.join(temporaryRoot, "invalid.tar.gz");
  await fs.promises.mkdir(tarSource);
  await fs.promises.symlink(
    "missing-target",
    path.join(tarSource, "telegram-owl"),
  );
  await tar.c({ cwd: tarSource, file: tarArchive, gzip: true }, [
    "telegram-owl",
  ]);

  await assert.rejects(
    extractTarBinary(
      tarArchive,
      path.join(temporaryRoot, "tar-output", "telegram-owl"),
      "telegram-owl",
    ),
    /is not a regular file/u,
  );

  const zipArchive = path.join(temporaryRoot, "duplicate.zip");
  await createDuplicateZip(zipArchive, "telegram-owl.exe");
  await assert.rejects(
    extractZipBinary(
      zipArchive,
      path.join(temporaryRoot, "zip-output", "telegram-owl.exe"),
      "telegram-owl.exe",
    ),
    /Duplicate telegram-owl\.exe entry/u,
  );
});

test("builds packages from verified release archives", async (t) => {
  const temporaryRoot = await fs.promises.mkdtemp(
    path.join(os.tmpdir(), "telegram-owl-build-packages-"),
  );
  t.after(() =>
    fs.promises.rm(temporaryRoot, { force: true, recursive: true }),
  );
  const version = "1.2.3-rc.1";
  const archiveRoot = path.join(temporaryRoot, "archives");
  const outputRoot = path.join(temporaryRoot, "packages");
  await createFixtureArchives(archiveRoot, version);

  await buildPackages(archiveRoot, version, outputRoot);

  const main = JSON.parse(
    await fs.promises.readFile(
      path.join(outputRoot, "telegram-owl", "package.json"),
      "utf8",
    ),
  );
  assert.equal(main.version, version);
  assert.equal(Object.keys(main.optionalDependencies).length, 5);

  for (const definition of platforms) {
    const directoryName = definition.packageName.replace("@beeyev/", "");
    const packageRoot = path.join(outputRoot, directoryName);
    const manifest = JSON.parse(
      await fs.promises.readFile(
        path.join(packageRoot, "package.json"),
        "utf8",
      ),
    );
    assert.equal(manifest.name, definition.packageName);
    assert.deepEqual(manifest.os, [definition.platform]);
    assert.deepEqual(manifest.cpu, [definition.arch]);

    const binary = path.join(packageRoot, "bin", definition.binaryName);
    const stat = await fs.promises.stat(binary);
    assert.equal(stat.isFile(), true);
    assert.notEqual(stat.mode & 0o111, 0);
  }
});

test("fails when an archive checksum does not match", async (t) => {
  const temporaryRoot = await fs.promises.mkdtemp(
    path.join(os.tmpdir(), "telegram-owl-checksum-failure-"),
  );
  t.after(() =>
    fs.promises.rm(temporaryRoot, { force: true, recursive: true }),
  );
  const version = "1.2.3";
  const archiveRoot = path.join(temporaryRoot, "archives");
  await createFixtureArchives(archiveRoot, version);
  const archive = path.join(archiveRoot, archiveName(platforms[0], version));
  await fs.promises.appendFile(archive, "tampered");

  await assert.rejects(
    buildPackages(archiveRoot, version, path.join(temporaryRoot, "packages")),
    /Checksum mismatch/u,
  );
});

test("fails clearly when the output directory already exists", async (t) => {
  const temporaryRoot = await fs.promises.mkdtemp(
    path.join(os.tmpdir(), "telegram-owl-existing-output-"),
  );
  t.after(() =>
    fs.promises.rm(temporaryRoot, { force: true, recursive: true }),
  );
  const outputRoot = path.join(temporaryRoot, "packages");
  await fs.promises.mkdir(outputRoot);

  await assert.rejects(
    buildPackages(path.join(temporaryRoot, "archives"), "1.2.3", outputRoot),
    /Output directory already exists/u,
  );
});

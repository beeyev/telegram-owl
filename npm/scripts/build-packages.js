"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");
const { pipeline } = require("node:stream/promises");
const tar = require("tar");
const yauzl = require("yauzl");
const { archiveName, platforms } = require("../platforms");

const packageRoot = path.resolve(__dirname, "..");
const maximumBinarySize = 128 * 1024 * 1024;
const versionPattern = /^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$/u;

function validateBinarySize(size, binaryName, archive) {
  if (!Number.isSafeInteger(size) || size <= 0 || size > maximumBinarySize) {
    throw new Error(`Invalid ${binaryName} size in ${archive}: ${size}`);
  }
}

function parseChecksums(contents) {
  const checksums = new Map();
  for (const line of contents.split(/\r?\n/u)) {
    if (!line.trim()) {
      continue;
    }

    const match = /^([a-fA-F0-9]{64})\s+\*?(.+)$/u.exec(line);
    if (!match) {
      throw new Error(`Invalid checksum line: ${line}`);
    }

    const [, checksum, filename] = match;
    if (checksums.has(filename)) {
      throw new Error(`Duplicate checksum for ${filename}`);
    }
    checksums.set(filename, checksum.toLowerCase());
  }
  return checksums;
}

async function sha256(filename) {
  const hash = crypto.createHash("sha256");
  for await (const chunk of fs.createReadStream(filename)) {
    hash.update(chunk);
  }
  return hash.digest("hex");
}

async function verifyArchive(filename, expectedChecksum) {
  if (!expectedChecksum) {
    throw new Error(`No checksum found for ${path.basename(filename)}`);
  }

  const actualChecksum = await sha256(filename);
  if (actualChecksum !== expectedChecksum) {
    throw new Error(
      `Checksum mismatch for ${path.basename(filename)}: ` +
        `expected ${expectedChecksum}, got ${actualChecksum}`,
    );
  }
}

async function extractTarBinary(archive, destination, binaryName) {
  let matches = 0;
  let validationError;
  const destinationDirectory = path.dirname(destination);
  await fs.promises.mkdir(destinationDirectory, { recursive: true });
  await tar.x({
    cwd: destinationDirectory,
    file: archive,
    filter(entryPath, entry) {
      if (entryPath !== binaryName) {
        return false;
      }
      matches += 1;
      if (matches > 1) {
        validationError ??= new Error(
          `Duplicate ${binaryName} entry in ${archive}`,
        );
        return false;
      }
      if (entry.type !== "File") {
        validationError = new Error(
          `${binaryName} is not a regular file in ${archive}`,
        );
        return false;
      }
      try {
        validateBinarySize(entry.size, binaryName, archive);
      } catch (error) {
        validationError = error;
        return false;
      }
      return true;
    },
    strict: true,
  });

  if (validationError) {
    throw validationError;
  }
  if (matches !== 1) {
    throw new Error(
      `Expected one ${binaryName} entry in ${archive}, found ${matches}`,
    );
  }
}

function extractZipBinary(archive, destination, binaryName) {
  return new Promise((resolve, reject) => {
    yauzl.open(
      archive,
      { lazyEntries: true, validateEntrySizes: true },
      (openError, zipFile) => {
        if (openError) {
          reject(openError);
          return;
        }

        let matches = 0;
        let settled = false;
        const finish = (error) => {
          if (settled) {
            return;
          }
          settled = true;
          zipFile.close();
          if (error) {
            reject(error);
          } else {
            resolve();
          }
        };

        zipFile.once("error", finish);
        zipFile.once("end", () => {
          if (matches !== 1) {
            finish(
              new Error(
                `Expected one ${binaryName} entry in ${archive}, ` +
                  `found ${matches}`,
              ),
            );
            return;
          }
          finish();
        });
        zipFile.on("entry", (entry) => {
          if (entry.fileName !== binaryName) {
            zipFile.readEntry();
            return;
          }
          try {
            validateBinarySize(entry.uncompressedSize, binaryName, archive);
          } catch (error) {
            finish(error);
            return;
          }

          matches += 1;
          if (matches > 1) {
            finish(new Error(`Duplicate ${binaryName} entry in ${archive}`));
            return;
          }

          zipFile.openReadStream(entry, async (streamError, stream) => {
            if (streamError) {
              finish(streamError);
              return;
            }
            try {
              await fs.promises.mkdir(path.dirname(destination), {
                recursive: true,
              });
              await pipeline(
                stream,
                fs.createWriteStream(destination, {
                  flags: "wx",
                  mode: 0o755,
                }),
              );
              zipFile.readEntry();
            } catch (error) {
              finish(error);
            }
          });
        });

        zipFile.readEntry();
      },
    );
  });
}

async function extractBinary(archive, destination, binaryName) {
  if (archive.endsWith(".tar.gz")) {
    await extractTarBinary(archive, destination, binaryName);
  } else if (archive.endsWith(".zip")) {
    await extractZipBinary(archive, destination, binaryName);
  } else {
    throw new Error(`Unsupported archive format: ${archive}`);
  }

  await fs.promises.chmod(destination, 0o755);
  const stat = await fs.promises.stat(destination);
  if (!stat.isFile() || stat.size <= 0 || stat.size > maximumBinarySize) {
    throw new Error(`Invalid extracted binary: ${destination}`);
  }
}

async function writeJSON(filename, value) {
  await fs.promises.writeFile(filename, `${JSON.stringify(value, null, 2)}\n`);
}

async function copyCommonFiles(destination) {
  await Promise.all(
    ["LICENSE", "README.md"].map((filename) =>
      fs.promises.copyFile(
        path.join(packageRoot, filename),
        path.join(destination, filename),
      ),
    ),
  );
}

function platformManifest(baseManifest, definition, version) {
  return {
    name: definition.packageName,
    version,
    description:
      `${baseManifest.description} ` +
      `(${definition.platform}-${definition.arch})`,
    license: baseManifest.license,
    author: baseManifest.author,
    homepage: baseManifest.homepage,
    bugs: baseManifest.bugs,
    repository: baseManifest.repository,
    os: [definition.platform],
    cpu: [definition.arch],
    files: ["bin", "LICENSE", "README.md"],
    publishConfig: { access: "public" },
  };
}

function mainManifest(baseManifest, version) {
  const manifest = {
    ...baseManifest,
    version,
    optionalDependencies: Object.fromEntries(
      platforms.map((definition) => [definition.packageName, version]),
    ),
  };
  delete manifest.dependencies;
  delete manifest.devDependencies;
  delete manifest.private;
  delete manifest.scripts;
  return manifest;
}

async function buildPackages(archiveRoot, version, outputRoot) {
  if (!versionPattern.test(version)) {
    throw new Error(`Invalid release version: ${version}`);
  }

  const resolvedArchiveRoot = path.resolve(archiveRoot);
  const resolvedOutputRoot = path.resolve(outputRoot);
  if (resolvedArchiveRoot === resolvedOutputRoot) {
    throw new Error("Archive and output directories must be different");
  }

  try {
    await fs.promises.mkdir(resolvedOutputRoot);
  } catch (error) {
    if (error.code === "EEXIST") {
      throw new Error(`Output directory already exists: ${resolvedOutputRoot}`);
    }
    throw error;
  }
  const baseManifest = JSON.parse(
    await fs.promises.readFile(path.join(packageRoot, "package.json"), "utf8"),
  );
  const checksumFilename = path.join(
    resolvedArchiveRoot,
    `telegram-owl_v${version}_checksums.txt`,
  );
  const checksums = parseChecksums(
    await fs.promises.readFile(checksumFilename, "utf8"),
  );

  for (const definition of platforms) {
    const filename = archiveName(definition, version);
    const archive = path.join(resolvedArchiveRoot, filename);
    await verifyArchive(archive, checksums.get(filename));

    const directoryName = definition.packageName.replace("@beeyev/", "");
    const destination = path.join(resolvedOutputRoot, directoryName);
    await fs.promises.mkdir(destination);
    await copyCommonFiles(destination);
    await writeJSON(
      path.join(destination, "package.json"),
      platformManifest(baseManifest, definition, version),
    );
    await extractBinary(
      archive,
      path.join(destination, "bin", definition.binaryName),
      definition.binaryName,
    );
  }

  const mainDestination = path.join(resolvedOutputRoot, "telegram-owl");
  await fs.promises.mkdir(path.join(mainDestination, "bin"), {
    recursive: true,
  });
  await copyCommonFiles(mainDestination);
  await Promise.all([
    fs.promises.copyFile(
      path.join(packageRoot, "bin", "telegram-owl.js"),
      path.join(mainDestination, "bin", "telegram-owl.js"),
    ),
    fs.promises.copyFile(
      path.join(packageRoot, "platforms.js"),
      path.join(mainDestination, "platforms.js"),
    ),
  ]);
  await fs.promises.chmod(
    path.join(mainDestination, "bin", "telegram-owl.js"),
    0o755,
  );
  await writeJSON(
    path.join(mainDestination, "package.json"),
    mainManifest(baseManifest, version),
  );
}

async function main() {
  const [archiveRoot, version, outputRoot] = process.argv.slice(2);
  if (!archiveRoot || !version || !outputRoot) {
    throw new Error(
      "Usage: build-packages.js <archive-root> <version> <output-root>",
    );
  }
  await buildPackages(archiveRoot, version, outputRoot);
}

if (require.main === module) {
  main().catch((error) => {
    process.stderr.write(`Failed to build npm packages: ${error.message}\n`);
    process.exitCode = 1;
  });
}

module.exports = {
  buildPackages,
  extractTarBinary,
  extractZipBinary,
  mainManifest,
  parseChecksums,
  platformManifest,
  validateBinarySize,
  verifyArchive,
};

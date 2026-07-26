"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");
const { pipeline } = require("node:stream/promises");
const tar = require("tar");
const yazl = require("yazl");
const { archiveName, platforms } = require("../platforms");

async function fileSHA256(filename) {
  const hash = crypto.createHash("sha256");
  for await (const chunk of fs.createReadStream(filename)) {
    hash.update(chunk);
  }
  return hash.digest("hex");
}

function createZip(source, filename, binaryName) {
  return new Promise((resolve, reject) => {
    const zipFile = new yazl.ZipFile();
    zipFile.addFile(source, binaryName, { mode: 0o100755 });
    zipFile.end();
    pipeline(zipFile.outputStream, fs.createWriteStream(filename)).then(
      resolve,
      reject,
    );
  });
}

async function createFixtureArchives(root, version) {
  await fs.promises.mkdir(root, { recursive: true });
  const checksumLines = [];

  for (const definition of platforms) {
    const sourceRoot = path.join(
      root,
      `source-${definition.platform}-${definition.arch}`,
    );
    await fs.promises.mkdir(sourceRoot);
    const source = path.join(sourceRoot, definition.binaryName);
    await fs.promises.writeFile(
      source,
      "#!/usr/bin/env node\n" +
        `process.stdout.write(${JSON.stringify(
          `${definition.packageName} fixture\n`,
        )});\n`,
      { mode: 0o755 },
    );

    const filename = archiveName(definition, version);
    const archive = path.join(root, filename);
    if (filename.endsWith(".tar.gz")) {
      await tar.c({ cwd: sourceRoot, file: archive, gzip: true }, [
        definition.binaryName,
      ]);
    } else {
      await createZip(source, archive, definition.binaryName);
    }
    checksumLines.push(`${await fileSHA256(archive)}  ${filename}`);
  }

  await fs.promises.writeFile(
    path.join(root, `telegram-owl_v${version}_checksums.txt`),
    `${checksumLines.join("\n")}\n`,
  );
}

module.exports = {
  createFixtureArchives,
  fileSHA256,
};

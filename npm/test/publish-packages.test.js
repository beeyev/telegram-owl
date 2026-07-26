"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const {
  publishPackages,
  registryIntegrity,
} = require("../scripts/publish-packages");
const { platforms } = require("../platforms");

function packageDefinitions(version) {
  return [
    ...platforms.map((definition) => definition.packageName),
    "telegram-owl",
  ].map((name, index) => ({
    identity: `${name}@${version}`,
    integrity: `sha512-integrity-${index}`,
    tarball: `/tarballs/package-${index}.tgz`,
  }));
}

function publisherDependencies(packages, published = new Map()) {
  const publishCalls = [];
  const output = [];
  return {
    dependencies: {
      discover() {
        return packages;
      },
      getRegistryIntegrity(identity) {
        return published.get(identity) ?? null;
      },
      runNpm(args) {
        publishCalls.push(args);
      },
      output: {
        write(message) {
          output.push(message);
        },
      },
    },
    output,
    publishCalls,
  };
}

test("publishes stable packages in platform-first order with latest tag", () => {
  const version = "1.2.3";
  const packages = packageDefinitions(version);
  const { dependencies, publishCalls } = publisherDependencies(packages);

  publishPackages("/tarballs", version, undefined, dependencies);

  assert.deepEqual(
    publishCalls.map((args) => args[1]),
    packages.map((definition) => definition.tarball),
  );
  for (const args of publishCalls) {
    assert.deepEqual(args.slice(2), ["--access", "public", "--tag", "latest"]);
  }
});

test("uses next for prereleases and accepts an explicit tag", () => {
  const version = "1.2.3-rc.1";
  const packages = packageDefinitions(version);
  const first = publisherDependencies(packages);
  publishPackages("/tarballs", version, undefined, first.dependencies);
  assert.equal(first.publishCalls[0].at(-1), "next");

  const second = publisherDependencies(packages);
  publishPackages("/tarballs", version, "bootstrap", second.dependencies);
  assert.equal(second.publishCalls[0].at(-1), "bootstrap");
});

test("skips matching published packages and rejects integrity changes", () => {
  const version = "1.2.3";
  const packages = packageDefinitions(version);
  const matching = new Map([[packages[0].identity, packages[0].integrity]]);
  const first = publisherDependencies(packages, matching);

  publishPackages("/tarballs", version, undefined, first.dependencies);

  assert.equal(first.publishCalls.length, packages.length - 1);
  assert.match(first.output.join(""), /is already published/u);

  const mismatching = new Map([[packages[0].identity, "sha512-different"]]);
  const second = publisherDependencies(packages, mismatching);
  assert.throws(
    () => publishPackages("/tarballs", version, undefined, second.dependencies),
    /already exists with different tarball integrity/u,
  );
  assert.equal(second.publishCalls.length, 0);
});

test("rejects invalid dist-tags before querying the registry", () => {
  const version = "1.2.3";
  const packages = packageDefinitions(version);
  const { dependencies, publishCalls } = publisherDependencies(packages);
  let queried = false;
  dependencies.getRegistryIntegrity = () => {
    queried = true;
    return null;
  };

  assert.throws(
    () => publishPackages("/tarballs", version, "invalid/tag", dependencies),
    /Invalid npm dist-tag/u,
  );
  assert.equal(queried, false);
  assert.equal(publishCalls.length, 0);
});

test("registry lookup treats E404 as absent and propagates other errors", () => {
  const missing = registryIntegrity("@scope/package@1.2.3", () => ({
    status: 1,
    stderr: "npm error code E404\nnpm error 404 Not Found",
    stdout: "",
  }));
  assert.equal(missing, null);

  assert.throws(
    () =>
      registryIntegrity("@scope/package@1.2.3", () => ({
        status: 1,
        stderr: "npm error code E401",
        stdout: "",
      })),
    /Failed to query @scope\/package@1.2.3: npm error code E401/u,
  );
});

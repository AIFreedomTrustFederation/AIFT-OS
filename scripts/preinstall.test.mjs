import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const { preinstall } = JSON.parse(
  readFileSync(path.join(root, "package.json"), "utf8"),
).scripts;

function checkPreinstall(userAgent) {
  const cwd = mkdtempSync(path.join(tmpdir(), "aift-preinstall-"));
  try {
    const lockfiles = ["package-lock.json", "yarn.lock"];
    for (const name of lockfiles) {
      writeFileSync(path.join(cwd, name), `existing ${name}\n`);
    }
    const result = spawnSync("sh", ["-c", preinstall], {
      cwd,
      env: { ...process.env, npm_config_user_agent: userAgent },
      encoding: "utf8",
    });
    const files = {};
    for (const name of lockfiles) {
      const filename = path.join(cwd, name);
      files[name] = existsSync(filename) ? readFileSync(filename, "utf8") : null;
    }
    return { status: result.status, stderr: result.stderr, files };
  } finally {
    rmSync(cwd, { recursive: true, force: true });
  }
}

test("npm is rejected without deleting existing lockfiles", () => {
  const result = checkPreinstall("npm/10.9.2 node/v22.16.0");
  assert.equal(result.status, 1);
  assert.match(result.stderr, /Use pnpm instead/);
  assert.equal(result.files["package-lock.json"], "existing package-lock.json\n");
  assert.equal(result.files["yarn.lock"], "existing yarn.lock\n");
});

test("yarn is rejected without deleting existing lockfiles", () => {
  const result = checkPreinstall("yarn/1.22.22 npm/? node/v22.16.0");
  assert.equal(result.status, 1);
  assert.equal(result.files["package-lock.json"], "existing package-lock.json\n");
  assert.equal(result.files["yarn.lock"], "existing yarn.lock\n");
});

test("missing package-manager identity is rejected without deleting lockfiles", () => {
  const result = checkPreinstall("");
  assert.equal(result.status, 1);
  assert.equal(result.files["package-lock.json"], "existing package-lock.json\n");
  assert.equal(result.files["yarn.lock"], "existing yarn.lock\n");
});

test("pnpm retains the existing alternate-lockfile cleanup behavior", () => {
  const result = checkPreinstall("pnpm/11.25.0 npm/? node/v24.0.0");
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.files["package-lock.json"], null);
  assert.equal(result.files["yarn.lock"], null);
});

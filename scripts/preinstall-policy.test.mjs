import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { test } from 'node:test';

const packageJson = JSON.parse(readFileSync(new URL('../package.json', import.meta.url), 'utf8'));
const guard = packageJson.scripts.preinstall;

function runInDisposableCheckout(userAgent) {
  const directory = mkdtempSync(join(tmpdir(), 'aift-install-guard-'));
  const files = ['package-lock.json', 'yarn.lock'];
  const sentinels = ['package-lock-sentinel', 'yarn-lock-sentinel'];
  try {
    files.forEach((name, index) => writeFileSync(join(directory, name), sentinels[index]));
    const result = spawnSync('sh', ['-c', guard], {
      cwd: directory,
      env: { ...process.env, npm_config_user_agent: userAgent },
      encoding: 'utf8',
    });
    assert.equal(result.error, undefined, `unable to run install guard: ${result.error}`);
    files.forEach((name, index) => {
      assert.equal(readFileSync(join(directory, name), 'utf8'), sentinels[index], `${name} was altered`);
    });
    return result;
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

test('pnpm install is accepted without modifying either lockfile', () => {
  const result = runInDisposableCheckout('pnpm/11.25.0 npm/? node/v24.0.0');
  assert.equal(result.status, 0, result.stderr);
});

test('npm install is rejected without modifying either lockfile', () => {
  const result = runInDisposableCheckout('npm/10.9.2 node/v22.0.0');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /Use pnpm instead/);
});

test('unknown installer is rejected without modifying either lockfile', () => {
  const result = runInDisposableCheckout('unknown/1.0');
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /Use pnpm instead/);
});

import { readdirSync, statSync } from "node:fs";
import { spawnSync } from "node:child_process";
import path from "node:path";

const roots = ["artifacts", "lib", "scripts"];
const testPattern = /\.(test|spec)\.[cm]?[jt]sx?$/;
const ignored = new Set(["node_modules", "dist", "build", ".next", ".cache"]);

function walk(dir, files = []) {
  let entries;
  try {
    entries = readdirSync(dir);
  } catch {
    return files;
  }

  for (const entry of entries) {
    if (ignored.has(entry)) {
      continue;
    }
    const fullPath = path.join(dir, entry);
    const stat = statSync(fullPath);
    if (stat.isDirectory()) {
      walk(fullPath, files);
    } else if (testPattern.test(entry)) {
      files.push(fullPath);
    }
  }
  return files;
}

const tests = roots.flatMap((root) => walk(root));

if (tests.length === 0) {
  console.log("SKIP: no JavaScript or TypeScript test files found in active workspace packages.");
  process.exit(0);
}

const result = spawnSync(process.execPath, ["--test", ...tests], {
  stdio: "inherit",
  shell: false,
});

process.exit(result.status ?? 1);

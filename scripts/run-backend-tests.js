#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

const rootDir = path.resolve(__dirname, '..');
const archiveDir = path.join(rootDir, 'tests', '_backend');
const backendDir = path.join(rootDir, 'backend');
const copiedFiles = [];
const touchedDirs = new Set();
let cleaned = false;

function walkFiles(dir) {
  if (!fs.existsSync(dir)) {
    return [];
  }

  const entries = fs.readdirSync(dir, { withFileTypes: true });
  return entries.flatMap((entry) => {
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      return walkFiles(fullPath);
    }
    return fullPath.endsWith('_test.go') ? [fullPath] : [];
  });
}

function sameFile(left, right) {
  return fs.readFileSync(left).equals(fs.readFileSync(right));
}

function toArchivedDestination(sourcePath) {
  const relativePath = path.relative(archiveDir, sourcePath);
  const destinationPath = path.join(backendDir, relativePath);

  if (!destinationPath.startsWith(backendDir + path.sep)) {
    throw new Error(`Refusing to materialize outside backend: ${destinationPath}`);
  }

  return destinationPath;
}

function materializeBackendTests() {
  const testFiles = walkFiles(archiveDir);

  for (const sourcePath of testFiles) {
    const destinationPath = toArchivedDestination(sourcePath);
    const destinationDir = path.dirname(destinationPath);

    fs.mkdirSync(destinationDir, { recursive: true });
    touchedDirs.add(destinationDir);

    if (fs.existsSync(destinationPath)) {
      if (!sameFile(sourcePath, destinationPath)) {
        throw new Error(`Refusing to overwrite existing backend test file: ${destinationPath}`);
      }
      fs.rmSync(destinationPath, { force: true });
    }

    fs.copyFileSync(sourcePath, destinationPath);
    copiedFiles.push(destinationPath);
  }
}

function removeEmptyTouchedDirs() {
  const dirs = [...touchedDirs].sort((a, b) => b.length - a.length);

  for (const dir of dirs) {
    let current = dir;
    while (current.startsWith(backendDir + path.sep) && current !== backendDir) {
      try {
        fs.rmdirSync(current);
      } catch {
        break;
      }
      current = path.dirname(current);
    }
  }
}

function cleanup() {
  if (cleaned) {
    return;
  }
  cleaned = true;

  for (const filePath of copiedFiles.reverse()) {
    fs.rmSync(filePath, { force: true });
  }
  removeEmptyTouchedDirs();
}

function main() {
  const goArgs = process.argv.slice(2);
  const args = goArgs.length > 0 ? goArgs : ['./backend/...'];

  process.on('exit', cleanup);

  try {
    materializeBackendTests();
    const result = spawnSync('go', ['test', ...args], {
      cwd: rootDir,
      stdio: 'inherit',
      env: { ...process.env },
    });

    cleanup();

    if (result.error) {
      throw result.error;
    }
    process.exit(result.status ?? 1);
  } catch (error) {
    cleanup();
    console.error(error.message);
    process.exit(1);
  }
}

main();

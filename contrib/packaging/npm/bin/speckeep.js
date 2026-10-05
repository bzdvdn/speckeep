#!/usr/bin/env node
'use strict';

// Thin launcher: execs the native speckeep binary downloaded by
// scripts/install.js (or pointed to via SPECKEEP_BINARY_PATH), forwarding
// argv/stdio/exit-code transparently. All real logic lives in the Go binary.

const fs = require('fs');
const path = require('path');
const { spawnSync } = require('child_process');

function fail(msg) {
  console.error(`[speckeep] error: ${msg}`);
  process.exit(1);
}

function resolveBinary() {
  if (process.env.SPECKEEP_BINARY_PATH) {
    return process.env.SPECKEEP_BINARY_PATH;
  }
  const binName = process.platform === 'win32' ? 'speckeep.exe' : 'speckeep';
  return path.join(__dirname, '..', '.bin', binName);
}

function main() {
  const binaryPath = resolveBinary();

  if (!fs.existsSync(binaryPath)) {
    fail(
      `native binary not found at ${binaryPath}.\n` +
        'The postinstall download may have failed or been skipped ' +
        '(e.g. `npm install --ignore-scripts`). Re-run:\n' +
        '  node ' +
        path.join(__dirname, '..', 'scripts', 'install.js') +
        '\nor install speckeep directly: https://github.com/bzdvdn/speckeep#install'
    );
  }

  const result = spawnSync(binaryPath, process.argv.slice(2), {
    stdio: 'inherit',
  });

  if (result.error) {
    fail(`failed to launch ${binaryPath}: ${result.error.message}`);
  }
  process.exit(result.status === null ? 1 : result.status);
}

main();

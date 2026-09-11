#!/usr/bin/env node
'use strict';

// Postinstall: downloads the speckeep release binary matching this npm
// package's version + the current OS/arch from GitHub Releases, verifies its
// sha256 against the release's sha256sum.txt, and extracts it into
// ./.bin/ next to this script. bin/speckeep.js execs that binary at runtime.
//
// This package ships no compiled code of its own — speckeep is a Go binary,
// this is just a launcher so `npx speckeep` / `npm install -g speckeep` work.

const fs = require('fs');
const path = require('path');
const https = require('https');
const crypto = require('crypto');
const { spawnSync } = require('child_process');

const REPO_OWNER = 'bzdvdn';
const REPO_NAME = 'speckeep';
const PKG_ROOT = path.join(__dirname, '..');
const BIN_DIR = path.join(PKG_ROOT, '.bin');
const MAX_REDIRECTS = 5;

function log(msg) {
  console.log(`[speckeep] ${msg}`);
}

function fail(msg) {
  console.error(`[speckeep] error: ${msg}`);
  process.exit(1);
}

function resolveTag() {
  if (process.env.SPECKEEP_VERSION) {
    const v = process.env.SPECKEEP_VERSION;
    return v.startsWith('v') ? v : `v${v}`;
  }
  const pkg = require(path.join(PKG_ROOT, 'package.json'));
  return `v${pkg.version}`;
}

function resolvePlatform() {
  const goosMap = { linux: 'linux', darwin: 'darwin', win32: 'windows' };
  const goarchMap = { x64: 'amd64', arm64: 'arm64' };

  const goos = goosMap[process.platform];
  const goarch = goarchMap[process.arch];

  if (!goos || !goarch) {
    fail(
      `unsupported platform ${process.platform}/${process.arch}. ` +
        'speckeep publishes linux/darwin/windows on amd64/arm64. ' +
        'Install manually: https://github.com/bzdvdn/speckeep#install'
    );
  }

  const isWindows = goos === 'windows';
  return {
    goos,
    goarch,
    ext: isWindows ? 'zip' : 'tar.gz',
    binName: isWindows ? 'speckeep.exe' : 'speckeep',
  };
}

function get(url, redirectsLeft) {
  return new Promise((resolve, reject) => {
    https
      .get(url, { headers: { 'User-Agent': 'speckeep-npm-installer' } }, (res) => {
        const { statusCode, headers } = res;
        if (statusCode >= 300 && statusCode < 400 && headers.location) {
          res.resume();
          if (redirectsLeft <= 0) {
            reject(new Error(`too many redirects fetching ${url}`));
            return;
          }
          resolve(get(headers.location, redirectsLeft - 1));
          return;
        }
        if (statusCode !== 200) {
          res.resume();
          reject(new Error(`GET ${url} -> HTTP ${statusCode}`));
          return;
        }
        const chunks = [];
        res.on('data', (c) => chunks.push(c));
        res.on('end', () => resolve(Buffer.concat(chunks)));
        res.on('error', reject);
      })
      .on('error', reject);
  });
}

function fetchBuffer(url) {
  return get(url, MAX_REDIRECTS);
}

function sha256Hex(buf) {
  return crypto.createHash('sha256').update(buf).digest('hex');
}

function expectedChecksum(sumsText, assetName) {
  const line = sumsText
    .split('\n')
    .map((l) => l.trim())
    .find((l) => l.endsWith(assetName));
  if (!line) return null;
  return line.split(/\s+/)[0].toLowerCase();
}

function extractArchive(archivePath, destDir) {
  fs.mkdirSync(destDir, { recursive: true });
  // Windows 10+ ships bsdtar (tar.exe) which handles both .zip and .tar.gz,
  // so a single `tar` invocation works cross-platform without extra deps.
  const result = spawnSync('tar', ['-xf', archivePath, '-C', destDir], {
    stdio: 'inherit',
  });
  if (result.error || result.status !== 0) {
    fail(
      `failed to extract ${archivePath} with tar. ` +
        'Install a tar-compatible tool (bsdtar/GNU tar) or install speckeep manually: ' +
        'https://github.com/bzdvdn/speckeep#install'
    );
  }
}

async function main() {
  if (process.env.SPECKEEP_SKIP_DOWNLOAD) {
    log('SPECKEEP_SKIP_DOWNLOAD set, skipping binary download.');
    return;
  }
  if (process.env.SPECKEEP_BINARY_PATH) {
    log('SPECKEEP_BINARY_PATH set, skipping download (bin/speckeep.js will exec it directly).');
    return;
  }

  const tag = resolveTag();
  const { goos, goarch, ext, binName } = resolvePlatform();
  const asset = `speckeep_${tag}_${goos}_${goarch}.${ext}`;
  const releaseBase = `https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${tag}`;
  const assetUrl = `${releaseBase}/${asset}`;
  const sumsUrl = `${releaseBase}/sha256sum.txt`;

  const destBinary = path.join(BIN_DIR, binName);
  if (fs.existsSync(destBinary)) {
    log(`already present: ${destBinary}`);
    return;
  }

  log(`downloading ${asset} (${tag})...`);
  let archiveBuf;
  try {
    archiveBuf = await fetchBuffer(assetUrl);
  } catch (err) {
    fail(`download failed: ${err.message}\nURL: ${assetUrl}`);
    return;
  }

  try {
    const sumsText = (await fetchBuffer(sumsUrl)).toString('utf8');
    const expected = expectedChecksum(sumsText, asset);
    if (expected) {
      const actual = sha256Hex(archiveBuf);
      if (actual !== expected) {
        fail(`checksum mismatch for ${asset}: expected ${expected}, got ${actual}`);
        return;
      }
      log('checksum verified.');
    } else {
      log(`warning: no checksum entry found for ${asset} in sha256sum.txt, skipping verification.`);
    }
  } catch (err) {
    log(`warning: could not fetch/verify sha256sum.txt (${err.message}); continuing unverified.`);
  }

  const tmpDir = fs.mkdtempSync(path.join(require('os').tmpdir(), 'speckeep-'));
  const archivePath = path.join(tmpDir, asset);
  fs.writeFileSync(archivePath, archiveBuf);

  extractArchive(archivePath, BIN_DIR);
  fs.rmSync(tmpDir, { recursive: true, force: true });

  if (!fs.existsSync(destBinary)) {
    fail(`archive did not contain expected binary: ${binName}`);
    return;
  }
  if (goos !== 'windows') {
    fs.chmodSync(destBinary, 0o755);
  }

  log(`installed: ${destBinary}`);
}

main().catch((err) => fail(err.stack || String(err)));

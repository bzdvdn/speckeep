#!/usr/bin/env node
'use strict';

// Postinstall: downloads the speckeep release binary matching this npm
// package's version + the current OS/arch from GitHub Releases, verifies its
// sha256 against the release's sha256sum.txt, and extracts it into
// ./.bin/ next to this script. bin/speckeep.js execs that binary at runtime.
//
// This package ships no compiled code of its own — speckeep is a Go binary,
// this is just a launcher so `npx speckeep` / `npm install -g speckeep` work.
//
// Archive extraction is implemented in pure Node (zlib + a small tar/zip
// reader) so it works on Windows without relying on an external `tar`/`unzip`.

const fs = require('fs');
const path = require('path');
const https = require('https');
const crypto = require('crypto');
const zlib = require('zlib');

const REPO_OWNER = 'bzdvdn';
const REPO_NAME = 'speckeep';
const PKG_ROOT = path.join(__dirname, '..');
const BIN_DIR = path.join(PKG_ROOT, '.bin');
const MAX_REDIRECTS = 5;
const DOWNLOAD_ATTEMPTS = 3;
const SOCKET_TIMEOUT_MS = 60000;

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
  // `speckeepVersion` lets the npm package version move independently of the
  // release the native binary is downloaded from (e.g. a launcher-only fix).
  const target = pkg.speckeepVersion || pkg.version;
  return String(target).startsWith('v') ? String(target) : `v${target}`;
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
    const req = https.get(url, { headers: { 'User-Agent': 'speckeep-npm-installer' } }, (res) => {
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
    });
    req.on('error', reject);
    req.setTimeout(SOCKET_TIMEOUT_MS, () => {
      req.destroy(new Error(`timeout after ${SOCKET_TIMEOUT_MS}ms fetching ${url}`));
    });
  });
}

function fetchBuffer(url) {
  return get(url, MAX_REDIRECTS);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

// fetchBufferWithRetry retries transient network failures (ECONNRESET, socket
// hangs, timeouts) — important on Windows and flaky connections.
async function fetchBufferWithRetry(url, attempts) {
  let lastErr;
  for (let attempt = 1; attempt <= attempts; attempt++) {
    try {
      return await fetchBuffer(url);
    } catch (err) {
      lastErr = err;
      if (attempt < attempts) {
        log(`download attempt ${attempt}/${attempts} failed (${err.message}); retrying...`);
        await sleep(500 * attempt);
      }
    }
  }
  throw lastErr;
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

// extractFromTarGz reads a .tar.gz buffer and returns the bytes of `binName`.
function extractFromTarGz(buf, binName) {
  const tar = zlib.gunzipSync(buf);
  let off = 0;
  while (off + 512 <= tar.length) {
    const name = tar.toString('utf8', off, off + 100).replace(/\0.*$/, '');
    if (!name) break; // zero block: end of archive
    const sizeField = tar
      .toString('utf8', off + 124, off + 136)
      .replace(/\0.*$/, '')
      .trim();
    const size = sizeField ? parseInt(sizeField, 8) : 0;
    const type = tar[off + 156];
    const base = name.split('/').pop();
    const dataStart = off + 512;
    if ((type === 0 || type === 0x30) && base === binName) {
      return tar.slice(dataStart, dataStart + size);
    }
    off = dataStart + Math.ceil(size / 512) * 512;
  }
  throw new Error(`${binName} not found in tar.gz archive`);
}

// extractFromZip reads a .zip buffer and returns the bytes of `binName`.
// Handles stored (0) and deflate (8) entries; the release zips use deflate.
function extractFromZip(buf, binName) {
  const EOCD_SIG = 0x06054b50;
  const CDH_SIG = 0x02014b50;
  const LFH_SIG = 0x04034b50;

  let eocd = -1;
  const minEocd = Math.max(0, buf.length - 22 - 0xffff);
  for (let i = buf.length - 22; i >= minEocd; i--) {
    if (buf.readUInt32LE(i) === EOCD_SIG) {
      eocd = i;
      break;
    }
  }
  if (eocd < 0) throw new Error('zip: end-of-central-directory not found');

  const entries = buf.readUInt16LE(eocd + 10);
  let p = buf.readUInt32LE(eocd + 16);
  for (let n = 0; n < entries; n++) {
    if (buf.readUInt32LE(p) !== CDH_SIG) throw new Error('zip: bad central directory header');
    const method = buf.readUInt16LE(p + 10);
    const compSize = buf.readUInt32LE(p + 20);
    const nameLen = buf.readUInt16LE(p + 28);
    const extraLen = buf.readUInt16LE(p + 30);
    const commentLen = buf.readUInt16LE(p + 32);
    const localOffset = buf.readUInt32LE(p + 42);
    const name = buf.toString('utf8', p + 46, p + 46 + nameLen);

    if (name.split('/').pop() === binName) {
      if (buf.readUInt32LE(localOffset) !== LFH_SIG) throw new Error('zip: bad local file header');
      const lNameLen = buf.readUInt16LE(localOffset + 26);
      const lExtraLen = buf.readUInt16LE(localOffset + 28);
      const dataStart = localOffset + 30 + lNameLen + lExtraLen;
      const data = buf.slice(dataStart, dataStart + compSize);
      if (method === 0) return data;
      if (method === 8) return zlib.inflateRawSync(data);
      throw new Error(`zip: unsupported compression method ${method}`);
    }
    p += 46 + nameLen + extraLen + commentLen;
  }
  throw new Error(`${binName} not found in zip archive`);
}

function extractBinary(archiveBuf, ext, binName) {
  if (ext === 'tar.gz') return extractFromTarGz(archiveBuf, binName);
  if (ext === 'zip') return extractFromZip(archiveBuf, binName);
  throw new Error(`unsupported archive format: ${ext}`);
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
    archiveBuf = await fetchBufferWithRetry(assetUrl, DOWNLOAD_ATTEMPTS);
  } catch (err) {
    fail(`download failed: ${err.message}\nURL: ${assetUrl}`);
    return;
  }

  try {
    const sumsText = (await fetchBufferWithRetry(sumsUrl, 2)).toString('utf8');
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

  let binary;
  try {
    binary = extractBinary(archiveBuf, ext, binName);
  } catch (err) {
    fail(`failed to extract ${asset}: ${err.message}`);
    return;
  }
  if (!binary || binary.length === 0) {
    fail(`archive ${asset} contained an empty ${binName}`);
    return;
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });
  fs.writeFileSync(destBinary, binary);
  if (goos !== 'windows') {
    fs.chmodSync(destBinary, 0o755);
  }

  log(`installed: ${destBinary}`);
}

// Exported for the packaging smoke test; no-op when required as a dependency.
module.exports = { extractBinary, extractFromTarGz, extractFromZip };

if (require.main === module) {
  main().catch((err) => fail(err.stack || String(err)));
}

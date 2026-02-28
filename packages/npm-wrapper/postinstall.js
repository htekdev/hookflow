#!/usr/bin/env node

const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { execSync } = require('child_process');

const REPO = 'htekdev/gh-hookflow';
const BIN_DIR = path.join(__dirname, 'bin');

// Determine platform and architecture
function getPlatformInfo() {
  const platform = os.platform();
  const arch = os.arch();

  let osName;
  switch (platform) {
    case 'win32':
      osName = 'windows';
      break;
    case 'darwin':
      osName = 'darwin';
      break;
    case 'linux':
      osName = 'linux';
      break;
    default:
      throw new Error(`Unsupported platform: ${platform}`);
  }

  let archName;
  switch (arch) {
    case 'x64':
    case 'amd64':
      archName = 'amd64';
      break;
    case 'arm64':
    case 'aarch64':
      archName = 'arm64';
      break;
    default:
      throw new Error(`Unsupported architecture: ${arch}`);
  }

  const ext = platform === 'win32' ? '.exe' : '';
  return {
    osName,
    archName,
    binaryName: `hookflow-${osName}-${archName}${ext}`,
    ext
  };
}

// Get latest release info from GitHub
function getLatestRelease() {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: `/repos/${REPO}/releases/latest`,
      headers: {
        'User-Agent': 'hookflow-npm-installer'
      }
    };

    https.get(options, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        if (res.statusCode === 200) {
          resolve(JSON.parse(data));
        } else if (res.statusCode === 404) {
          // No releases yet, try tags
          reject(new Error('No releases found. The CLI may not be published yet.'));
        } else {
          reject(new Error(`GitHub API error: ${res.statusCode}`));
        }
      });
    }).on('error', reject);
  });
}

// Download a file from URL
function downloadFile(url, destPath) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destPath);
    
    function followRedirects(url) {
      https.get(url, (res) => {
        if (res.statusCode === 302 || res.statusCode === 301) {
          followRedirects(res.headers.location);
          return;
        }
        
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: ${res.statusCode}`));
          return;
        }

        res.pipe(file);
        file.on('finish', () => {
          file.close();
          resolve();
        });
      }).on('error', (err) => {
        fs.unlink(destPath, () => {});
        reject(err);
      });
    }

    followRedirects(url);
  });
}

// Main installation function
async function install() {
  console.log('Installing hookflow CLI...');

  try {
    const { osName, archName, binaryName, ext } = getPlatformInfo();
    console.log(`Platform: ${osName}-${archName}`);

    // Ensure bin directory exists
    if (!fs.existsSync(BIN_DIR)) {
      fs.mkdirSync(BIN_DIR, { recursive: true });
    }

    const destPath = path.join(BIN_DIR, binaryName);

    // Try to get the latest release
    let downloadUrl;
    try {
      const release = await getLatestRelease();
      console.log(`Latest release: ${release.tag_name}`);

      // Find the matching asset
      const asset = release.assets.find(a => a.name === binaryName);
      if (!asset) {
        throw new Error(`Binary not found in release: ${binaryName}`);
      }
      downloadUrl = asset.browser_download_url;
    } catch (err) {
      console.log(`Note: ${err.message}`);
      console.log('Attempting direct download from releases...');
      
      // Try direct download from latest release (fallback)
      downloadUrl = `https://github.com/${REPO}/releases/latest/download/${binaryName}`;
    }

    console.log(`Downloading from: ${downloadUrl}`);
    await downloadFile(downloadUrl, destPath);

    // Make executable on Unix
    if (os.platform() !== 'win32') {
      fs.chmodSync(destPath, 0o755);
    }

    console.log(`✓ Installed hookflow to ${destPath}`);

    // Verify installation
    try {
      const version = execSync(`"${destPath}" version`, { encoding: 'utf8' }).trim();
      console.log(`✓ ${version}`);
    } catch (e) {
      // Version check failed, but binary exists
      console.log('✓ Binary installed (version check skipped)');
    }

  } catch (err) {
    console.error(`Installation failed: ${err.message}`);
    console.error('\nYou can also install manually:');
    console.error('  1. go install github.com/htekdev/gh-hookflow/cmd/hookflow@latest');
    console.error('  2. Or download from: https://github.com/htekdev/gh-hookflow/releases');
    process.exit(1);
  }
}

install();

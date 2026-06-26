const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const versionFile = path.join(__dirname, '../VERSION');
const version = fs.readFileSync(versionFile, 'utf8').trim();

console.log(`Updating project to version: ${version}`);

// Function to update package.json
const updatePackageJson = (filePath) => {
  if (fs.existsSync(filePath)) {
    const pkg = JSON.parse(fs.readFileSync(filePath, 'utf8'));
    pkg.version = version;
    fs.writeFileSync(filePath, JSON.stringify(pkg, null, 2) + '\n');
    console.log(`Updated ${filePath}`);
  }
};

// Function to update manifest.json
const updateManifest = (filePath) => {
  if (fs.existsSync(filePath)) {
    // Read manifest.ts usually, but here we might need to handle build artifacts or source
    // For now, let's assume we update package.json which drives the build
  }
};

// Find all package.json files
const findPackageJsons = (dir, fileList = []) => {
  const files = fs.readdirSync(dir);
  files.forEach(file => {
    const filePath = path.join(dir, file);
    const stat = fs.statSync(filePath);
    if (stat.isDirectory()) {
      if (file !== 'node_modules' && file !== '.git') {
        findPackageJsons(filePath, fileList);
      }
    } else {
      if (file === 'package.json') {
        fileList.push(filePath);
      }
    }
  });
  return fileList;
};

const packageJsons = findPackageJsons(path.join(__dirname, '..'));
packageJsons.forEach(updatePackageJson);

console.log('Version update complete.');

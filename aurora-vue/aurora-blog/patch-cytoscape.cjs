// patch-cytoscape.js
// This script patches cytoscape's package.json to add the missing "import" condition
// for the UMD bundle, which is required by Vite 8 (Rolldown).

const fs = require('fs');
const path = require('path');

const cytoscapePackageJson = path.join(__dirname, 'node_modules', 'cytoscape', 'package.json');

if (!fs.existsSync(cytoscapePackageJson)) {
  console.log('cytoscape not found, skipping patch');
  process.exit(0);
}

const pkg = JSON.parse(fs.readFileSync(cytoscapePackageJson, 'utf8'));

if (pkg.exports && pkg.exports['./dist/cytoscape.umd.js']) {
  const umdExport = pkg.exports['./dist/cytoscape.umd.js'];
  if (!umdExport.import) {
    umdExport.import = './dist/cytoscape.umd.js';
    fs.writeFileSync(cytoscapePackageJson, JSON.stringify(pkg, null, 2));
    console.log('✓ Patched cytoscape package.json (added import condition for UMD bundle)');
  } else {
    console.log('✓ cytoscape already patched');
  }
} else {
  console.log('✗ cytoscape exports field not found or UMD export missing');
}

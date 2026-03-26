import { cp, mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const packageDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const distDir = path.join(packageDir, 'dist');

const cssEntries = [
  'src/design-tokens/default.css',
  'src/design-tokens/dark.css',
  'src/css/global.css',
  'src/css/transition.css',
  'src/css/nprogress.css',
  'src/css/ui.css',
];

async function readCssBundle() {
  const parts = await Promise.all(
    cssEntries.map(async (relativePath) =>
      readFile(path.join(packageDir, relativePath), 'utf8'),
    ),
  );

  return parts.join('\n\n');
}

async function main() {
  await rm(distDir, { force: true, recursive: true });
  await mkdir(distDir, { recursive: true });

  const cssBundle = await readCssBundle();
  await writeFile(path.join(distDir, 'design.css'), cssBundle, 'utf8');
  await writeFile(path.join(distDir, 'index.mjs'), "import './design.css';\n", 'utf8');

  await cp(path.join(packageDir, 'src/css'), distDir, { recursive: true });

  await cp(
    path.join(packageDir, 'src/scss-bem/bem.scss'),
    path.join(distDir, 'bem.scss'),
  );
  await cp(
    path.join(packageDir, 'src/scss-bem/constants.scss'),
    path.join(distDir, 'constants.scss'),
  );
}

await main();

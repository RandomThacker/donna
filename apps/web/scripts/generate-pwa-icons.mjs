import { mkdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const iconsDir = path.join(root, "public", "icons");

const outputs = [
  { input: "donna-icon.svg", output: "icon-192.png", size: 192 },
  { input: "donna-icon.svg", output: "icon-512.png", size: 512 },
  { input: "donna-icon-maskable.svg", output: "icon-maskable-512.png", size: 512 },
  { input: "donna-icon.svg", output: "apple-touch-icon.png", size: 180 },
];

await mkdir(iconsDir, { recursive: true });

for (const item of outputs) {
  const inputPath = path.join(iconsDir, item.input);
  const outputPath = path.join(iconsDir, item.output);
  const svg = await readFile(inputPath);
  await sharp(svg).resize(item.size, item.size).png().toFile(outputPath);
  console.log(`wrote ${item.output}`);
}

console.log("PWA icons generated.");

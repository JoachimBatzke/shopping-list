/**
 * Create PWA icons from source image
 * Run with: node scripts/create-pwa-icons.js
 */

const fs = require('fs')
const path = require('path')

const SOURCE_IMAGE = String.raw`C:\Users\Joachim\Documents\List PWA\Update 08-12-2025 Categories and highlighted links\Backup jorlist\WFXT3ucWzqRKX2BmWI93s3KAxbRvQQSH\og-image-300x300.png`
const PUBLIC_DIR = path.join(__dirname, '..', 'public')

const ICONS = [
  'pwa-192x192.png',
  'pwa-512x512.png',
  'apple-touch-icon.png'
]

async function main() {
  // Check if source exists
  if (!fs.existsSync(SOURCE_IMAGE)) {
    console.error('Source image not found:', SOURCE_IMAGE)
    process.exit(1)
  }

  // Try to use sharp for resizing
  let sharp
  try {
    sharp = require('sharp')
  } catch (e) {
    sharp = null
  }

  if (sharp) {
    // Use sharp for proper resizing
    console.log('Using sharp for image resizing...')

    await sharp(SOURCE_IMAGE)
      .resize(192, 192)
      .toFile(path.join(PUBLIC_DIR, 'pwa-192x192.png'))
    console.log('Created: pwa-192x192.png')

    await sharp(SOURCE_IMAGE)
      .resize(512, 512)
      .toFile(path.join(PUBLIC_DIR, 'pwa-512x512.png'))
    console.log('Created: pwa-512x512.png')

    await sharp(SOURCE_IMAGE)
      .resize(180, 180)
      .toFile(path.join(PUBLIC_DIR, 'apple-touch-icon.png'))
    console.log('Created: apple-touch-icon.png')

  } else {
    // Fallback: just copy the original (300x300 is close enough for most uses)
    console.log('sharp not installed, copying original image...')
    console.log('For proper resizing, run: npm install sharp')
    console.log('')

    const sourceData = fs.readFileSync(SOURCE_IMAGE)
    ICONS.forEach(name => {
      fs.writeFileSync(path.join(PUBLIC_DIR, name), sourceData)
      console.log(`Created: ${name} (copy of original 300x300)`)
    })
  }

  console.log('')
  console.log('Done!')
}

main().catch(console.error)

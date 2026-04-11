const { copyFileSync, existsSync, mkdirSync, readFileSync } = require('node:fs')
const { join } = require('node:path')
const { spawnSync } = require('node:child_process')

const run = (command, args, options = {}) => {
  const result = spawnSync(command, args, {
    stdio: 'inherit',
    shell: process.platform === 'win32',
    ...options
  })

  if (result.status !== 0) {
    process.exit(result.status || 1)
  }
}

if (!existsSync(join(process.cwd(), 'android'))) {
  console.error('[Android] Missing android project. Run `pnpm android:add` first.')
  process.exit(1)
}

const androidApiOrigin =
  process.env.CHEM_ANDROID_API_ORIGIN ||
  process.env.CHEM_SERVER_ORIGIN ||
  process.env.VITE_SERVER_ORIGIN ||
  process.env.VITE_API_ORIGIN ||
  ''
if (!androidApiOrigin) {
  console.error('[Android] Missing CHEM_ANDROID_API_ORIGIN. Release builds must target a real backend origin.')
  process.exit(1)
}

const packageJson = JSON.parse(readFileSync(join(process.cwd(), 'package.json'), 'utf-8'))
const hasSigning = [
  process.env.CHEM_ANDROID_KEYSTORE_PATH,
  process.env.CHEM_ANDROID_KEYSTORE_PASSWORD,
  process.env.CHEM_ANDROID_KEY_ALIAS,
  process.env.CHEM_ANDROID_KEY_PASSWORD
].every((value) => typeof value === 'string' && value.trim().length > 0)

console.log(`[Android] Using API origin: ${androidApiOrigin}`)
console.log(`[Android] Release signing: ${hasSigning ? 'enabled' : 'disabled (unsigned APK)'}`)

run('node', ['scripts/android-sync.cjs'], {
  env: {
    ...process.env,
    CHEM_ANDROID_API_ORIGIN: androidApiOrigin,
    CHEM_SERVER_ORIGIN: androidApiOrigin,
    VITE_SERVER_ORIGIN: androidApiOrigin,
    VITE_API_ORIGIN: androidApiOrigin
  }
})

const gradleWrapper = process.platform === 'win32' ? 'gradlew.bat' : './gradlew'
run(gradleWrapper, ['assembleRelease'], { cwd: join(process.cwd(), 'android') })

const releaseDir = join(process.cwd(), 'release', 'android')
const sourceApk = join(
  process.cwd(),
  'android',
  'app',
  'build',
  'outputs',
  'apk',
  'release',
  hasSigning ? 'app-release.apk' : 'app-release-unsigned.apk'
)
const targetApk = join(
  releaseDir,
  `ChemistryUNO-${packageJson.version}-android-release${hasSigning ? '' : '-unsigned'}.apk`
)

if (!existsSync(sourceApk)) {
  console.error(`[Android] Release APK not found: ${sourceApk}`)
  process.exit(1)
}

mkdirSync(releaseDir, { recursive: true })
copyFileSync(sourceApk, targetApk)

console.log(`[Android] Release APK build completed: ${sourceApk}`)
console.log(`[Android] Copied release APK to: ${targetApk}`)

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

const packageJson = JSON.parse(readFileSync(join(process.cwd(), 'package.json'), 'utf-8'))
const releaseDir = join(process.cwd(), 'release', 'android')

run('node', ['scripts/android-sync.cjs'])

const gradleWrapper = process.platform === 'win32' ? 'gradlew.bat' : './gradlew'
run(gradleWrapper, ['assembleDebug'], { cwd: join(process.cwd(), 'android') })

const sourceApk = join(process.cwd(), 'android', 'app', 'build', 'outputs', 'apk', 'debug', 'app-debug.apk')
const targetApk = join(releaseDir, `ChemistryUNO-${packageJson.version}-android-debug.apk`)

mkdirSync(releaseDir, { recursive: true })
copyFileSync(sourceApk, targetApk)

console.log(`[Android] Debug APK build completed: ${sourceApk}`)
console.log(`[Android] Copied debug APK to: ${targetApk}`)

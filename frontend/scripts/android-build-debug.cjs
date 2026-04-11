const { existsSync } = require('node:fs')
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

run('node', ['scripts/android-sync.cjs'])

const gradleWrapper = process.platform === 'win32' ? 'gradlew.bat' : './gradlew'
run(gradleWrapper, ['assembleDebug'], { cwd: join(process.cwd(), 'android') })

console.log('[Android] Debug APK build completed: android/app/build/outputs/apk/debug/app-debug.apk')

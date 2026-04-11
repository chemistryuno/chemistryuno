const { spawnSync } = require('node:child_process')

const androidApiOrigin =
  process.env.CHEM_ANDROID_API_ORIGIN ||
  process.env.CHEM_SERVER_ORIGIN ||
  process.env.VITE_SERVER_ORIGIN ||
  'http://10.0.2.2:8080'

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

console.log(`[Android] Using API origin: ${androidApiOrigin}`)

run('pnpm', ['build'], {
  env: {
    ...process.env,
    VITE_SERVER_ORIGIN: androidApiOrigin,
    VITE_API_ORIGIN: androidApiOrigin
  }
})

run('pnpm', ['cap', 'sync', 'android'])
run('node', ['scripts/android-prepare-project.cjs'])

console.log('[Android] Web assets built and synced to android project.')

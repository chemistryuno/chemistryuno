const { existsSync, readFileSync, writeFileSync } = require('node:fs')
const { join } = require('node:path')

const appBuildGradlePath = join(process.cwd(), 'android', 'app', 'build.gradle')

if (!existsSync(appBuildGradlePath)) {
  console.error('[Android] Missing android/app/build.gradle. Run `pnpm android:add` first.')
  process.exit(1)
}

let buildGradle = readFileSync(appBuildGradlePath, 'utf-8')

if (buildGradle.includes('computedVersionCode') && buildGradle.includes('hasReleaseSigning')) {
  console.log('[Android] app/build.gradle already prepared.')
  process.exit(0)
}

const injectedPrelude = `apply plugin: 'com.android.application'

import groovy.json.JsonSlurper

def frontendPackage = new JsonSlurper().parse(file('../../package.json'))
def frontendVersionName = (frontendPackage.version ?: '1.0.0').toString()
def versionParts = frontendVersionName.tokenize('.').collect { part ->
    part.isInteger() ? part.toInteger() : 0
}
while (versionParts.size() < 3) {
    versionParts << 0
}
def computedVersionCode = (versionParts[0] * 10000) + (versionParts[1] * 100) + versionParts[2]

def releaseStoreFilePath = System.getenv('CHEM_ANDROID_KEYSTORE_PATH')
def releaseStorePassword = System.getenv('CHEM_ANDROID_KEYSTORE_PASSWORD')
def releaseKeyAlias = System.getenv('CHEM_ANDROID_KEY_ALIAS')
def releaseKeyPassword = System.getenv('CHEM_ANDROID_KEY_PASSWORD')
def hasReleaseSigning = [releaseStoreFilePath, releaseStorePassword, releaseKeyAlias, releaseKeyPassword].every {
    it != null && it.trim()
}`

if (!buildGradle.includes("import groovy.json.JsonSlurper")) {
  buildGradle = buildGradle.replace(
    "apply plugin: 'com.android.application'",
    injectedPrelude
  )
}

buildGradle = buildGradle.replace(/versionCode\s+\d+/, 'versionCode computedVersionCode')
buildGradle = buildGradle.replace(/versionName\s+"[^"]+"/, 'versionName frontendVersionName')

if (!buildGradle.includes('signingConfigs {')) {
  buildGradle = buildGradle.replace(
    '    buildTypes {',
    `    signingConfigs {
        release {
            if (hasReleaseSigning) {
                storeFile file(releaseStoreFilePath)
                storePassword releaseStorePassword
                keyAlias releaseKeyAlias
                keyPassword releaseKeyPassword
            }
        }
    }
    buildTypes {`
  )
}

if (!buildGradle.includes('signingConfig signingConfigs.release')) {
  buildGradle = buildGradle.replace(
    '            minifyEnabled false',
    `            minifyEnabled false
            if (hasReleaseSigning) {
                signingConfig signingConfigs.release
            }`
  )
}

writeFileSync(appBuildGradlePath, buildGradle)
console.log('[Android] Prepared android/app/build.gradle for versioning and optional release signing.')

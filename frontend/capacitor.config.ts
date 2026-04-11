import type { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'com.chemistryuno.client',
  appName: 'Chemistry UNO',
  webDir: 'dist',
  bundledWebRuntime: false,
  server: {
    cleartext: true
  },
  android: {
    allowMixedContent: true
  }
}

export default config

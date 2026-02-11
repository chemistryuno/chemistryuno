import type { CapacitorConfig } from '@capacitor/cli'

const config: CapacitorConfig = {
  appId: 'com.chemistryuno.app',
  appName: 'ChemistryUNO',
  webDir: 'dist',
  server: {
    // 移动端连接的后端服务器地址，打包前请修改为你的实际服务器地址
    url: 'https://cu.tomsite.us.kg',
    cleartext: true
  },
  android: {
    buildOptions: {
      keystorePath: undefined,
      keystoreAlias: undefined,
      keystorePassword: undefined,
      keystoreAliasPassword: undefined,
      releaseType: 'APK'
    }
  },
  ios: {
    scheme: 'ChemistryUNO'
  }
}

export default config

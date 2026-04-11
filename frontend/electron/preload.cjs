const { contextBridge } = require('electron')

const apiOrigin = process.env.CHEM_CLIENT_API_ORIGIN || 'http://127.0.0.1:8080'

contextBridge.exposeInMainWorld('__CHEM_RUNTIME_CONFIG', {
  apiOrigin
})

const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

const normalizeApiOrigin = (raw: string): string => {
  const value = trimTrailingSlash(String(raw || '').trim())
  if (!value) return ''

  if (value.endsWith('/api')) {
    return value.slice(0, -4)
  }

  return value
}

type RuntimeConfig = {
  apiOrigin?: string
}

const readRuntimeConfig = (): RuntimeConfig => {
  const win = window as Window & { __CHEM_RUNTIME_CONFIG?: RuntimeConfig }
  return win.__CHEM_RUNTIME_CONFIG || {}
}

const resolveApiOrigin = (): string => {
  const envOrigin = normalizeApiOrigin(
    String(
      import.meta.env.VITE_SERVER_ORIGIN ||
      import.meta.env.VITE_API_ORIGIN ||
      import.meta.env.VITE_API_BASE_URL ||
      ''
    )
  )
  if (envOrigin) {
    return envOrigin
  }

  return normalizeApiOrigin(String(readRuntimeConfig().apiOrigin || ''))
}

export const API_ORIGIN = resolveApiOrigin()
export const API_BASE_URL = API_ORIGIN ? `${API_ORIGIN}/api` : '/api'

export const buildApiURL = (path: string): string => {
  if (/^https?:\/\//i.test(path)) {
    return path
  }

  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  if (API_BASE_URL === '/api') {
    return `${API_BASE_URL}${normalizedPath}`
  }

  return `${trimTrailingSlash(API_BASE_URL)}${normalizedPath}`
}

const toWebSocketOrigin = (origin: string): string => {
  return origin.replace(/^http:/i, 'ws:').replace(/^https:/i, 'wss:')
}

export const WS_URL = API_ORIGIN
  ? `${toWebSocketOrigin(API_ORIGIN)}/api/ws`
  : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/ws`

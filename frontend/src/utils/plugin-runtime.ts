import { pluginAPI } from './api'

type PluginMeta = {
  id: number
  name?: string
  version?: string
  has_script?: boolean
  has_client_script?: boolean
}

const loadedPluginIds = new Set<number>()
const messageHandlers = new Map<number, Array<(payload: any, meta: any) => void>>()

function createPluginApi() {
  const baseURL = '/api'
  const getAuthHeaders = (): Record<string, string> => {
    const token = localStorage.getItem('token')
    if (!token) return {}
    return { Authorization: `Bearer ${token}` }
  }

  const request = async (method: string, path: string, data?: any, extraHeaders?: Record<string, string>) => {
    const url = path.startsWith('http') ? path : `${baseURL}${path.startsWith('/') ? '' : '/'}${path}`
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...getAuthHeaders(),
      ...(extraHeaders || {})
    }
    const res = await fetch(url, {
      method,
      headers,
      body: data === undefined ? undefined : JSON.stringify(data)
    })

    const contentType = res.headers.get('content-type') || ''
    const isJson = contentType.includes('application/json')
    const payload = isJson ? await res.json() : await res.text()

    if (!res.ok) {
      const message = isJson && payload?.error ? payload.error : `HTTP ${res.status}`
      throw new Error(message)
    }

    return payload
  }

  return {
    request,
    get: (path: string) => request('GET', path),
    post: (path: string, data?: any) => request('POST', path, data),
    put: (path: string, data?: any) => request('PUT', path, data),
    del: (path: string, data?: any) => request('DELETE', path, data),
  }
}

function runPluginScript(plugin: PluginMeta, code: string) {
  const api = createPluginApi()
  const context = {
    plugin,
    api,
    console,
    storage: localStorage,
    onMessage: (handler: (payload: any, meta: any) => void) => {
      if (!messageHandlers.has(plugin.id)) messageHandlers.set(plugin.id, [])
      messageHandlers.get(plugin.id)!.push(handler)
    }
  }

  try {
    const fn = new Function('plugin', 'api', 'context', code)
    fn(plugin, api, context)
  } catch (err) {
    console.error(`[Plugin:${plugin.name || plugin.id}] script error:`, err)
  }
}

export async function loadPluginScripts() {
  try {
    const res = await pluginAPI.getPluginsWithCards()
    const plugins: PluginMeta[] = res.data || []

    for (const plugin of plugins) {
      const hasClient = plugin.has_client_script ?? plugin.has_script
      if (!hasClient) continue
      if (loadedPluginIds.has(plugin.id)) continue

      try {
        const scriptRes = await pluginAPI.getPluginScript(plugin.id)
        const code = typeof scriptRes.data === 'string' ? scriptRes.data : ''
        if (!code.trim()) continue

        runPluginScript(plugin, code)
        loadedPluginIds.add(plugin.id)
      } catch (err) {
        console.warn(`[Plugin:${plugin.name || plugin.id}] failed to load script`, err)
      }
    }
  } catch (err) {
    console.warn('[Plugin] failed to load plugins', err)
  }
}

export function dispatchPluginMessage(message: any) {
  const data = message?.data ?? message
  const pluginId = data?.plugin_id
  const payload = data?.payload
  const meta = {
    plugin_id: pluginId,
    scope: data?.scope,
    room_id: data?.room_id,
    uid: data?.uid,
    timestamp: data?.timestamp
  }

  if (typeof pluginId === 'number' && messageHandlers.has(pluginId)) {
    for (const handler of messageHandlers.get(pluginId) || []) {
      try {
        handler(payload, meta)
      } catch (err) {
        console.error(`[Plugin:${pluginId}] message handler error:`, err)
      }
    }
    return
  }

  for (const handlers of messageHandlers.values()) {
    for (const handler of handlers) {
      try {
        handler(payload, meta)
      } catch (err) {
        console.error('[Plugin] message handler error:', err)
      }
    }
  }
}

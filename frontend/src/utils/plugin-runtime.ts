import { defineComponent, h, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import router from '../router'
import { pluginAPI } from './api'

type PluginMeta = {
  id: number
  name?: string
  version?: string
  has_script?: boolean
  has_client_script?: boolean
  config_schema?: string
}

type PluginRouteDefinition = {
  path: string
  name?: string
  title?: string
  description?: string
  requiresAuth?: boolean
  adminOnly?: boolean
  coWorkerOnly?: boolean
  render?: (ctx: any) => string | HTMLElement | void
  onMount?: (ctx: any) => void
  onUnmount?: (ctx: any) => void
}

type PluginRouteRecord = {
  path: string
  name: string
  source: 'script' | 'schema'
  title?: string
}

type PluginConfigField = {
  key: string
  type: string
}

type PluginRoutePage = {
  path: string
  title?: string
  description?: string
  content_html?: string
  requires_auth?: boolean
  admin_only?: boolean
  co_worker_only?: boolean
}

const loadedPluginIds = new Set<number>()
const loadedPluginMeta = new Map<number, PluginMeta>()
const messageHandlers = new Map<number, Array<(payload: any, meta: any) => void>>()
const pluginRoutes = new Map<number, Map<string, PluginRouteRecord>>() // key=path

function ensurePluginRouteMap(pluginID: number) {
  if (!pluginRoutes.has(pluginID)) {
    pluginRoutes.set(pluginID, new Map<string, PluginRouteRecord>())
  }
  return pluginRoutes.get(pluginID)!
}

function normalizeRoutePath(path: string): string {
  const raw = String(path || '').trim()
  if (!raw.startsWith('/')) throw new Error('route.path 必须以 / 开头')
  if (raw.length > 160) throw new Error('route.path 过长')
  return raw
}

function sanitizeRouteName(path: string): string {
  return path.replace(/[^a-zA-Z0-9]/g, '_').replace(/_+/g, '_').slice(0, 48)
}

function hasRoutePath(path: string): boolean {
  return router.getRoutes().some((route) => route.path === path)
}

function parseConfigSchema(raw: string | undefined): PluginConfigField[] {
  if (!raw || !String(raw).trim()) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((item) => item && typeof item.key === 'string')
      .map((item) => ({ key: String(item.key), type: String(item.type || 'text') }))
  } catch {
    return []
  }
}

function parseRoutePages(raw: string): PluginRoutePage[] {
  if (!raw || !String(raw).trim()) return []
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return []
    return parsed
      .filter((item) => item && typeof item.path === 'string')
      .map((item) => ({
        path: String(item.path),
        title: typeof item.title === 'string' ? item.title : '',
        description: typeof item.description === 'string' ? item.description : '',
        content_html: typeof item.content_html === 'string' ? item.content_html : '',
        requires_auth: item.requires_auth !== false,
        admin_only: Boolean(item.admin_only),
        co_worker_only: Boolean(item.co_worker_only)
      }))
  } catch {
    return []
  }
}

function buildRouteComponent(plugin: PluginMeta, routeDef: PluginRouteDefinition, api: ReturnType<typeof createPluginApi>) {
  const title = routeDef.title || plugin.name || `Plugin #${plugin.id}`
  const description = routeDef.description || '由插件提供的自定义页面'

  return defineComponent({
    name: `PluginRoute_${plugin.id}_${sanitizeRouteName(routeDef.path)}`,
    setup() {
      const containerRef = ref<HTMLElement | null>(null)
      const currentRoute = useRoute()

      const buildContext = () => ({
        plugin,
        api,
        router,
        route: currentRoute,
        params: currentRoute.params,
        query: currentRoute.query,
        storage: localStorage,
        console,
        container: containerRef.value,
        navigate: (to: any) => router.push(to)
      })

      onMounted(() => {
        const ctx = buildContext()
        try {
          if (typeof routeDef.render === 'function' && ctx.container) {
            const output = routeDef.render(ctx)
            if (typeof output === 'string') {
              ctx.container.innerHTML = output
            } else if (output instanceof HTMLElement) {
              ctx.container.innerHTML = ''
              ctx.container.appendChild(output)
            }
          }
          if (typeof routeDef.onMount === 'function') {
            routeDef.onMount(ctx)
          }
        } catch (err) {
          console.error(`[Plugin:${plugin.name || plugin.id}] route mount error`, err)
        }
      })

      onBeforeUnmount(() => {
        if (typeof routeDef.onUnmount !== 'function') return
        try {
          routeDef.onUnmount(buildContext())
        } catch (err) {
          console.error(`[Plugin:${plugin.name || plugin.id}] route unmount error`, err)
        }
      })

      return () =>
        h('div', { class: 'min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-white p-4 md:p-8 selection:bg-blue-500/30' }, [
          h('div', { class: 'max-w-6xl mx-auto relative z-10' }, [
            h('div', { class: 'flex items-center justify-between mb-6 gap-3 flex-wrap' }, [
              h('div', {}, [
                h('p', { class: 'text-[10px] font-black text-slate-400 uppercase tracking-widest' }, `Plugin Route · ${plugin.name || plugin.id}`),
                h('h1', { class: 'text-2xl font-black tracking-tight text-slate-900 dark:text-white' }, title),
                h('p', { class: 'text-sm text-slate-500 dark:text-slate-400 mt-1' }, description)
              ]),
              h(
                'button',
                {
                  class: 'px-3 py-2 rounded-xl border border-slate-200 dark:border-white/10 bg-white dark:bg-white/5 text-xs font-black uppercase tracking-widest hover:bg-slate-100 dark:hover:bg-white/10 transition-colors',
                  onClick: () => router.back()
                },
                '返回'
              )
            ]),
            h('div', { class: 'bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] overflow-hidden shadow-sm' }, [
              h('div', { class: 'p-4 md:p-6' }, [h('div', { ref: containerRef, class: 'min-h-[180px]' })])
            ])
          ])
        ])
    }
  })
}

function createPluginApi() {
  const baseURL = '/api'

  const request = async (method: string, path: string, data?: any, extraHeaders?: Record<string, string>) => {
    const url = path.startsWith('http') ? path : `${baseURL}${path.startsWith('/') ? '' : '/'}${path}`
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(extraHeaders || {})
    }
    const res = await fetch(url, {
      method,
      headers,
      credentials: 'include', // 自动发送cookie中的token
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

function registerPluginRoute(
  plugin: PluginMeta,
  api: ReturnType<typeof createPluginApi>,
  routeDef: PluginRouteDefinition,
  source: 'script' | 'schema' = 'script'
) {
  const path = normalizeRoutePath(routeDef.path)
  const routeMap = ensurePluginRouteMap(plugin.id)
  if (routeMap.has(path)) {
    return routeMap.get(path)!.name
  }
  if (hasRoutePath(path)) {
    throw new Error(`路由冲突: ${path} 已存在`)
  }

  const autoName = `plugin_${plugin.id}_${sanitizeRouteName(path)}`
  const name = String(routeDef.name || autoName).trim()
  if (!name) throw new Error('route.name 无效')
  if (router.hasRoute(name)) throw new Error(`路由名称冲突: ${name}`)

  router.addRoute({
    path,
    name,
    component: buildRouteComponent(plugin, routeDef, api),
    meta: {
      requiresAuth: routeDef.requiresAuth !== false,
      adminOnly: Boolean(routeDef.adminOnly),
      coWorkerOnly: Boolean(routeDef.coWorkerOnly),
      pluginRoute: true,
      pluginId: plugin.id
    }
  })

  routeMap.set(path, {
    path,
    name,
    source,
    title: routeDef.title || ''
  })
  return name
}

function unregisterPluginRoute(plugin: PluginMeta, nameOrPath: string) {
  const routeMap = ensurePluginRouteMap(plugin.id)
  const target = String(nameOrPath || '').trim()
  if (!target) return false

  let name = target
  if (routeMap.has(target)) {
    name = routeMap.get(target)!.name
  } else {
    const found = Array.from(routeMap.entries()).find(([, routeRecord]) => routeRecord.name === target)
    if (found) {
      routeMap.delete(found[0])
    }
  }

  if (router.hasRoute(name)) {
    router.removeRoute(name)
  }
  for (const [path, record] of routeMap.entries()) {
    if (record.name === name) {
      routeMap.delete(path)
      break
    }
  }
  return true
}

function removePluginRoutesBySource(pluginID: number, source: 'script' | 'schema') {
  const routeMap = ensurePluginRouteMap(pluginID)
  const toDelete = Array.from(routeMap.values()).filter((record) => record.source === source)
  for (const record of toDelete) {
    if (router.hasRoute(record.name)) {
      router.removeRoute(record.name)
    }
    routeMap.delete(record.path)
  }
}

async function applySchemaRoutes(plugin: PluginMeta, api: ReturnType<typeof createPluginApi>) {
  const schema = parseConfigSchema(plugin.config_schema)
  const routeFields = schema.filter((field) => field.type === 'route_list')
  removePluginRoutesBySource(plugin.id, 'schema')
  if (!routeFields.length) return

  const settingsRes = await pluginAPI.getPluginSettings(plugin.id)
  const settings = settingsRes?.data?.settings || {}

  for (const field of routeFields) {
    const pages = parseRoutePages(settings[field.key] || '[]')
    pages.forEach((page, index) => {
      if (!page.path) return
      const routeName = `plugin_${plugin.id}_${sanitizeRouteName(field.key)}_${index}`
      registerPluginRoute(
        plugin,
        api,
        {
          path: page.path,
          name: routeName,
          title: page.title || plugin.name || `Plugin #${plugin.id}`,
          description: page.description || '插件配置中定义的页面',
          requiresAuth: page.requires_auth !== false,
          adminOnly: Boolean(page.admin_only),
          coWorkerOnly: Boolean(page.co_worker_only),
          render: () => page.content_html || '<div class="text-sm text-slate-400">该页面尚未配置内容</div>'
        },
        'schema'
      )
    })
  }
}

function runPluginScript(plugin: PluginMeta, code: string, api: ReturnType<typeof createPluginApi>) {
  const context = {
    plugin,
    api,
    console,
    storage: localStorage,
    router: {
      push: (to: any) => router.push(to),
      replace: (to: any) => router.replace(to),
      current: () => router.currentRoute.value,
      listPluginRoutes: () => listRegisteredPluginRoutes(plugin.id)
    },
    onMessage: (handler: (payload: any, meta: any) => void) => {
      if (!messageHandlers.has(plugin.id)) messageHandlers.set(plugin.id, [])
      messageHandlers.get(plugin.id)!.push(handler)
    },
    registerRoute: (routeDef: PluginRouteDefinition) => registerPluginRoute(plugin, api, routeDef, 'script'),
    unregisterRoute: (nameOrPath: string) => unregisterPluginRoute(plugin, nameOrPath)
  }

  try {
    const fn = new Function('plugin', 'api', 'context', code)
    fn(plugin, api, context)
  } catch (err) {
    console.error(`[Plugin:${plugin.name || plugin.id}] script error:`, err)
  }
}

async function loadOnePlugin(plugin: PluginMeta) {
  loadedPluginMeta.set(plugin.id, plugin)
  const api = createPluginApi()
  await applySchemaRoutes(plugin, api)

  const hasClient = plugin.has_client_script ?? plugin.has_script
  if (!hasClient || loadedPluginIds.has(plugin.id)) {
    return
  }

  const scriptRes = await pluginAPI.getPluginScript(plugin.id)
  const code = typeof scriptRes.data === 'string' ? scriptRes.data : ''
  if (!code.trim()) return
  runPluginScript(plugin, code, api)
  loadedPluginIds.add(plugin.id)
}

export async function loadPluginScripts() {
  try {
    const res = await pluginAPI.getPluginsWithCards()
    const plugins: PluginMeta[] = res.data || []

    for (const plugin of plugins) {
      try {
        await loadOnePlugin(plugin)
      } catch (err) {
        console.warn(`[Plugin:${plugin.name || plugin.id}] failed to load`, err)
      }
    }
  } catch (err) {
    console.warn('[Plugin] failed to load plugins', err)
  }
}

export async function initializePluginRuntime() {
  // Token存储在HttpOnly Cookie中，检查user信息判断是否已登录
  const user = localStorage.getItem('user')
  if (!user) return
  await loadPluginScripts()
}

export function listRegisteredPluginRoutes(pluginID?: number): Array<PluginRouteRecord & { plugin_id: number }> {
  const result: Array<PluginRouteRecord & { plugin_id: number }> = []
  for (const [id, routeMap] of pluginRoutes.entries()) {
    if (typeof pluginID === 'number' && id !== pluginID) continue
    for (const record of routeMap.values()) {
      result.push({ ...record, plugin_id: id })
    }
  }
  return result.sort((a, b) => a.path.localeCompare(b.path))
}

export async function refreshPluginConfiguredRoutes(pluginID: number) {
  const api = createPluginApi()
  let plugin = loadedPluginMeta.get(pluginID)
  if (!plugin) {
    const res = await pluginAPI.getPluginsWithCards()
    const found = (res.data || []).find((item: PluginMeta) => item.id === pluginID)
    if (!found) return
    plugin = found as PluginMeta
    loadedPluginMeta.set(pluginID, plugin)
  }
  if (!plugin) return
  await applySchemaRoutes(plugin, api)
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

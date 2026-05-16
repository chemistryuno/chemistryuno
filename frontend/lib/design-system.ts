export type DesignTone =
  | 'primary'
  | 'secondary'
  | 'success'
  | 'warning'
  | 'danger'
  | 'info'
  | 'special'
  | 'slate'

export type ComponentFamily =
  | 'pageShell'
  | 'headerNavigation'
  | 'sectionPanel'
  | 'card'
  | 'button'
  | 'iconButton'
  | 'formField'
  | 'badge'
  | 'tabs'
  | 'modal'
  | 'dataList'
  | 'emptyState'
  | 'loadingState'
  | 'errorState'
  | 'toast'
  | 'profileWidget'
  | 'gameCard'
  | 'roomCard'
  | 'announcement'
  | 'chatPanel'
  | 'adminSurface'

export type ComponentTemplate = {
  family: ComponentFamily
  source: string[]
  useWhen: string
  structure: string[]
  anatomy: string[]
  states: string[]
  responsive: string[]
  accessibility: string[]
  className: string
  example: string
}

export type ButtonVariant = 'solid' | 'soft' | 'ghost' | 'outline'
export type ButtonSize = 'xs' | 'sm' | 'md' | 'lg' | 'icon'
export type PanelDensity = 'compact' | 'normal' | 'spacious'

type ClassValue = string | false | null | undefined

const joinClasses = (...classes: ClassValue[]) => classes.filter(Boolean).join(' ')

export const designSources = {
  global: [
    'frontend/src/index.css',
    'frontend/src/theme.css',
    'frontend/src/styles/components.css',
    'frontend/src/styles/lobby.css',
  ],
  pages: [
    'frontend/src/pages/Lobby.vue',
    'frontend/src/pages/Profile.vue',
    'frontend/src/pages/Admin.vue',
    'frontend/src/pages/GameRoom.vue',
    'frontend/src/pages/Chat.vue',
  ],
  components: [
    'frontend/src/components/CustomDialog.vue',
    'frontend/src/components/ChatBox.vue',
    'frontend/src/components/AnnouncementTicker.vue',
    'frontend/src/components/GameToast.vue',
    'frontend/src/components/LevelProgress.vue',
    'frontend/src/components/profile/ProfileHeader.vue',
    'frontend/src/components/profile/StatsGrid.vue',
  ],
} as const

export const designTokens = {
  app: {
    viewport: 'min-h-screen bg-slate-50 text-slate-900 dark:bg-[#0a0a0c] dark:text-slate-200',
    contentWidth: 'max-w-[1400px] mx-auto px-4 md:px-6',
    backgroundDecor: 'fixed inset-0 overflow-hidden pointer-events-none',
  },
  text: {
    title: 'font-black uppercase tracking-tight text-slate-900 dark:text-white',
    heading: 'text-sm font-black uppercase tracking-widest text-slate-800 dark:text-white',
    label: 'text-[9px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500',
    meta: 'font-mono text-[9px] uppercase tracking-widest text-slate-400',
    body: 'text-xs leading-relaxed text-slate-600 dark:text-slate-300',
    value: 'font-mono font-black text-slate-900 dark:text-white',
  },
  surface: {
    page: 'bg-slate-50 dark:bg-[#0a0a0c]',
    panel: 'bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 shadow-sm',
    glass: 'bg-white/80 dark:bg-slate-900/60 border border-slate-200 dark:border-white/10 backdrop-blur-xl shadow-2xl',
    inset: 'bg-slate-50 dark:bg-white/[0.02] border border-slate-200 dark:border-white/5',
    input: 'bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10',
  },
  radius: {
    control: 'rounded-xl',
    panel: 'rounded-2xl',
    largePanel: 'rounded-[2rem]',
    pill: 'rounded-full',
  },
  motion: {
    in: 'animate-in fade-in',
    zoom: 'animate-in fade-in zoom-in',
    slideUp: 'animate-in fade-in slide-in-from-bottom-4',
    touch: 'active:scale-95 transition-transform',
    standard: 'transition-all duration-200',
  },
  focus: 'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-500',
  scroll: {
    normal: 'custom-scrollbar',
    hidden: 'custom-scrollbar-hidden',
  },
} as const

export const toneClasses: Record<DesignTone, {
  solid: string
  soft: string
  outline: string
  text: string
  border: string
  ring: string
  focusBorder: string
}> = {
  primary: {
    solid: 'bg-blue-600 hover:bg-blue-500 text-white shadow-blue-500/20',
    soft: 'bg-blue-500/10 hover:bg-blue-500/20 text-blue-600 dark:text-blue-400',
    outline: 'border-blue-500/20 text-blue-600 dark:text-blue-400 hover:bg-blue-500/10',
    text: 'text-blue-600 dark:text-blue-400',
    border: 'border-blue-500/20',
    ring: 'focus-visible:outline-blue-500',
    focusBorder: 'focus:border-blue-500/50',
  },
  secondary: {
    solid: 'bg-slate-700 hover:bg-slate-600 text-white shadow-slate-500/10',
    soft: 'bg-slate-100 hover:bg-slate-200 text-slate-600 dark:bg-white/5 dark:hover:bg-white/10 dark:text-slate-300',
    outline: 'border-slate-200 dark:border-white/10 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-white/5',
    text: 'text-slate-600 dark:text-slate-300',
    border: 'border-slate-200 dark:border-white/10',
    ring: 'focus-visible:outline-slate-400',
    focusBorder: 'focus:border-slate-400/50',
  },
  success: {
    solid: 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-emerald-500/20',
    soft: 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400',
    outline: 'border-emerald-500/20 text-emerald-600 dark:text-emerald-400 hover:bg-emerald-500/10',
    text: 'text-emerald-600 dark:text-emerald-400',
    border: 'border-emerald-500/20',
    ring: 'focus-visible:outline-emerald-500',
    focusBorder: 'focus:border-emerald-500/50',
  },
  warning: {
    solid: 'bg-amber-500 hover:bg-amber-400 text-white shadow-amber-500/20',
    soft: 'bg-amber-500/10 hover:bg-amber-500/20 text-amber-600 dark:text-amber-400',
    outline: 'border-amber-500/20 text-amber-600 dark:text-amber-400 hover:bg-amber-500/10',
    text: 'text-amber-600 dark:text-amber-400',
    border: 'border-amber-500/20',
    ring: 'focus-visible:outline-amber-500',
    focusBorder: 'focus:border-amber-500/50',
  },
  danger: {
    solid: 'bg-red-600 hover:bg-red-500 text-white shadow-red-500/20',
    soft: 'bg-red-500/10 hover:bg-red-500/20 text-red-600 dark:text-red-400',
    outline: 'border-red-500/20 text-red-600 dark:text-red-400 hover:bg-red-500/10',
    text: 'text-red-600 dark:text-red-400',
    border: 'border-red-500/20',
    ring: 'focus-visible:outline-red-500',
    focusBorder: 'focus:border-red-500/50',
  },
  info: {
    solid: 'bg-cyan-600 hover:bg-cyan-500 text-white shadow-cyan-500/20',
    soft: 'bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-600 dark:text-cyan-400',
    outline: 'border-cyan-500/20 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/10',
    text: 'text-cyan-600 dark:text-cyan-400',
    border: 'border-cyan-500/20',
    ring: 'focus-visible:outline-cyan-500',
    focusBorder: 'focus:border-cyan-500/50',
  },
  special: {
    solid: 'bg-purple-600 hover:bg-purple-500 text-white shadow-purple-500/20',
    soft: 'bg-purple-500/10 hover:bg-purple-500/20 text-purple-600 dark:text-purple-400',
    outline: 'border-purple-500/20 text-purple-600 dark:text-purple-400 hover:bg-purple-500/10',
    text: 'text-purple-600 dark:text-purple-400',
    border: 'border-purple-500/20',
    ring: 'focus-visible:outline-purple-500',
    focusBorder: 'focus:border-purple-500/50',
  },
  slate: {
    solid: 'bg-slate-900 hover:bg-slate-800 text-white shadow-slate-500/10 dark:bg-white/10 dark:hover:bg-white/15',
    soft: 'bg-slate-500/10 hover:bg-slate-500/20 text-slate-600 dark:text-slate-300',
    outline: 'border-slate-500/20 text-slate-600 dark:text-slate-300 hover:bg-slate-500/10',
    text: 'text-slate-500 dark:text-slate-400',
    border: 'border-slate-500/20',
    ring: 'focus-visible:outline-slate-500',
    focusBorder: 'focus:border-slate-500/50',
  },
}

const buttonSizes: Record<ButtonSize, string> = {
  xs: 'min-h-8 px-3 py-1.5 text-[9px]',
  sm: 'min-h-9 px-3.5 py-2 text-[10px]',
  md: 'min-h-10 px-4 py-2 text-xs',
  lg: 'min-h-11 px-5 py-3 text-xs',
  icon: 'h-10 w-10 p-0',
}

export function buttonClasses(options: {
  tone?: DesignTone
  variant?: ButtonVariant
  size?: ButtonSize
  block?: boolean
  loading?: boolean
} = {}) {
  const tone = options.tone ?? 'primary'
  const variant = options.variant ?? 'solid'
  const size = options.size ?? 'md'
  const color =
    variant === 'solid'
      ? toneClasses[tone].solid
      : variant === 'soft'
        ? toneClasses[tone].soft
        : variant === 'outline'
          ? joinClasses('border', toneClasses[tone].outline)
          : 'bg-transparent text-slate-500 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-white/5'

  return joinClasses(
    'inline-flex items-center justify-center gap-2 rounded-xl font-black uppercase tracking-widest shadow-lg transition-all active:scale-95 disabled:cursor-not-allowed disabled:opacity-50',
    buttonSizes[size],
    toneClasses[tone].ring,
    color,
    options.block && 'w-full',
    options.loading && 'pointer-events-none opacity-80',
  )
}

export function iconButtonClasses(options: {
  tone?: DesignTone
  size?: 'sm' | 'md' | 'lg'
  dangerHover?: boolean
} = {}) {
  const size = options.size ?? 'md'
  const tone = options.tone ?? 'slate'
  const sizes = {
    sm: 'h-8 w-8',
    md: 'h-10 w-10',
    lg: 'h-11 w-11',
  }

  return joinClasses(
    'inline-flex items-center justify-center rounded-xl transition-all active:scale-95',
    sizes[size],
    options.dangerHover
      ? 'text-slate-400 hover:bg-red-500/10 hover:text-red-500'
      : joinClasses(toneClasses[tone].text, 'hover:bg-slate-100 dark:hover:bg-white/5'),
    toneClasses[tone].ring,
  )
}

export function panelClasses(options: {
  density?: PanelDensity
  glass?: boolean
  interactive?: boolean
} = {}) {
  const density = options.density ?? 'normal'
  const padding = {
    compact: 'p-3',
    normal: 'p-4 sm:p-6',
    spacious: 'p-6 sm:p-8',
  }[density]

  return joinClasses(
    options.glass ? designTokens.surface.glass : designTokens.surface.panel,
    designTokens.radius.panel,
    padding,
    options.interactive && 'transition-all hover:border-blue-500/40 hover:bg-slate-50 dark:hover:bg-white/[0.08]',
  )
}

export function inputClasses(options: {
  tone?: DesignTone
  invalid?: boolean
  compact?: boolean
} = {}) {
  const tone = options.invalid ? 'danger' : options.tone ?? 'primary'
  return joinClasses(
    'w-full rounded-xl px-4 font-bold text-slate-900 transition-all placeholder:text-slate-400 focus:outline-none dark:text-white dark:placeholder:text-slate-600',
    options.compact ? 'py-2 text-[11px]' : 'py-3 text-sm',
    designTokens.surface.input,
    options.invalid ? 'border-red-500 focus:border-red-500' : toneClasses[tone].focusBorder,
  )
}

export function badgeClasses(options: {
  tone?: DesignTone
  pulse?: boolean
} = {}) {
  const tone = options.tone ?? 'primary'
  return joinClasses(
    'inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[9px] font-black uppercase',
    toneClasses[tone].soft,
    toneClasses[tone].border,
    options.pulse && 'before:h-1.5 before:w-1.5 before:rounded-full before:bg-current before:animate-pulse',
  )
}

export function modalClasses(options: {
  width?: 'sm' | 'md' | 'lg' | 'xl'
} = {}) {
  const widths = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-2xl',
  }
  return {
    overlay: 'viewport-modal-overlay bg-slate-900/60 dark:bg-black/80 backdrop-blur-md p-4',
    panel: joinClasses(
      'viewport-modal-panel w-full bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-[2rem] shadow-2xl animate-in zoom-in',
      widths[options.width ?? 'md'],
    ),
  }
}

export const componentTemplates: Record<ComponentFamily, ComponentTemplate> = {
  pageShell: {
    family: 'pageShell',
    source: ['App.vue', 'Lobby.vue', 'Profile.vue'],
    useWhen: 'A route owns the full viewport.',
    structure: ['root viewport', 'optional background decor', 'constrained content', 'modal mounts'],
    anatomy: ['slate page background', 'compact content width', 'subtle blue/cyan decor'],
    states: ['loading', 'empty', 'offline', 'auth redirect'],
    responsive: ['single column on mobile', 'split panes from lg/xl'],
    accessibility: ['main content remains reachable', 'focus is trapped only inside active modals'],
    className: joinClasses(designTokens.app.viewport, designTokens.app.contentWidth),
    example: '<main class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200" />',
  },
  headerNavigation: {
    family: 'headerNavigation',
    source: ['Lobby.vue', 'Profile.vue', 'Admin.vue'],
    useWhen: 'Route navigation, tabs, user chip, or back actions.',
    structure: ['left identity/action group', 'segmented nav', 'mobile drawer trigger'],
    anatomy: ['glass header', 'rounded segmented controls', 'icon plus tiny label'],
    states: ['active', 'hover', 'disabled', 'mobile open'],
    responsive: ['drawer or horizontal scroll for long nav'],
    accessibility: ['active state is visible', 'icon-only buttons have labels'],
    className: 'bg-white/60 dark:bg-black/40 backdrop-blur-2xl border-b border-slate-200 dark:border-white/5',
    example: '<nav class="p-1 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-2xl" />',
  },
  sectionPanel: {
    family: 'sectionPanel',
    source: ['Profile.vue', 'LevelProgress.vue', 'Admin.vue'],
    useWhen: 'A page section, settings area, or data block.',
    structure: ['surface', 'optional header', 'content', 'optional footer'],
    anatomy: ['white/dark panel', 'thin border', 'icon square header'],
    states: ['static', 'interactive hover', 'loading'],
    responsive: ['p-4 on mobile', 'p-6 or p-8 on desktop'],
    accessibility: ['heading level follows page outline'],
    className: panelClasses(),
    example: '<section class="bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/10 rounded-2xl p-6" />',
  },
  card: {
    family: 'card',
    source: ['StatsGrid.vue', 'Lobby.vue', 'ProfileHeader.vue'],
    useWhen: 'Repeated objects, stats, selectable options, and summaries.',
    structure: ['status/meta', 'primary title/value', 'details', 'actions'],
    anatomy: ['rounded-xl or rounded-2xl', 'small uppercase labels', 'mono values'],
    states: ['hover', 'selected', 'disabled', 'loading'],
    responsive: ['stable grid tracks', 'truncate long content'],
    accessibility: ['clickable cards use button/link semantics'],
    className: panelClasses({ density: 'compact', interactive: true }),
    example: '<button class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl p-3" />',
  },
  button: {
    family: 'button',
    source: ['components.css', 'Lobby.vue', 'CustomDialog.vue'],
    useWhen: 'A clear command action.',
    structure: ['button', 'optional icon', 'label', 'optional loader'],
    anatomy: ['primary blue', 'semantic accents', 'uppercase heavy text'],
    states: ['hover', 'focus-visible', 'active', 'disabled', 'loading'],
    responsive: ['icon-only when space is tight', '40px minimum touch target'],
    accessibility: ['destructive actions use danger tone', 'icon-only actions have labels'],
    className: buttonClasses(),
    example: '<button class="inline-flex items-center gap-2 rounded-xl bg-blue-600 px-4 py-2 text-xs font-black uppercase text-white" />',
  },
  iconButton: {
    family: 'iconButton',
    source: ['Profile.vue', 'CustomDialog.vue', 'Lobby.vue'],
    useWhen: 'Close, back, refresh, edit, delete, copy, expand, or tool actions.',
    structure: ['square button', 'lucide icon'],
    anatomy: ['slate default', 'hover surface', 'semantic hover for danger'],
    states: ['hover', 'focus-visible', 'active', 'disabled'],
    responsive: ['h-10 w-10 on mobile', 'h-8 w-8 in dense tables'],
    accessibility: ['requires aria-label or title'],
    className: iconButtonClasses(),
    example: '<button aria-label="Close" class="h-10 w-10 rounded-xl hover:bg-slate-100 dark:hover:bg-white/5" />',
  },
  formField: {
    family: 'formField',
    source: ['Admin.vue', 'Lobby.vue', 'CustomDialog.vue'],
    useWhen: 'Text input, search, select, textarea, numeric settings, and modal forms.',
    structure: ['label', 'optional hint', 'control', 'optional validation text'],
    anatomy: ['tiny uppercase label', 'slate input surface', 'accent focus border'],
    states: ['default', 'focus', 'invalid', 'disabled', 'loading'],
    responsive: ['use mobile-safe sizing for frequent text fields'],
    accessibility: ['label is associated with control', 'validation text is close to field'],
    className: inputClasses({ compact: true }),
    example: '<input class="bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-xl px-4 py-2" />',
  },
  badge: {
    family: 'badge',
    source: ['components.css', 'Lobby.vue', 'AnnouncementTicker.vue'],
    useWhen: 'Role, live state, room status, level, mode, severity, or count.',
    structure: ['inline-flex pill', 'optional dot or icon', 'label'],
    anatomy: ['tiny uppercase label', 'tinted background', 'semantic border'],
    states: ['live', 'paused', 'offline', 'critical'],
    responsive: ['hide long labels before marker'],
    accessibility: ['do not rely on color alone for critical state'],
    className: badgeClasses({ tone: 'success', pulse: true }),
    example: '<span class="inline-flex items-center gap-1 rounded-full border border-emerald-500/20 bg-emerald-500/10 px-2 py-0.5 text-[9px] font-black uppercase" />',
  },
  tabs: {
    family: 'tabs',
    source: ['Profile.vue', 'Admin.vue'],
    useWhen: 'Switching local panels or routed categories.',
    structure: ['segmented wrapper', 'repeated buttons', 'active state'],
    anatomy: ['rounded-2xl wrapper', 'rounded-xl items', 'active tinted background'],
    states: ['active', 'hover', 'disabled', 'overflow'],
    responsive: ['horizontal scroll or drawer for long tab sets'],
    accessibility: ['use buttons for local state and links for routes'],
    className: 'flex gap-1 overflow-x-auto rounded-2xl border border-slate-200 bg-white p-1 dark:border-white/10 dark:bg-white/5',
    example: '<div class="flex gap-1 rounded-2xl border border-slate-200 bg-white p-1 dark:border-white/10 dark:bg-white/5" />',
  },
  modal: {
    family: 'modal',
    source: ['CustomDialog.vue', 'Lobby.vue', 'Admin.vue'],
    useWhen: 'Blocking confirmation, creation/edit forms, details, legal content, and settings.',
    structure: ['Teleport', 'overlay', 'panel', 'header', 'scroll body', 'footer actions'],
    anatomy: ['viewport overlay', 'blurred dark backdrop', 'large rounded panel'],
    states: ['enter', 'loading', 'validation error', 'disabled confirm'],
    responsive: ['max-height 90dvh', 'body scroll contained'],
    accessibility: ['close button is reachable', 'prompt inputs handle IME composition'],
    className: modalClasses().panel,
    example: '<div class="viewport-modal-overlay bg-slate-900/60 dark:bg-black/80 backdrop-blur-md p-4" />',
  },
  dataList: {
    family: 'dataList',
    source: ['Admin.vue', 'AdminSurveyResponses.vue', 'ReplayRoom.vue'],
    useWhen: 'Admin records, logs, users, history, surveys, plugins, and dense settings.',
    structure: ['toolbar filters', 'scroll container', 'rows/cards', 'expandable detail'],
    anatomy: ['thin borders', 'tiny metadata', 'mono IDs and timestamps'],
    states: ['selected', 'expanded', 'empty', 'loading', 'filtered', 'live pending'],
    responsive: ['tables become stacked cards on mobile'],
    accessibility: ['table semantics or button semantics for expandable rows'],
    className: 'rounded-2xl border border-slate-200 bg-white dark:border-white/10 dark:bg-[#111114] overflow-hidden',
    example: '<div class="rounded-2xl border border-slate-200 bg-white dark:border-white/10 dark:bg-[#111114]" />',
  },
  emptyState: {
    family: 'emptyState',
    source: ['lobby.css', 'Profile.vue', 'ChatBox.vue'],
    useWhen: 'No rooms, records, messages, achievements, or results.',
    structure: ['centered icon', 'primary label', 'secondary mono caption', 'optional action'],
    anatomy: ['low opacity slate icon', 'dashed or tinted container'],
    states: ['filtered empty', 'first-run empty'],
    responsive: ['compact in sidebars', 'larger in main content'],
    accessibility: ['text explains the state'],
    className: 'flex flex-col items-center justify-center rounded-2xl border-2 border-dashed border-slate-200 bg-slate-50 py-12 text-center dark:border-white/5 dark:bg-white/[0.02]',
    example: '<div class="flex flex-col items-center justify-center rounded-2xl border-2 border-dashed border-slate-200 py-12" />',
  },
  loadingState: {
    family: 'loadingState',
    source: ['App.vue', 'LevelProgress.vue', 'components.css'],
    useWhen: 'Route boot, async panels, tables, forms, and live streams.',
    structure: ['spinner or skeleton', 'reserved content space'],
    anatomy: ['blue/cyan spinner', 'slate pulse skeleton', 'atom loader for app boot'],
    states: ['loading', 'retryable error', 'stale cache'],
    responsive: ['reserve height to avoid jumps'],
    accessibility: ['use aria-busy for regions'],
    className: 'flex min-h-32 items-center justify-center',
    example: '<div class="flex min-h-32 items-center justify-center" aria-busy="true" />',
  },
  errorState: {
    family: 'errorState',
    source: ['ChatBox.vue', 'LevelProgress.vue', 'CustomDialog.vue'],
    useWhen: 'Failed loads, blocked account actions, dangerous admin actions, or unavailable features.',
    structure: ['tinted panel', 'icon square', 'message', 'metadata', 'recovery action'],
    anatomy: ['rose/red error', 'amber warning', 'slate unavailable'],
    states: ['retry', 'dismiss', 'appeal', 'details expanded'],
    responsive: ['single-column actions on mobile'],
    accessibility: ['critical dynamic errors use role alert'],
    className: 'rounded-2xl border border-rose-500/20 bg-rose-500/10 p-4',
    example: '<div role="alert" class="rounded-2xl border border-rose-500/20 bg-rose-500/10 p-4" />',
  },
  toast: {
    family: 'toast',
    source: ['GameToast.vue', 'ReactionToast.vue', 'CustomDialog.vue'],
    useWhen: 'Temporary game feedback, system notification, success/error message, or live status.',
    structure: ['fixed container', 'toast card', 'icon', 'title/message', 'optional progress'],
    anatomy: ['glass tinted card', 'semantic glow for game feedback'],
    states: ['info', 'success', 'warning', 'error', 'enter', 'leave'],
    responsive: ['side inset on mobile'],
    accessibility: ['important messages are announced by text'],
    className: 'rounded-xl border border-blue-500/20 bg-blue-500/10 px-4 py-3 shadow-lg backdrop-blur-xl',
    example: '<div class="fixed right-4 top-20 z-[9999] flex flex-col gap-3 max-sm:left-3 max-sm:right-3" />',
  },
  profileWidget: {
    family: 'profileWidget',
    source: ['ProfileHeader.vue', 'StatsGrid.vue', 'SettingsPanel.vue'],
    useWhen: 'User cards, stats, security settings, identity, level, account controls.',
    structure: ['profile card', 'avatar square', 'UID/meta', 'nickname', 'role badge', 'stats/actions'],
    anatomy: ['top accent line', 'nested rounded avatar', 'mono numeric values'],
    states: ['editable', 'admin/user roles', 'missing data', 'loading'],
    responsive: ['sidebar on desktop', 'full width on mobile'],
    accessibility: ['edit buttons have labels', 'avatar has fallback'],
    className: 'relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-xl dark:border-white/10 dark:bg-[#111114]',
    example: '<section class="relative overflow-hidden rounded-2xl border border-slate-200 bg-white p-6 shadow-xl dark:border-white/10 dark:bg-[#111114]" />',
  },
  gameCard: {
    family: 'gameCard',
    source: ['index.css', 'GameRoom.vue', 'ChemicalKeyboard.vue'],
    useWhen: 'Chemistry cards, reaction cards, deck details, hand cards, and game previews.',
    structure: ['fixed aspect card', 'formula', 'type/status meta', 'optional count/action'],
    anatomy: ['uno-card', 'element classes', 'stable dimensions'],
    states: ['playable', 'selected', 'disabled', 'reaction candidate', 'animating'],
    responsive: ['scale by container constraints'],
    accessibility: ['button label includes formula and action'],
    className: 'uno-card aspect-[2.5/3.5] rounded-2xl p-3 shadow-lg transition-all',
    example: '<button class="uno-card element-H aspect-[2.5/3.5] w-24 rounded-2xl p-3" />',
  },
  roomCard: {
    family: 'roomCard',
    source: ['frontend/src/styles/lobby.css', 'Lobby.vue'],
    useWhen: 'Lobby room list, active room admin list, rejoinable room, or spectate card.',
    structure: ['status header', 'type badges', 'title/id', 'metadata', 'occupancy', 'actions'],
    anatomy: ['room-card glass surface', 'status pill', 'mono room ID', 'progress track'],
    states: ['waiting', 'starting', 'playing', 'full', 'private', 'ranked'],
    responsive: ['dense room-grid on mobile', '2-3 columns on larger screens'],
    accessibility: ['join/spectate/terminate are separate buttons'],
    className: 'room-card',
    example: '<article class="room-card"><header class="room-card-header" /></article>',
  },
  announcement: {
    family: 'announcement',
    source: ['AnnouncementTicker.vue', 'Lobby.vue'],
    useWhen: 'Ticker, persistent announcement card, join modal, maintenance or emergency notice.',
    structure: ['type badge', 'content', 'optional close', 'timing metadata'],
    anatomy: ['fixed blue ticker', 'semantic tinted cards'],
    states: ['emergency', 'maintenance', 'info', 'dismissed', 'rotating', 'expired'],
    responsive: ['ticker truncates content', 'modal uses modal template'],
    accessibility: ['critical type is shown with text'],
    className: 'rounded-xl border border-amber-500/20 bg-amber-500/10 p-4',
    example: '<div class="rounded-xl border border-amber-500/20 bg-amber-500/10 p-4" />',
  },
  chatPanel: {
    family: 'chatPanel',
    source: ['ChatBox.vue', 'ChatBox.css', 'Chat.vue'],
    useWhen: 'Global chat, room chat, private messages, friend interactions, invite messages.',
    structure: ['shell', 'optional header', 'scrollable messages', 'system messages', 'input bar'],
    anatomy: ['rounded 28px glass panel', 'blue own-message bubble', 'custom scrollbar'],
    states: ['empty', 'system', 'own/other', 'private', 'game invite', 'blocked', 'composing'],
    responsive: ['full-width mobile', 'compact header', 'larger touch input'],
    accessibility: ['send on Enter respects IME composition'],
    className: 'flex flex-col bg-white/95 dark:bg-slate-900/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[28px] overflow-hidden shadow-2xl',
    example: '<section class="flex flex-col overflow-hidden rounded-[28px] border bg-white/95 backdrop-blur-xl dark:bg-slate-900/60" />',
  },
  adminSurface: {
    family: 'adminSurface',
    source: ['Admin.vue', 'Admin.css', 'AdminAnticheat.vue'],
    useWhen: 'Admin dashboards, logs, surveys, plugin records, anticheat panels, and system config.',
    structure: ['stats', 'tab nav', 'filter toolbar', 'dense records', 'detail modal/drawer'],
    anatomy: ['dense typography', 'mono IDs', 'cyan/emerald live states'],
    states: ['loading', 'filtered', 'expanded', 'live', 'reconnecting', 'offline', 'permission-limited'],
    responsive: ['filters wrap', 'tables become cards', 'modals use viewport overlay'],
    accessibility: ['filters retain focus during live updates'],
    className: 'space-y-4',
    example: '<section class="space-y-4"><div class="grid gap-3 md:grid-cols-4" /></section>',
  },
}

export const domainAdaptationRules = {
  chemistryCards: [
    'Use uno-card and existing element/card classes before adding new color rules.',
    'Keep board/card dimensions stable with aspect ratio and fixed tracks.',
    'Pair formula visuals with accessible action labels.',
  ],
  roomOccupancy: [
    'Use status pill, player count, and progress bar together.',
    'Waiting uses emerald; starting/live uses blue or cyan; ranked/playing uses amber; blocked/full uses slate or red.',
  ],
  reactionFeedback: [
    'Use concise animation and semantic color.',
    'Do not use animation alone to indicate whether a play is valid.',
  ],
  phlogistonAndLevels: [
    'Use mono numeric values and amber for phlogiston.',
    'Use blue-to-cyan progress fills and LevelBadge for level context.',
  ],
  tutorialsAndAnnouncements: [
    'Highlight tutorial targets with restrained blue/cyan accents.',
    'Ticker announcements stay fixed top and compact.',
    'Persistent announcements use the announcement or modal template.',
  ],
} as const

export const consistencyChecklist = [
  'Select the closest component template.',
  'Define both light and dark surfaces, borders, text, icons, and accents.',
  'Specify mobile and desktop layout behavior.',
  'Keep touch targets at least 40px, preferably 44px on mobile.',
  'Use lucide-vue-next icons for command buttons when available.',
  'Cover hover, focus-visible, active, disabled, loading, selected, expanded, and error states as applicable.',
  'Handle long text without overlap.',
  'Use existing animation utilities and preserve reduced-motion behavior.',
  'Use custom scrollbar utilities for dense scroll regions.',
  'Use viewport-modal-overlay and viewport-modal-panel for full-viewport modals.',
  'Make domain context visible through icons, labels, or metadata.',
  'Document intentional divergence as domain-specific or add a new template.',
] as const

export function getComponentTemplate(family: ComponentFamily) {
  return componentTemplates[family]
}

export const chemistryUnoDesignLibrary = {
  designSources,
  designTokens,
  toneClasses,
  componentTemplates,
  domainAdaptationRules,
  consistencyChecklist,
  buttonClasses,
  iconButtonClasses,
  panelClasses,
  inputClasses,
  badgeClasses,
  modalClasses,
  getComponentTemplate,
} as const

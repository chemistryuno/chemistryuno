<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-800 dark:text-slate-300 font-sans selection:bg-blue-500/30 transition-colors duration-500">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 rounded-full blur-[120px] animate-pulse"></div>
      <div class="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-purple-600/5 rounded-full blur-[120px]"></div>
    </div>

    <div class="relative z-10 flex flex-col min-h-screen">
      <!-- Top Navigation -->
      <header class="px-4 py-3 sm:px-6 sm:py-4 border-b border-slate-200 dark:border-white/5 backdrop-blur-md bg-white/70 dark:bg-black/20 sticky top-0 z-50">
        <div class="max-w-[1400px] mx-auto flex justify-between items-center">
          <div class="flex items-center gap-3">
            <button @click="router.push('/')" class="p-2 hover:bg-slate-200 dark:hover:bg-white/5 rounded-xl transition-all text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-white group">
              <ArrowLeft class="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
            </button>
            <div>
              <h1 class="text-base sm:text-xl font-black text-slate-900 dark:text-white tracking-tighter flex items-center gap-2">
                <PhlogistonIcon :size="20" color="#f59e0b" class="shrink-0" />
                实验室储备燃素排行榜
              </h1>
              <p class="text-[8px] text-slate-500 font-mono uppercase tracking-[0.15em] mt-0.5 hidden sm:block">全球科研储备燃素实时榜单</p>
            </div>
          </div>
        </div>
      </header>

      <main class="flex-1 max-w-[1100px] mx-auto w-full px-3 sm:px-6 py-4 sm:py-6">
        <!-- Top Section: Mode Switch & Stats -->
        <div class="flex flex-col gap-3 mb-4 sm:mb-6">
          <!-- Row 1: Mode Switch + Search -->
          <div class="flex items-center gap-2 sm:gap-3">
            <!-- Mode Switch Tabs -->
            <div class="flex items-center gap-1 bg-slate-200/50 dark:bg-white/5 p-1 rounded-xl border border-slate-200 dark:border-white/5 shrink-0">
              <button
                @click="rankingMode = 'total'"
                :class="cn(
                  'px-3 py-1.5 rounded-lg text-[9px] font-black uppercase tracking-wider transition-all',
                  rankingMode === 'total'
                    ? 'bg-blue-600/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
                )"
              >全量存储</button>
              <button
                @click="rankingMode = 'monthly'"
                :class="cn(
                  'px-3 py-1.5 rounded-lg text-[9px] font-black uppercase tracking-wider transition-all',
                  rankingMode === 'monthly'
                    ? 'bg-indigo-600/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20 shadow-sm'
                    : 'text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'
                )"
              >本月活跃</button>
            </div>

            <!-- Search Bar -->
            <div class="flex-1 relative group">
              <div class="absolute inset-0 bg-blue-500/5 rounded-xl blur group-focus-within:bg-blue-500/10 transition-all"></div>
              <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-slate-400 group-focus-within:text-blue-500 transition-colors" />
              <input
                v-model="searchTerm"
                placeholder="搜索玩家..."
                class="relative w-full h-9 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl pl-9 pr-4 text-[11px] text-slate-700 dark:text-white focus:outline-none focus:border-blue-500/50 transition-all font-medium"
              />
              <div v-if="isSearching" class="absolute right-3 top-1/2 -translate-y-1/2">
                <Loader2 class="w-3.5 h-3.5 text-blue-500 animate-spin" />
              </div>
            </div>
          </div>

          <!-- Row 2: Stats chips -->
          <div class="grid grid-cols-3 gap-2 sm:flex sm:flex-wrap sm:gap-2">
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-2.5 py-2 flex items-center gap-1.5 shadow-sm">
              <div class="w-6 h-6 bg-amber-500/10 border border-amber-500/20 rounded-lg flex items-center justify-center text-amber-500 shrink-0">
                <Target class="w-3 h-3" />
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">衰减</span>
                <p class="text-[9px] font-bold text-slate-600 dark:text-slate-300 leading-tight truncate">前10%每周-2%</p>
              </div>
            </div>
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-2.5 py-2 flex items-center gap-1.5 shadow-sm">
              <div class="w-6 h-6 bg-blue-500/10 border border-blue-500/20 rounded-lg flex items-center justify-center text-blue-500 shrink-0">
                <RefreshCw class="w-3 h-3" />
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">赛季</span>
                <p class="text-[9px] font-bold text-slate-600 dark:text-slate-300 leading-tight">活跃中</p>
              </div>
            </div>
            <div class="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-xl px-2.5 py-2 flex items-center gap-1.5 shadow-sm">
              <div class="w-6 h-6 bg-purple-500/10 border border-purple-500/20 rounded-lg flex items-center justify-center text-purple-500 shrink-0">
                <ShieldCheck class="w-3 h-3" />
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-[7px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-none mb-0.5">总数</span>
                <p class="text-[9px] font-bold text-slate-600 dark:text-slate-300 leading-tight">{{ leaderboard.length }}人</p>
              </div>
            </div>
          </div>
        </div>

        <!-- Leaderboard Container -->
        <div class="bg-white/80 dark:bg-[#121216]/60 backdrop-blur-xl border border-slate-200 dark:border-white/10 rounded-[20px] sm:rounded-[24px] overflow-hidden shadow-2xl dark:shadow-none">
          <div v-if="loading" class="py-20 flex flex-col items-center justify-center">
            <Loader2 class="w-8 h-8 animate-spin text-blue-500 mb-3" />
            <p class="text-[9px] font-black uppercase tracking-widest text-slate-500">加载排行榜...</p>
          </div>
          <div v-else>
            <!-- ===== Mobile Card List (hidden on sm+) ===== -->
            <div class="sm:hidden divide-y divide-slate-100 dark:divide-white/5">
              <div
                v-for="(player, idx) in (searchTerm ? searchResults : leaderboard)"
                :key="player.uid"
                :class="cn(
                  'flex items-center gap-2.5 px-3 py-3 transition-colors',
                  Number(player.uid) === Number(user.uid) ? 'bg-blue-50/70 dark:bg-blue-500/[0.04]' : 'active:bg-slate-50 dark:active:bg-white/[0.02]'
                )"
              >
                <!-- Rank badge -->
                <span :class="cn(
                  'w-7 h-7 rounded-lg flex items-center justify-center text-[11px] font-black italic shadow shrink-0',
                  (!searchTerm ? idx : (rankingMode === 'monthly' ? player.monthly_rank : player.rank) - 1) === 0 ? 'bg-amber-500 text-amber-950' :
                  (!searchTerm ? idx : (rankingMode === 'monthly' ? player.monthly_rank : player.rank) - 1) === 1 ? 'bg-slate-300 text-slate-900' :
                  (!searchTerm ? idx : (rankingMode === 'monthly' ? player.monthly_rank : player.rank) - 1) === 2 ? 'bg-amber-700 text-white' :
                  'bg-slate-100 dark:bg-white/5 text-slate-500'
                )">
                  {{ searchTerm ? (rankingMode === 'monthly' ? player.monthly_rank : player.rank) : idx + 1 }}
                </span>

                <!-- Avatar -->
                <div
                  @click="showResearcherProfile(player.uid)"
                  class="relative w-10 h-10 rounded-xl bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center overflow-hidden shrink-0 cursor-pointer active:scale-95 transition-transform shadow-inner"
                >
                  <UserAvatar :avatar="player.avatar" />
                  <div v-if="player.is_online" class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-emerald-500 border-2 border-white dark:border-[#121216] rounded-full"></div>
                </div>

                <!-- Name + meta -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1 flex-wrap">
                    <span
                      @click="showResearcherProfile(player.uid)"
                      class="text-xs font-black text-slate-900 dark:text-white truncate cursor-pointer hover:text-blue-500 transition-colors max-w-[120px]"
                    >{{ player.nickname || player.username }}</span>
                    <span v-if="Number(player.uid) === Number(user.uid)" class="text-[7px] bg-blue-600 px-1 py-0.5 rounded font-black text-white shrink-0">我</span>
                    <span v-if="player.is_banned" class="text-[7px] bg-red-600 px-1 py-0.5 rounded font-black text-white animate-pulse shrink-0">作弊</span>
                  </div>
                  <div class="flex items-center gap-1.5 mt-0.5 flex-wrap">
                    <div class="flex items-center gap-0.5 bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded border border-slate-200 dark:border-white/5">
                      <LevelBadge :level="player.level || 1" :tier="player.tier" :tier-name="player.tier_name" size="xs" :show-level="false" />
                      <span class="text-[7px] font-black text-slate-500 dark:text-slate-400">Lv.{{ player.level || 1 }}</span>
                    </div>
                    <span v-if="player.total_games > 0" class="text-[7px] text-slate-400">胜率 {{ Math.round((player.win_count / player.total_games) * 100) }}%</span>
                    <span :class="['text-[7px] font-bold', player.is_online ? 'text-emerald-500' : 'text-slate-400/70']">
                      {{ player.is_online ? '在线' : formatLastOfflineText(player.last_offline_at) }}
                    </span>
                    <span v-if="player.bounty > 0" class="flex items-center gap-0.5 text-rose-500">
                      <Flame class="w-2.5 h-2.5" />
                      <span class="text-[7px] font-black">{{ player.bounty }}</span>
                    </span>
                  </div>
                </div>

                <!-- Phlogiston + action buttons -->
                <div class="flex flex-col items-end gap-1.5 shrink-0">
                  <div class="flex items-center gap-1">
                    <span class="text-sm font-black text-amber-600 dark:text-amber-500 font-mono leading-none">
                      {{ Math.floor(rankingMode === 'monthly' ? player.monthly_points : player.points) }}
                    </span>
                    <PhlogistonIcon :size="12" color="#f59e0b" />
                  </div>
                  <div v-if="Number(player.uid) !== Number(user.uid)" class="flex items-center gap-1">
                    <button v-if="player.is_online" @click="handleDuel(player)" title="单挑"
                      class="p-1.5 bg-blue-600/10 hover:bg-blue-600 text-blue-600 hover:text-white border border-blue-500/20 rounded-lg transition-all active:scale-95">
                      <Swords class="w-3 h-3" />
                    </button>
                    <button @click="openBountyModal(player)" title="悬赏"
                      class="p-1.5 bg-rose-600/10 hover:bg-rose-600 text-rose-600 hover:text-white border border-rose-500/20 rounded-lg transition-all active:scale-95">
                      <Crosshair class="w-3 h-3" />
                    </button>
                    <button v-if="!isFriend(player.uid)" @click="handleAddFriend(player)" title="加好友"
                      class="p-1.5 bg-amber-600/10 hover:bg-amber-600 text-amber-600 hover:text-white border border-amber-500/20 rounded-lg transition-all active:scale-95">
                      <UserPlus class="w-3 h-3" />
                    </button>
                    <button v-else @click="startPrivateChat(player)" title="私信"
                      class="p-1.5 bg-emerald-600/10 hover:bg-emerald-600 text-emerald-600 hover:text-white border border-emerald-500/20 rounded-lg transition-all active:scale-95">
                      <MessageCircle class="w-3 h-3" />
                    </button>
                  </div>
                  <span v-else class="text-[7px] font-black text-blue-600 dark:text-blue-500 uppercase tracking-wider italic">本人</span>
                </div>
              </div>

              <!-- 自我排名 (Mobile) -->
              <div
                v-if="!searchTerm && myRankInfo && !leaderboard.find(p => Number(p.uid) === Number(user.uid))"
                class="flex items-center gap-2.5 px-3 py-3 bg-blue-50/50 dark:bg-blue-500/[0.05] border-t-2 border-dashed border-blue-200 dark:border-blue-500/20"
              >
                <span class="w-7 h-7 rounded-lg flex items-center justify-center text-[11px] font-black italic bg-blue-100 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 shrink-0">
                  {{ myRankInfo.rank }}
                </span>
                <div class="relative w-10 h-10 rounded-xl bg-blue-100 dark:bg-blue-500/20 border border-blue-200 dark:border-blue-500/30 flex items-center justify-center overflow-hidden shrink-0 shadow-inner">
                  <UserAvatar :avatar="myRankInfo.avatar" />
                  <div v-if="myRankInfo.is_online" class="absolute bottom-0 right-0 w-2.5 h-2.5 bg-emerald-500 border-2 border-white dark:border-[#121216] rounded-full"></div>
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-1">
                    <span class="text-xs font-black text-blue-600 dark:text-blue-400 truncate max-w-[120px]">{{ myRankInfo.nickname || myRankInfo.username }}</span>
                    <span class="text-[7px] bg-blue-600 px-1 py-0.5 rounded font-black text-white shrink-0">我</span>
                  </div>
                  <div class="flex items-center gap-1.5 mt-0.5 flex-wrap">
                    <div class="flex items-center gap-0.5 bg-blue-500/10 dark:bg-blue-500/5 px-1.5 py-0.5 rounded border border-blue-500/20">
                      <LevelBadge :level="myRankInfo.level || 1" :tier="myRankInfo.tier" :tier-name="myRankInfo.tier_name" size="xs" :show-level="false" />
                      <span class="text-[7px] font-black text-blue-600 dark:text-blue-400">Lv.{{ myRankInfo.level || 1 }}</span>
                    </div>
                    <span class="text-[7px] text-blue-400/60">百名开外</span>
                  </div>
                </div>
                <div class="flex flex-col items-end shrink-0">
                  <div class="flex items-center gap-1">
                    <span class="text-sm font-black text-blue-600 dark:text-blue-400 font-mono">
                      {{ Math.floor(rankingMode === 'monthly' ? myRankInfo.monthly_points : myRankInfo.points) }}
                    </span>
                    <PhlogistonIcon :size="12" color="#2563eb" />
                  </div>
                  <span class="text-[7px] font-black text-blue-600 dark:text-blue-500 uppercase tracking-wider italic mt-0.5">已认证</span>
                </div>
              </div>
            </div>

            <!-- ===== Desktop Table (hidden below sm) ===== -->
            <div class="hidden sm:block overflow-x-auto">
              <table class="w-full border-collapse">
                <thead>
                  <tr class="bg-slate-50/50 dark:bg-white/[0.02] border-b border-slate-100 dark:border-white/5 text-left">
                    <th class="px-6 py-4 text-[9px] font-black text-slate-500 uppercase tracking-widest">{{ searchTerm ? '搜索' : '排名' }}</th>
                    <th class="px-5 py-4 text-[9px] font-black text-slate-500 uppercase tracking-widest">玩家</th>
                    <th class="px-5 py-4 text-[9px] font-black text-slate-500 uppercase tracking-widest">储备燃素</th>
                    <th class="px-5 py-4 text-[9px] font-black text-slate-500 uppercase tracking-widest">悬赏</th>
                    <th class="px-6 py-4 text-[9px] font-black text-slate-500 uppercase tracking-widest text-right">操作</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-white/5">
                  <tr
                    v-for="(player, idx) in (searchTerm ? searchResults : leaderboard)"
                    :key="player.uid"
                    :class="cn(
                      'group transition-colors',
                      Number(player.uid) === Number(user.uid) ? 'bg-blue-50/70 dark:bg-blue-500/[0.03]' : 'hover:bg-slate-50/50 dark:hover:bg-white/[0.02]'
                    )"
                  >
                    <td class="px-6 py-2">
                      <div class="flex items-center gap-3">
                        <template v-if="!searchTerm">
                          <span :class="cn(
                            'w-6 h-6 rounded-lg flex items-center justify-center text-[10px] font-black italic shadow-lg',
                            idx === 0 ? 'bg-amber-500 text-amber-950 dark:text-black' :
                            idx === 1 ? 'bg-slate-300 text-slate-900 dark:text-black' :
                            idx === 2 ? 'bg-amber-700 text-white' :
                            'bg-slate-100 dark:bg-white/5 text-slate-500'
                          )">{{ idx + 1 }}</span>
                        </template>
                        <template v-else>
                          <span :class="cn(
                            'w-6 h-6 rounded-lg flex items-center justify-center text-[10px] font-black italic shadow-lg',
                            (rankingMode === 'monthly' ? player.monthly_rank : player.rank) === 1 ? 'bg-amber-500 text-amber-950 dark:text-black' :
                            (rankingMode === 'monthly' ? player.monthly_rank : player.rank) === 2 ? 'bg-slate-300 text-slate-900 dark:text-black' :
                            (rankingMode === 'monthly' ? player.monthly_rank : player.rank) === 3 ? 'bg-amber-700 text-white' :
                            'bg-slate-100 dark:bg-white/5 text-slate-500'
                          )">{{ rankingMode === 'monthly' ? player.monthly_rank : player.rank }}</span>
                        </template>
                      </div>
                    </td>
                    <td class="px-5 py-2">
                      <div class="flex items-center gap-3">
                        <div
                          @click="showResearcherProfile(player.uid)"
                          class="relative w-8 h-8 rounded-lg bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center text-sm overflow-hidden shrink-0 cursor-pointer hover:ring-2 hover:ring-blue-500/50 transition-all shadow-inner"
                        >
                          <UserAvatar :avatar="player.avatar" />
                          <div v-if="player.is_online" class="absolute bottom-0 right-0 w-2 h-2 bg-emerald-500 border-2 border-white dark:border-[#121216] rounded-full"></div>
                        </div>
                        <div class="flex flex-col">
                          <span
                            @click="showResearcherProfile(player.uid)"
                            class="text-xs font-black text-slate-900 dark:text-white group-hover:text-blue-500 transition-colors flex items-center gap-1.5 flex-wrap cursor-pointer"
                          >
                            {{ player.nickname || player.username }}
                            <span v-if="Number(player.uid) === Number(user.uid)" class="text-[7px] bg-blue-600 px-1 py-0.5 rounded uppercase font-black tracking-widest text-white">我</span>
                            <span v-if="player.is_banned" class="text-[7px] bg-red-600 px-1 py-0.5 rounded uppercase font-black tracking-widest text-white animate-pulse">作弊</span>
                          </span>
                          <div class="flex items-center gap-2 mt-0.5 flex-wrap">
                            <span class="text-[7px] font-mono text-slate-400 dark:text-slate-500 uppercase tracking-tighter">UID: {{ player.uid }}</span>
                            <div class="flex items-center gap-1 bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded border border-slate-200 dark:border-white/5">
                              <LevelBadge :level="player.level || 1" :tier="player.tier" :tier-name="player.tier_name" size="xs" :show-level="false" />
                              <span class="text-[7px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-widest">Lv.{{ player.level || 1 }} {{ player.tier_name || '实习研究员' }}</span>
                            </div>
                            <span v-if="player.total_games > 0" class="text-[6px] font-bold text-slate-400/80 uppercase tracking-widest">
                              胜率: {{ Math.round((player.win_count / player.total_games) * 100) }}% ({{ player.total_games }}场)
                            </span>
                            <span :class="['text-[6px] font-bold uppercase tracking-widest', player.is_online ? 'text-emerald-500' : 'text-slate-400/80']">
                              {{ player.is_online ? '在线' : '上次下线 · ' + formatLastOfflineText(player.last_offline_at) }}
                            </span>
                          </div>
                        </div>
                      </div>
                    </td>
                    <td class="px-5 py-2">
                      <div class="flex items-center gap-1.5">
                        <span class="text-sm font-black text-amber-600 dark:text-amber-500 font-mono tracking-tighter">
                          {{ Math.floor(rankingMode === 'monthly' ? player.monthly_points : player.points) }}
                        </span>
                        <PhlogistonIcon :size="12" color="#f59e0b" />
                      </div>
                    </td>
                    <td class="px-5 py-2">
                      <div v-if="player.bounty > 0" class="flex items-center gap-1 text-rose-500">
                        <Flame class="w-2.5 h-2.5" />
                        <span class="text-xs font-black font-mono tracking-tighter">{{ player.bounty }}</span>
                      </div>
                      <div v-else class="text-[7px] font-bold text-slate-400 dark:text-slate-600 uppercase italic opacity-40 leading-none">—</div>
                    </td>
                    <td class="px-6 py-2 text-right">
                      <div v-if="Number(player.uid) !== Number(user.uid)" class="flex items-center justify-end gap-1.5">
                        <button v-if="player.is_online" @click="handleDuel(player)" title="单挑"
                          class="p-2 bg-blue-600/10 hover:bg-blue-600 text-blue-600 hover:text-white border border-blue-500/20 rounded-lg transition-all active:scale-95 shadow-sm">
                          <Swords class="w-3 h-3" />
                        </button>
                        <button @click="openBountyModal(player)" title="发布悬赏"
                          class="p-2 bg-rose-600/10 hover:bg-rose-600 text-rose-600 hover:text-white border border-rose-500/20 rounded-lg transition-all active:scale-95 shadow-sm">
                          <Crosshair class="w-3 h-3" />
                        </button>
                        <button v-if="!isFriend(player.uid)" @click="handleAddFriend(player)" title="加好友"
                          class="p-2 bg-amber-600/10 hover:bg-amber-600 text-amber-600 hover:text-white border border-amber-500/20 rounded-lg transition-all active:scale-95 shadow-sm">
                          <UserPlus class="w-3 h-3" />
                        </button>
                        <button v-else @click="startPrivateChat(player)" title="私信"
                          class="p-2 bg-emerald-600/10 hover:bg-emerald-600 text-emerald-600 hover:text-white border border-emerald-500/20 rounded-lg transition-all active:scale-95 shadow-sm">
                          <MessageCircle class="w-3 h-3" />
                        </button>
                      </div>
                      <div v-else class="text-[8px] font-black text-blue-600 dark:text-blue-500 uppercase tracking-widest italic pr-1">本人</div>
                    </td>
                  </tr>

                  <!-- 自我排名展示 (Desktop, 当不在前100名且未搜索时) -->
                  <tr
                    v-if="!searchTerm && myRankInfo && !leaderboard.find(p => Number(p.uid) === Number(user.uid))"
                    class="bg-blue-50/50 dark:bg-blue-500/[0.05] border-t-2 border-dashed border-slate-200 dark:border-white/10"
                  >
                    <td class="px-6 py-3">
                      <span class="w-6 h-6 rounded-lg flex items-center justify-center text-[10px] font-black italic bg-blue-100 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400">
                        {{ myRankInfo.rank }}
                      </span>
                    </td>
                    <td class="px-5 py-3">
                      <div class="flex items-center gap-3">
                        <div class="relative w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-500/20 border border-blue-200 dark:border-blue-500/30 flex items-center justify-center text-sm overflow-hidden shrink-0 shadow-inner">
                          <UserAvatar :avatar="myRankInfo.avatar" />
                          <div v-if="myRankInfo.is_online" class="absolute bottom-0 right-0 w-2 h-2 bg-emerald-500 border-2 border-white dark:border-[#121216] rounded-full"></div>
                        </div>
                        <div class="flex flex-col">
                          <span class="text-xs font-black text-blue-600 dark:text-blue-400 flex items-center gap-1.5 flex-wrap">
                            {{ myRankInfo.nickname || myRankInfo.username }}
                            <span class="text-[7px] bg-blue-600 px-1 py-0.5 rounded uppercase font-black tracking-widest text-white">我</span>
                          </span>
                          <div class="flex items-center gap-2 mt-0.5 flex-wrap">
                            <span class="text-[7px] font-mono text-blue-400/60 uppercase tracking-tighter">UID: {{ myRankInfo.uid }}</span>
                            <div class="flex items-center gap-1 bg-blue-500/10 dark:bg-blue-500/5 px-1.5 py-0.5 rounded border border-blue-500/20">
                              <LevelBadge :level="myRankInfo.level || 1" :tier="myRankInfo.tier" :tier-name="myRankInfo.tier_name" size="xs" :show-level="false" />
                              <span class="text-[7px] font-black text-blue-600 dark:text-blue-400 uppercase tracking-widest">Lv.{{ myRankInfo.level || 1 }} {{ myRankInfo.tier_name || '实习研究员' }}</span>
                            </div>
                            <span v-if="myRankInfo.total_games > 0" class="text-[6px] font-bold text-blue-400/60 uppercase tracking-widest">
                              胜率: {{ Math.round((myRankInfo.win_count / myRankInfo.total_games) * 100) }}% ({{ myRankInfo.total_games }}场)
                            </span>
                          </div>
                          <span class="text-[6px] font-bold text-blue-400/40 uppercase tracking-widest mt-0.5 ml-0.5">百名开外</span>
                        </div>
                      </div>
                    </td>
                    <td class="px-5 py-3">
                      <div class="flex items-center gap-1.5">
                        <span class="text-sm font-black text-blue-600 dark:text-blue-400 font-mono tracking-tighter">
                          {{ Math.floor(rankingMode === 'monthly' ? myRankInfo.monthly_points : myRankInfo.points) }}
                        </span>
                        <PhlogistonIcon :size="12" color="#2563eb" />
                      </div>
                    </td>
                    <td class="px-5 py-3">
                      <div v-if="myRankInfo.bounty > 0" class="flex items-center gap-1 text-rose-500">
                        <Flame class="w-2.5 h-2.5" />
                        <span class="text-xs font-black font-mono tracking-tighter">{{ myRankInfo.bounty }}</span>
                      </div>
                      <div v-else class="text-[7px] font-bold text-slate-400 dark:text-slate-600 uppercase italic opacity-40 leading-none">—</div>
                    </td>
                    <td class="px-6 py-3 text-right">
                      <span class="text-[8px] font-black text-blue-600 dark:text-blue-500 uppercase tracking-widest italic pr-1">已认证</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Load More Button -->
            <div v-if="!searchTerm && displayCount < allLeaderboardData.length" class="py-4 px-6 bg-slate-50/50 dark:bg-white/[0.02] border-t border-slate-100 dark:border-white/5 text-center">
              <button
                @click="loadMoreLeaderboard"
                :disabled="isLoadingMore"
                class="px-6 py-2 bg-blue-500/10 hover:bg-blue-500/20 disabled:opacity-50 text-blue-600 dark:text-blue-400 text-sm font-black rounded-lg transition-all"
              >
                <span v-if="isLoadingMore" class="flex items-center gap-2">
                  <Loader2 class="w-3 h-3 animate-spin" />
                  加载中...
                </span>
                <span v-else>加载更多 ({{ allLeaderboardData.length - displayCount }})</span>
              </button>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- Bounty Issue Modal -->
    <div v-if="showBountyModal" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
       <div class="absolute inset-0 bg-slate-900/40 dark:bg-black/80 backdrop-blur-md" @click="showBountyModal = false"></div>
       <div class="relative w-full max-w-sm bg-white dark:bg-[#121216] border border-slate-200 dark:border-rose-500/30 rounded-[32px] shadow-2xl overflow-hidden animate-in zoom-in duration-300">
          <div class="p-6 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-rose-500/5">
             <div class="flex items-center gap-3">
                <div class="w-10 h-10 bg-rose-500/10 border border-rose-500/20 rounded-xl flex items-center justify-center text-rose-500">
                   <Target class="w-5 h-5" />
                </div>
                <div>
                   <h2 class="text-lg font-black text-slate-900 dark:text-white uppercase tracking-tight">发布悬赏令</h2>
                   <p class="text-[8px] text-rose-500/60 font-mono uppercase tracking-widest mt-0.5">Configure_Target_Bounty</p>
                </div>
             </div>
             <button @click="showBountyModal = false" class="text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors">
                <X class="w-5 h-5" />
             </button>
          </div>
          
          <div class="p-8 space-y-6 text-center">
             <div class="flex flex-col items-center">
                <div class="w-16 h-16 rounded-2xl bg-slate-100 dark:bg-white/5 border border-slate-200 dark:border-white/10 flex items-center justify-center text-3xl mb-4 relative group overflow-hidden shadow-inner">
                   <UserAvatar :avatar="selectedTarget?.avatar" />
                   <div class="absolute -top-2 -right-2 w-6 h-6 bg-rose-600 rounded-full flex items-center justify-center border-4 border-white dark:border-[#121216]">
                      <Crosshair class="w-3 h-3 text-white" />
                   </div>
                </div>
                <p class="text-[9px] font-black uppercase tracking-widest text-slate-400 dark:text-slate-500">目标研究员</p>
                <h3 class="text-xl font-black text-slate-900 dark:text-white mt-0.5">{{ selectedTarget?.nickname || selectedTarget?.username }}</h3>
             </div>

             <div class="space-y-3">
                <div class="flex justify-between items-center px-1">
                   <label class="text-[9px] font-black text-slate-400 dark:text-slate-500 uppercase tracking-widest">投入科研燃素</label>
                   <span class="text-[9px] text-rose-600 dark:text-rose-500 font-mono font-bold">{{ userPoints }} 存储可用</span>
                </div>
                <div class="relative">
                   <input 
                      v-model="bountyAmount" 
                      type="number" 
                      placeholder="输入燃素数值..."
                      class="w-full h-12 bg-slate-50 dark:bg-black/40 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-white px-4 py-3 rounded-xl focus:ring-1 focus:ring-rose-500 outline-none transition-all font-mono text-base"
                   />
                </div>
                <p class="text-[8px] text-slate-500 leading-relaxed text-left px-1">
                  悬赏燃素将立刻扣除。任何人获胜均可平分，不可撤回。
                </p>
             </div>

             <div class="flex gap-3 pt-2">
                <button 
                   @click="showBountyModal = false"
                   class="flex-1 h-12 bg-slate-100 dark:bg-white/5 hover:bg-slate-200 dark:hover:bg-white/10 text-slate-500 font-bold rounded-xl transition-all uppercase tracking-widest text-[10px]"
                >
                   放弃
                </button>
                <button 
                   @click="handleCreateBounty"
                   :disabled="submitting || !bountyAmount || bountyAmount > userPoints"
                   class="flex-1 h-12 bg-rose-600 hover:bg-rose-500 text-white font-black rounded-xl transition-all shadow-lg disabled:grayscale flex items-center justify-center gap-2 group/btn"
                >
                   <template v-if="submitting">
                      <Loader2 class="w-4 h-4 animate-spin" />
                   </template>
                   <template v-else>
                      <Target class="w-3.5 h-3.5 group-hover:scale-125 transition-transform" />
                      <span class="uppercase tracking-widest text-[10px]">执行发布</span>
                   </template>
                </button>
             </div>
          </div>
       </div>
    </div>

    <!-- Floating Chat Toggle -->
    <button 
      @click="showChat = !showChat" 
      class="fixed bottom-6 right-6 z-50 w-14 h-14 bg-blue-600 hover:bg-blue-500 text-white rounded-[24px] shadow-2xl shadow-blue-500/30 flex items-center justify-center transition-all hover:scale-110 active:scale-95 group"
    >
      <MessageCircle class="w-6 h-6 group-hover:rotate-12 transition-transform" />
      <div v-if="hasNewMessage" class="absolute -top-1 -right-1 w-4 h-4 bg-rose-500 border-2 border-white dark:border-[#0a0a0c] rounded-full animate-pulse"></div>
    </button>

    <!-- Chat Sidebar/Modal -->
    <div 
      v-show="showChat"
      class="fixed bottom-24 right-6 z-50 w-[calc(100vw-3rem)] sm:w-[400px] shadow-2xl animate-in slide-in-from-bottom-10 duration-300 pointer-events-auto"
    >
      <ChatBox title="全球通信频率" maxHeight="500px" />
    </div>

    <!-- User Profile Modal -->
    <UserSpaceModal 
      :show="showProfileModal" 
      :uid="selectedProfileUID" 
      @close="showProfileModal = false" 
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { pointsAPI, gameAPI, friendAPI, authAPI } from '../utils/api'
import { useDialog } from '../utils/dialog'
import { Trophy, ArrowLeft, Loader2, Target, RefreshCw, ShieldCheck, Crosshair, Flame, X, Swords, MessageCircle, UserPlus, Search } from 'lucide-vue-next'
import PhlogistonIcon from '../components/icons/PhlogistonIcon.vue'
import { cn } from '../utils/cn'
import { formatLastOfflineForRanking } from '../utils/timeFormat'
import ChatBox from '../components/ChatBox.vue'
import LevelBadge from '../components/LevelBadge.vue'
import websocket from '../utils/websocket'
import UserSpaceModal from '../components/UserSpaceModal.vue'
import UserAvatar from '../components/UserAvatar.vue'

const router = useRouter()
const { showAlert, showPrompt } = useDialog()

let initialUser = {}
try {
  initialUser = JSON.parse(localStorage.getItem('user') || '{}')
} catch (e) {
  console.error('Failed to parse user in Ranking:', e)
}
interface User {
  uid?: number;
  username?: string;
  nickname?: string;
  avatar?: string;
  is_admin?: boolean;
}

const user = ref<User>(initialUser)

// Missing reactive variables
const selectedTarget = ref<any>(null)
const showBountyModal = ref(false)
const bountyAmount = ref(0)
const submitting = ref(false)

const leaderboard = ref<any[]>([])
const allLeaderboardData = ref<any[]>([])  // 完整排行榜数据
const displayCount = ref(50)  // 初始显示50条
const loading = ref(true)
const isLoadingMore = ref(false)
const myRankInfo = ref<any>(null)
const friendsList = ref<any[]>([])
const userPoints = ref(0)
const rankingMode = ref<'total' | 'monthly'>('total')

const searchTerm = ref('')
const searchResults = ref<any[]>([])
const isSearching = ref(false)

// 个人空间弹窗相关
const showProfileModal = ref(false)
const selectedProfileUID = ref<number | null>(null)

const showResearcherProfile = (uid: number) => {
  selectedProfileUID.value = uid
  showProfileModal.value = true
}

const formatLastOfflineText = (value: string | Date | null | undefined) => {
  return formatLastOfflineForRanking(value)
}

// 监听搜索词
let searchTimeout: any = null
watch(searchTerm, (newVal) => {
  if (searchTimeout) clearTimeout(searchTimeout)
  if (!newVal.trim()) {
    searchResults.value = []
    isSearching.value = false
    return
  }
  
  isSearching.value = true
  searchTimeout = setTimeout(async () => {
    try {
      const res = await authAPI.searchUsers(newVal)
      searchResults.value = res.data || []
    } catch (err) {
      console.error('搜索失败:', err)
      searchResults.value = []
    } finally {
      isSearching.value = false
    }
  }, 500)
})

const isFriend = (uid: number) => {
  return friendsList.value.some(f => Number(f.uid) === Number(uid))
}

const handleAddFriend = async (player: any) => {
  const message = await showPrompt('请输入申请信息（可选）:', '你好，我想和你一起进行化学实验。', '发送好友请求')
  if (message === null) return

  try {
    const displayName = player.nickname || player.username
    await friendAPI.sendRequest(player.uid, message)
    showAlert(`已向研究员 ${displayName} 发送同步请求，等待量子握手。`, '请求已发送')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '请求发送失败', '链路故障')
  }
}

const showChat = ref(false)
const hasNewMessage = ref(false)

const startPrivateChat = (player: any) => {
  if (!isFriend(player.uid)) {
    showAlert('只有互为好友的研究员才能开启单向加密传输。', '权限受限')
    return
  }
  showChat.value = true
  hasNewMessage.value = false
  const displayName = player.nickname || player.username
  nextTick(() => {
    window.dispatchEvent(new CustomEvent('start-private-chat', {
      detail: { uid: player.uid, username: displayName, nickname: player.nickname }
    }))
  })
}

const loadLeaderboard = async () => {
  try {
    loading.value = true
    displayCount.value = 50  // 重置为初始50条
    const [leaderRes, friendsRes] = await Promise.all([
      pointsAPI.getLeaderboard(rankingMode.value),
      friendAPI.getFriends()
    ])
    
    console.log('Leaderboard Raw Response Data:', leaderRes.data)
    
    allLeaderboardData.value = leaderRes.data.leaderboard || []
    leaderboard.value = allLeaderboardData.value.slice(0, 50)  // 初始只显示50条
    myRankInfo.value = leaderRes.data.self || null
    friendsList.value = friendsRes.data || []
    
    // 同时也尝试更新一下本地的用户分数实时显示
    const currentUid = user.value.uid
    const self = allLeaderboardData.value.find(p => Number(p.uid) === Number(currentUid)) || myRankInfo.value
    if (self) {
      userPoints.value = rankingMode.value === 'monthly' ? self.monthly_points : self.points
    }
  } catch (error) {
    console.error('Failed to load ranking data:', error)
  } finally {
    loading.value = false
  }
}

// 加载更多排行榜项
const loadMoreLeaderboard = () => {
  if (isLoadingMore.value || displayCount.value >= allLeaderboardData.value.length) return
  isLoadingMore.value = true
  
  // 模拟异步加载（实际上是同步的，但给用户Loading反馈）
  setTimeout(() => {
    const newCount = Math.min(displayCount.value + 50, allLeaderboardData.value.length)
    leaderboard.value = allLeaderboardData.value.slice(0, newCount)
    displayCount.value = newCount
    isLoadingMore.value = false
  }, 200)
}

watch(rankingMode, () => {
  loadLeaderboard()
})

const openBountyModal = (player: any) => {
  selectedTarget.value = player
  bountyAmount.value = 100
  showBountyModal.value = true
}

const handleCreateBounty = async () => {
  if (!selectedTarget.value || !bountyAmount.value) return
  if (bountyAmount.value > userPoints.value) {
    showAlert('科研燃素余额不足，无法发起此项悬赏', '核心功率受限')
    return
  }

  try {
    submitting.value = true
    await pointsAPI.createBounty(selectedTarget.value.uid, bountyAmount.value)
    const displayName = selectedTarget.value.nickname || selectedTarget.value.username
    showAlert(`已成功对研究员 ${displayName} 发布悬赏。`, '目标已锁定')
    showBountyModal.value = false
    loadLeaderboard() // 刷新列表
  } catch (error: any) {
    showAlert(error.response?.data?.error || '发布悬赏失败', '系统通讯故障')
  } finally {
    submitting.value = false
  }
}

const handleDuel = async (player: any) => {
  try {
    await gameAPI.initiateDuel(player.uid)
    // 后端会通过 WebSocket 广播 duel_start，这里只需提示
    const displayName = player.nickname || player.username
    showAlert(`已向 ${displayName} 发起单挑协议，正在建立量子隧道...`, '协议启动')
  } catch (error: any) {
    showAlert(error.response?.data?.error || '发起单挑失败', '系统通讯故障')
  }
}

const onChatMessage = () => {
  if (!showChat.value) {
    hasNewMessage.value = true
  }
}

onMounted(() => {
  loadLeaderboard()
  websocket.on('chat', onChatMessage)
  websocket.on('private_chat', onChatMessage)
})

onUnmounted(() => {
  websocket.off('chat', onChatMessage)
  websocket.off('private_chat', onChatMessage)
})
</script>

<style src="./Ranking.css" scoped></style>

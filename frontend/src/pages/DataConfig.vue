<template>
  <div class="min-h-screen bg-slate-50 dark:bg-[#0a0a0c] text-slate-900 dark:text-slate-200 p-4 lg:p-6 font-sans selection:bg-blue-500/30">
    <div class="fixed inset-0 overflow-hidden pointer-events-none">
      <div class="absolute top-[-10%] left-[-10%] w-[50%] h-[50%] bg-blue-500/5 rounded-full blur-[120px]" />
      <div class="absolute bottom-[-10%] right-[-10%] w-[50%] h-[50%] bg-emerald-500/5 rounded-full blur-[120px]" />
      <div class="absolute inset-0 bg-[url('/noise.svg')] opacity-20 brightness-50 contrast-150" />
    </div>

    <div class="max-w-6xl mx-auto relative z-10">
      <!-- 头部 -->
      <header class="flex flex-col lg:flex-row items-center justify-between gap-6 mb-8">
        <div class="flex items-center gap-6">
          <div class="relative group">
            <div class="absolute inset-x-0 inset-y-0 bg-emerald-500 blur-2xl opacity-20 group-hover:opacity-40 transition-opacity" />
            <div class="w-12 h-12 rounded-xl bg-white dark:bg-[#111114] border border-slate-200 dark:border-emerald-500/40 flex items-center justify-center relative z-10 shadow-2xl">
              <Database class="w-6 h-6 text-emerald-600 dark:text-emerald-400 group-hover:scale-110 transition-transform" />
            </div>
          </div>
          <div>
            <h1 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase flex items-center gap-3">
              Data Hub <span class="text-[8px] font-mono bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 px-1.5 py-0.5 rounded border border-emerald-500/30 not-italic">{{ userRole }}</span>
            </h1>
            <p class="text-slate-400 dark:text-slate-500 text-[10px] font-bold tracking-widest uppercase mt-0.5">
              核心数据配置与管理中心 / Core configuration center
            </p>
          </div>
        </div>

        <div class="flex items-center gap-4">
          <button 
            @click="router.push('/')"
            class="px-5 py-2 bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/5 hover:border-emerald-500/30 rounded-xl flex items-center gap-3 shadow-xl transition-all group"
          >
            <ArrowLeft class="w-4 h-4 text-emerald-600 dark:text-emerald-400 group-hover:-translate-x-1 transition-transform" />
            <span class="text-[10px] font-black text-slate-500 group-hover:text-slate-900 dark:group-hover:text-white uppercase tracking-widest">返回实验室</span>
          </button>
        </div>
      </header>

      <!-- 导航卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <!-- 反应方程式 -->
        <router-link 
          to="/data/reactions" 
          class="group relative overflow-hidden bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/5 p-6 rounded-2xl shadow-2xl hover:border-emerald-500/30 transition-all duration-500"
        >
          <div class="absolute -top-10 -right-10 opacity-5 group-hover:opacity-10 transition-opacity">
            <FlaskConical class="w-48 h-48 text-emerald-500 italic" />
          </div>
          <div class="relative z-10 space-y-4">
            <div class="flex items-center gap-3 text-emerald-600 dark:text-emerald-500">
              <span class="text-[8px] font-black uppercase tracking-[0.3em] px-2 py-0.5 bg-emerald-500/10 border border-emerald-500/20 rounded-full">Module 01</span>
              <div class="h-[1px] flex-1 bg-emerald-500/10"></div>
            </div>
            <h2 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase">反应方程式库</h2>
            <p class="text-slate-500 dark:text-slate-500 text-xs leading-relaxed font-medium">
              查看或管理游戏中允许的所有化学反应。审核、编辑或提交新的反应方程，这是游戏规则的核心。
            </p>
            <div class="pt-4 flex items-center gap-2 group-hover:translate-x-3 transition-transform duration-500 text-emerald-600 dark:text-emerald-400">
              <span class="text-[10px] font-black uppercase tracking-widest">进入配置界面</span>
              <ArrowRight class="w-4 h-4" />
            </div>
          </div>
        </router-link>

        <!-- 物质百科 -->
        <router-link 
          to="/data/substances" 
          class="group relative overflow-hidden bg-white dark:bg-[#111114] border border-slate-200 dark:border-white/5 p-6 rounded-2xl shadow-2xl hover:border-blue-500/30 transition-all duration-500"
        >
          <div class="absolute -top-10 -right-10 opacity-5 group-hover:opacity-10 transition-opacity">
            <Beaker class="w-48 h-48 text-blue-500 italic" />
          </div>
          <div class="relative z-10 space-y-4">
            <div class="flex items-center gap-3 text-blue-600 dark:text-blue-500">
              <span class="text-[8px] font-black uppercase tracking-[0.3em] px-2 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-full">Module 02</span>
              <div class="h-[1px] flex-1 bg-blue-500/10"></div>
            </div>
            <h2 class="text-xl font-black text-slate-900 dark:text-white italic tracking-tighter uppercase">物质百科全书</h2>
            <p class="text-slate-500 dark:text-slate-500 text-xs leading-relaxed font-medium">
              查阅或定义化学物质。你可以提交新发现的物质化学式，系统会自动分析涉及的元素需求。
            </p>
            <div class="pt-4 flex items-center gap-2 group-hover:translate-x-3 transition-transform duration-500 text-blue-600 dark:text-blue-400">
              <span class="text-[10px] font-black uppercase tracking-widest">进入配置界面</span>
              <ArrowRight class="w-4 h-4" />
            </div>
          </div>
        </router-link>
      </div>

      <!-- Footer -->
      <footer class="mt-10 pt-5 border-t border-slate-200 dark:border-white/5 flex items-center justify-between opacity-40">
        <div class="flex items-center gap-4 text-[8px] font-black uppercase tracking-[0.4em] text-slate-500">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          Data Sync Active
        </div>
        <div class="text-[8px] font-black uppercase tracking-[0.4em] text-slate-500">
          Experimental Phase 0.1.4
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { computed } from 'vue'
import { Database, ArrowLeft, ArrowRight, FlaskConical, Beaker } from 'lucide-vue-next'

const router = useRouter()
let user: any = {}
try {
  user = JSON.parse(localStorage.getItem('user') || '{}')
} catch (e) {
  console.error('Failed to parse user in DataConfig:', e)
}

const userRole = computed(() => {
  const role = user.role?.toUpperCase() || 'USER'
  if (role === 'ADMIN' || role === 'CO-WORKER') return `WIKI_CONSOLE [${role}]`
  return `WIKI_READER`
})
</script>
<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Plus, Minus, Trash2, Edit2, Hexagon, Save, X } from 'lucide-vue-next'
import { gameAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const decks = ref<any[]>([])
const isLoading = ref(true)
const { showAlert, showConfirm } = useDialog()

const editingDeck = ref<any>(null)
const newDeckName = ref('')
const selectedElements = ref<Record<string, number>>({})
const initialHandSize = ref(10)

const BUILTIN_DEFAULT_CARDS: Record<string, number> = {
  H: 12,
  O: 12,
  C: 4, N: 4, F: 4, Na: 4, Mg: 4, Al: 4,
  Si: 4, P: 4, S: 4, Cl: 4, K: 4, Ca: 4,
  Mn: 4, Fe: 4, Cu: 4, Zn: 4, Br: 4, I: 4, Ag: 4,
  '+2': 8, '+4': 4,
  He: 1, Ne: 1, Ar: 1, Kr: 1,
  Au: 4
}

// 可选元素列表
const ALL_ELEMENTS = [
  'H', 'He', 'C', 'N', 'O', 'F', 'Ne',
  'Na', 'Mg', 'Al', 'Si', 'P', 'S', 'Cl', 'Ar', 'K', 'Ca',
  'Fe', 'Cu', 'Zn', 'Ag', 'Au', 'Kr', 'I', 'Br', 'Mn'
]

const loadDecks = async () => {
  isLoading.value = true
  try {
    const res = await gameAPI.getMyDecks()
    const allDecks = res.data || []
    // 将全局牌组排在首位
    allDecks.sort((a: any, b: any) => {
      if (a.is_global && !b.is_global) return -1
      if (!a.is_global && b.is_global) return 1
      return 0
    })
    decks.value = allDecks
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

const openEdit = (deck: any) => {
  if (deck.is_global) return // 全局牌组只读
  editingDeck.value = deck
  newDeckName.value = deck.name
  selectedElements.value = { ...deck.cards }
  initialHandSize.value = deck.initial_cards || 10
}

const openCreate = () => {
  editingDeck.value = { id: 0, name: '' }
  newDeckName.value = '新实验方案'
  // 默认尝试使用系统牌组的内容作为模板
  const globalDeck = decks.value.find(d => d.is_global)
  if (globalDeck) {
    selectedElements.value = { ...globalDeck.cards }
    initialHandSize.value = globalDeck.initial_cards || 10
  } else {
    selectedElements.value = { ...BUILTIN_DEFAULT_CARDS }
    initialHandSize.value = 10
  }
}

const copyDeck = (deck: any) => {
  editingDeck.value = { id: 0, name: '' }
  newDeckName.value = `${deck.name} (副本)`
  selectedElements.value = { ...deck.cards }
  initialHandSize.value = deck.initial_cards || 10
}

const toggleElement = (el: string) => {
  if (selectedElements.value[el]) {
    delete selectedElements.value[el]
  } else {
    selectedElements.value[el] = 4
  }
}

const saveDeck = async () => {
  if (!newDeckName.value.trim() || (editingDeck.value?.is_global)) return
  
  try {
    if (editingDeck.value.id === 0) {
      await gameAPI.createMyDeck(newDeckName.value, selectedElements.value, initialHandSize.value)
    } else {
      await gameAPI.updateMyDeck(editingDeck.value.id, newDeckName.value, selectedElements.value, initialHandSize.value)
    }
    showAlert('卡组保存成功', '成功')
    editingDeck.value = null
    loadDecks()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '保存失败', '出错了')
  }
}

const deleteDeck = async (id: number) => {
  const confirmed = await showConfirm('确定要删除此卡组吗？', '删除卡组')
  if (!confirmed) return
  try {
    await gameAPI.deleteMyDeck(id)
    loadDecks()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '删除失败', '出错了')
  }
}

// 监控弹窗状态以禁用/启用背景滚动
watch(
  editingDeck,
  (deck) => {
    if (deck) {
      document.documentElement.style.overflow = 'hidden'
      document.body.style.overflow = 'hidden'
    } else {
      document.documentElement.style.overflow = ''
      document.body.style.overflow = ''
    }
  }
)

onMounted(loadDecks)
</script>

<template>
  <div class="space-y-5">
    <div class="flex items-center justify-between">
       <h3 class="text-base font-black italic tracking-tighter uppercase text-slate-900 dark:text-white">Deck Library <span class="text-slate-400 dark:text-white/20 text-[10px] lowercase font-mono not-italic ml-2">/ default + custom</span></h3>
       <button 
        @click="openCreate"
        class="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-black uppercase tracking-widest transition-all"
       >
         <Plus class="w-4 h-4" /> New Sequence
       </button>
    </div>

    <div v-if="isLoading" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div v-for="i in 2" :key="i" class="h-28 bg-slate-100 dark:bg-white/5 rounded-2xl animate-pulse"></div>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div v-for="deck in decks" :key="deck.id" 
        class="p-4 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/5 rounded-2xl group relative overflow-hidden"
      >
        <div class="absolute top-0 left-0 w-1 h-full bg-blue-500 transition-opacity"></div>
        
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 bg-blue-500/10 rounded-lg flex items-center justify-center">
              <Hexagon class="w-4 h-4 text-blue-500" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h4 class="font-black text-sm uppercase tracking-tight">{{ deck.name }}</h4>
                <span v-if="deck.is_global" class="px-1.5 py-0.5 bg-blue-500/20 text-blue-500 text-[8px] font-black rounded uppercase">System</span>
              </div>
              <p class="text-[10px] text-slate-500 font-bold uppercase tracking-widest">{{ Object.keys(deck.cards).length }} UNIQUE ELEMENTS</p>
            </div>
          </div>
          
          <div class="flex items-center gap-1 transition-opacity">
            <button v-if="!deck.is_global" @click="openEdit(deck)" class="p-1.5 text-slate-400 hover:text-blue-500 hover:bg-blue-500/10 rounded-lg transition-all" title="编辑序列">
              <Edit2 class="w-4 h-4" />
            </button>
            <button @click="copyDeck(deck)" class="p-1.5 text-slate-400 hover:text-emerald-500 hover:bg-emerald-500/10 rounded-lg transition-all" title="克隆为模板">
              <Plus class="w-4 h-4" />
            </button>
            <button v-if="!deck.is_global" @click="deleteDeck(deck.id)" class="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-500/10 rounded-lg transition-all">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div class="flex flex-wrap gap-1.5 opacity-60">
           <span v-for="(count, el) in deck.cards" :key="el" class="px-2 py-0.5 bg-slate-100 dark:bg-white/5 rounded text-[8px] font-mono font-bold">
             {{ el }}:{{ count }}
           </span>
        </div>
        
        <div v-if="deck.is_global" class="absolute top-4 right-4 text-[8px] font-black uppercase bg-amber-500/20 text-amber-500 px-2 py-1 rounded-md border border-amber-500/20">
          Global Stable
        </div>
      </div>
    </div>

    <!-- Edit Modal Overlay -->
     <Teleport to="body">
     <div v-if="editingDeck" class="viewport-modal-overlay z-[100] p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md" @click="editingDeck = null">
       <div class="relative w-full max-w-2xl bg-white dark:bg-slate-900 rounded-2xl border border-white/10 shadow-3xl overflow-hidden flex flex-col max-h-[90vh] pointer-events-auto" @click.stop>
          <div class="p-5 border-b border-white/5">
             <div class="flex items-center justify-between">
                <h3 class="text-lg font-black italic tracking-tighter uppercase">Deck Configuration</h3>
                <button @click="editingDeck = null" class="p-2 hover:bg-white/5 rounded-full"><X class="w-5 h-5" /></button>
             </div>
          </div>

          <div class="p-5 overflow-y-auto space-y-5">
             <div>
                <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Sequence Name</label>
                <input 
                  v-model="newDeckName"
                  type="text" 
                  placeholder="EXPERIMENTAL DECK LEGACY"
                  class="w-full bg-slate-100 dark:bg-white/5 border border-white/10 rounded-xl px-4 py-3 text-base font-black uppercase tracking-tight focus:border-blue-500 transition-all outline-none"
                />
             </div>

             <div>
                <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Initial Hand Size (Default 10)</label>
                <div class="flex items-center gap-4">
                  <div class="flex-1 flex items-center bg-slate-100 dark:bg-white/5 border border-white/10 rounded-xl overflow-hidden px-4">
                    <button 
                      @click="initialHandSize = Math.max(1, initialHandSize - 1)"
                      class="p-2 text-blue-500 hover:bg-white/5 transition-colors"
                    >
                      <Minus class="w-4 h-4" />
                    </button>
                    <input 
                      v-model.number="initialHandSize"
                      type="number"
                      min="1"
                      max="40"
                      class="w-full bg-transparent border-none text-center py-2 text-base font-black focus:outline-none [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
                    />
                    <button 
                      @click="initialHandSize = Math.min(40, initialHandSize + 1)"
                      class="p-2 text-blue-500 hover:bg-white/5 transition-colors"
                    >
                      <Plus class="w-4 h-4" />
                    </button>
                  </div>
                  <div class="text-[10px] font-black text-slate-400 uppercase w-20 text-center">
                    Cards
                  </div>
                </div>
             </div>

             <div>
                <div class="flex items-center justify-between mb-4">
                   <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">Element Matrix</label>
                   <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">{{ Object.keys(selectedElements).length }} Active Nodes</span>
                </div>
                
                <div class="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-8 gap-2">
                   <button 
                    v-for="el in ALL_ELEMENTS" 
                    :key="el"
                    @click="toggleElement(el)"
                    :class="[
                      'h-10 rounded-lg font-mono font-black text-xs flex items-center justify-center transition-all border-2',
                      selectedElements[el] 
                        ? 'bg-blue-600 border-blue-400 text-white shadow-lg shadow-blue-500/40' 
                        : 'bg-white/5 border-white/5 text-slate-500 grayscale hover:grayscale-0 hover:bg-white/10'
                    ]"
                   >
                     {{ el }}
                   </button>
                   <!-- Special Cards -->
                   <button 
                    v-for="spec in ['+2', '+4']" 
                    :key="spec"
                    @click="toggleElement(spec)"
                    :class="[
                      'h-10 rounded-lg font-mono font-black text-xs flex items-center justify-center transition-all border-2',
                      selectedElements[spec] 
                        ? 'bg-rose-600 border-rose-400 text-white shadow-lg shadow-rose-500/40' 
                        : 'bg-white/5 border-white/5 text-slate-500 grayscale hover:grayscale-0 hover:bg-white/10'
                    ]"
                   >
                     {{ spec }}
                   </button>
                </div>
             </div>

             <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div v-for="(count, el) in selectedElements" :key="el" class="flex items-center justify-between p-2 bg-white/5 rounded-xl border border-white/5">
                   <span class="font-mono font-black text-blue-400 uppercase tracking-widest text-xs">{{ el }}</span>
                   <div class="flex items-center gap-3">
                      <button @click="selectedElements[el] = Math.max(1, count - 1)" class="w-6 h-6 flex items-center justify-center rounded bg-white/10 hover:bg-white/20">-</button>
                      <span class="font-mono font-black text-sm w-4 text-center">{{ count }}</span>
                      <button @click="selectedElements[el] = Math.min(64, count + 1)" class="w-6 h-6 flex items-center justify-center rounded bg-white/10 hover:bg-white/20">+</button>
                   </div>
                </div>
             </div>
          </div>

          <div class="p-5 bg-black/20 border-t border-white/5">
             <button 
              @click="saveDeck"
              :disabled="!newDeckName.trim()"
              class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white py-3 rounded-xl font-black uppercase tracking-widest transition-all flex items-center justify-center gap-3"
             >
               <Save class="w-5 h-5" /> Save Configuration
             </button>
          </div>
       </div>
    </div>
     </Teleport>
  </div>
</template>

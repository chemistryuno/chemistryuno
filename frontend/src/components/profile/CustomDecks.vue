<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Plus, Trash2, Edit2, Hexagon, Save, AlertCircle, X } from 'lucide-vue-next'
import { gameAPI } from '../../utils/api'
import { useDialog } from '../../utils/dialog'

const decks = ref<any[]>([])
const isLoading = ref(true)
const { showAlert } = useDialog()

const editingDeck = ref<any>(null)
const newDeckName = ref('')
const selectedElements = ref<Record<string, number>>({})

// 可选元素列表
const ALL_ELEMENTS = [
  'H', 'He', 'Li', 'Be', 'B', 'C', 'N', 'O', 'F', 'Ne',
  'Na', 'Mg', 'Al', 'Si', 'P', 'S', 'Cl', 'Ar', 'K', 'Ca',
  'Fe', 'Cu', 'Zn', 'Ag', 'Au', 'Kr', 'I', 'Br', 'Mn'
]

const loadDecks = async () => {
  isLoading.value = true
  try {
    const res = await gameAPI.getMyDecks()
    decks.value = res.data
  } catch (e) {
    console.error(e)
  } finally {
    isLoading.value = false
  }
}

const openCreate = () => {
  editingDeck.value = { id: 0, name: '' }
  newDeckName.value = ''
  selectedElements.value = { 'H': 12, 'O': 12, 'C': 4, '+2': 8, '+4': 4 }
}

const toggleElement = (el: string) => {
  if (selectedElements.value[el]) {
    delete selectedElements.value[el]
  } else {
    selectedElements.value[el] = 4
  }
}

const saveDeck = async () => {
  if (!newDeckName.value.trim()) return
  
  try {
    if (editingDeck.value.id === 0) {
      await gameAPI.createMyDeck(newDeckName.value, selectedElements.value)
    } else {
      await gameAPI.updateMyDeck(editingDeck.value.id, newDeckName.value, selectedElements.value)
    }
    showAlert('卡组保存成功', '成功')
    editingDeck.value = null
    loadDecks()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '保存失败', '出错了')
  }
}

const deleteDeck = async (id: number) => {
  if (!confirm('确定要删除此卡组吗？')) return
  try {
    await gameAPI.deleteMyDeck(id)
    loadDecks()
  } catch (e: any) {
    showAlert(e.response?.data?.error || '删除失败', '出错了')
  }
}

onMounted(loadDecks)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
       <h3 class="text-xl font-black italic tracking-tighter uppercase text-slate-900 dark:text-white">Custom Decks</h3>
       <button 
        @click="openCreate"
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-black uppercase tracking-widest transition-all"
       >
         <Plus class="w-4 h-4" /> New Sequence
       </button>
    </div>

    <div v-if="isLoading" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div v-for="i in 2" :key="i" class="h-32 bg-slate-100 dark:bg-white/5 rounded-3xl animate-pulse"></div>
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <div v-for="deck in decks" :key="deck.id" 
        class="p-6 bg-white dark:bg-black/40 border border-slate-200 dark:border-white/5 rounded-3xl group relative overflow-hidden"
      >
        <div class="absolute top-0 left-0 w-1 h-full bg-blue-500 opacity-0 group-hover:opacity-100 transition-opacity"></div>
        
        <div class="flex items-center justify-between mb-4">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center">
              <Hexagon class="w-5 h-5 text-blue-500" />
            </div>
            <div>
              <h4 class="font-black text-sm uppercase tracking-tight">{{ deck.name }}</h4>
              <p class="text-[10px] text-slate-500 font-bold uppercase tracking-widest">{{ Object.keys(deck.cards).length }} UNIQUE ELEMENTS</p>
            </div>
          </div>
          
          <div v-if="!deck.is_global" class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
            <button @click="deleteDeck(deck.id)" class="p-2 text-slate-400 hover:text-red-500 hover:bg-red-500/10 rounded-lg transition-all">
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
    <div v-if="editingDeck" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
       <div class="absolute inset-0 bg-black/80 backdrop-blur-md" @click="editingDeck = null"></div>
       
       <div class="relative w-full max-w-2xl bg-white dark:bg-slate-900 rounded-[2.5rem] border border-white/10 shadow-3xl overflow-hidden flex flex-col max-h-[90vh]">
          <div class="p-8 border-b border-white/5">
             <div class="flex items-center justify-between">
                <h3 class="text-2xl font-black italic tracking-tighter uppercase">Deck Configuration</h3>
                <button @click="editingDeck = null" class="p-2 hover:bg-white/5 rounded-full"><X class="w-6 h-6" /></button>
             </div>
          </div>

          <div class="p-8 overflow-y-auto space-y-8">
             <div>
                <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest mb-2 block">Sequence Name</label>
                <input 
                  v-model="newDeckName"
                  type="text" 
                  placeholder="EXPERIMENTAL DECK ALPHA"
                  class="w-full bg-slate-100 dark:bg-white/5 border border-white/10 rounded-2xl px-6 py-4 text-lg font-black uppercase tracking-tight focus:border-blue-500 transition-all outline-none"
                />
             </div>

             <div>
                <div class="flex items-center justify-between mb-4">
                   <label class="text-[10px] font-black text-slate-500 uppercase tracking-widest block">Element Matrix</label>
                   <span class="text-[10px] font-black text-blue-500 uppercase tracking-widest">{{ Object.keys(selectedElements).length }} Active Nodes</span>
                </div>
                
                <div class="grid grid-cols-4 sm:grid-cols-6 lg:grid-cols-8 gap-3">
                   <button 
                    v-for="el in ALL_ELEMENTS" 
                    :key="el"
                    @click="toggleElement(el)"
                    :class="[
                      'h-12 rounded-xl font-mono font-black text-sm flex items-center justify-center transition-all border-2',
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
                      'h-12 rounded-xl font-mono font-black text-sm flex items-center justify-center transition-all border-2',
                      selectedElements[spec] 
                        ? 'bg-rose-600 border-rose-400 text-white shadow-lg shadow-rose-500/40' 
                        : 'bg-white/5 border-white/5 text-slate-500 grayscale hover:grayscale-0 hover:bg-white/10'
                    ]"
                   >
                     {{ spec }}
                   </button>
                </div>
             </div>

             <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div v-for="(count, el) in selectedElements" :key="el" class="flex items-center justify-between p-3 bg-white/5 rounded-2xl border border-white/5">
                   <span class="font-mono font-black text-blue-400 uppercase tracking-widest text-xs">{{ el }}</span>
                   <div class="flex items-center gap-3">
                      <button @click="selectedElements[el] = Math.max(1, count - 1)" class="w-6 h-6 flex items-center justify-center rounded bg-white/10 hover:bg-white/20">-</button>
                      <span class="font-mono font-black text-sm w-4 text-center">{{ count }}</span>
                      <button @click="selectedElements[el] = Math.min(64, count + 1)" class="w-6 h-6 flex items-center justify-center rounded bg-white/10 hover:bg-white/20">+</button>
                   </div>
                </div>
             </div>
          </div>

          <div class="p-8 bg-black/20 border-t border-white/5">
             <button 
              @click="saveDeck"
              :disabled="!newDeckName.trim()"
              class="w-full bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white py-4 rounded-2xl font-black uppercase tracking-widest transition-all flex items-center justify-center gap-3"
             >
               <Save class="w-5 h-5" /> Save Configuration
             </button>
          </div>
       </div>
    </div>
  </div>
</template>

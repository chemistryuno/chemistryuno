<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Plus, Trash2, FlaskConical, ChevronDown } from 'lucide-vue-next'

const props = defineProps<{
  modelValue?: string
}>()

const emit = defineEmits(['update:modelValue', 'change'])

const editor = ref({
  reactants: [{ coefficient: '', formula: '' }],
  products: [{ coefficient: '', formula: '' }],
  arrow: '='
})

const formatList = (list: { coefficient: string, formula: string }[]) => {
  return list
    .filter(item => item.formula.trim())
    .map(item => {
      const coef = item.coefficient.trim()
      const formula = item.formula.trim()
      return (coef && coef !== '1') ? `${coef}${formula}` : formula
    })
    .join(' + ')
}

const generatedDisplay = computed(() => {
  const left = formatList(editor.value.reactants)
  const right = formatList(editor.value.products)
  if (!left || !right) return ''
  return `${left} ${editor.value.arrow} ${right}`
})

watch(generatedDisplay, (val) => {
  emit('update:modelValue', val)
  emit('change', val)
})

const addSubstance = (type: 'reactants' | 'products') => {
  editor.value[type].push({ coefficient: '', formula: '' })
}

const removeSubstance = (type: 'reactants' | 'products', index: number) => {
  if (editor.value[type].length > 1) {
    editor.value[type].splice(index, 1)
  } else {
    editor.value[type][index] = { coefficient: '', formula: '' }
  }
}

const reset = () => {
  editor.value = {
    reactants: [{ coefficient: '', formula: '' }],
    products: [{ coefficient: '', formula: '' }],
    arrow: '='
  }
}

defineExpose({ reset })
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-col gap-4">
      <!-- 反应物区域 -->
      <div class="space-y-3">
        <div v-for="(item, index) in editor.reactants" :key="'r-'+index" class="flex items-center gap-2 p-2 bg-blue-500/5 dark:bg-blue-500/10 rounded-xl border border-blue-500/10 group/item">
          <input 
            v-model="item.coefficient"
            type="text" 
            placeholder="1"
            class="w-10 px-1 py-1 bg-black/20 border border-blue-400/10 rounded-lg text-center text-blue-500 font-black text-xs placeholder:text-slate-600 outline-none focus:border-blue-500/40 transition-all"
          />
          <input 
            v-model="item.formula"
            type="text" 
            placeholder="H2"
            class="flex-1 px-2 py-1 bg-black/20 border border-blue-400/10 rounded-lg text-white font-mono text-xs placeholder:text-slate-600 outline-none focus:border-blue-500/40 transition-all"
          />
          <button @click.prevent="removeSubstance('reactants', index)" class="p-1.5 text-slate-500 hover:text-red-400 opacity-0 group-hover/item:opacity-100 transition-all">
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>
        <button 
          @click.prevent="addSubstance('reactants')"
          class="w-full py-2 bg-blue-500/5 hover:bg-blue-500/10 border border-dashed border-blue-500/30 rounded-xl text-[9px] font-black text-blue-500 flex items-center justify-center gap-1 uppercase tracking-widest transition-all"
        >
          <Plus class="w-3 h-3" /> Reactant
        </button>
      </div>

      <!-- 箭头/等号选择 -->
      <div class="flex items-center justify-center gap-2 py-1">
        <div class="h-px flex-1 bg-gradient-to-r from-transparent to-blue-500/10" />
        <div class="relative group">
          <select v-model="editor.arrow" 
            class="appearance-none bg-slate-900/80 border border-blue-400/30 rounded-lg pl-8 pr-8 py-1.5 text-blue-400 font-black text-xs outline-none transition-all cursor-pointer hover:bg-blue-500/10 focus:ring-1 focus:ring-blue-500/50 min-w-[80px] text-center"
          >
            <option value="=" class="bg-slate-900">=</option>
            <option value="→" class="bg-slate-900">→</option>
            <option value="⇌" class="bg-slate-900">⇌</option>
          </select>
          <FlaskConical class="w-3 h-3 text-blue-500/40 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none group-hover:text-blue-500/60 transition-colors" />
          <ChevronDown class="w-3 h-3 text-blue-500/40 absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none group-hover:text-blue-500/60 transition-colors" />
        </div>
        <div class="h-px flex-1 bg-gradient-to-l from-transparent to-blue-500/10" />
      </div>

      <!-- 生成物区域 -->
      <div class="space-y-3">
        <div v-for="(item, index) in editor.products" :key="'p-'+index" class="flex items-center gap-2 p-2 bg-green-500/5 dark:bg-green-500/10 rounded-xl border border-green-500/10 group/item">
          <input 
            v-model="item.coefficient"
            type="text" 
            placeholder="1"
            class="w-10 px-1 py-1 bg-black/20 border border-green-400/10 rounded-lg text-center text-green-500 font-black text-xs placeholder:text-slate-600 outline-none focus:border-green-500/40 transition-all"
          />
          <input 
            v-model="item.formula"
            type="text" 
            placeholder="H2O"
            class="flex-1 px-2 py-1 bg-black/20 border border-green-400/10 rounded-lg text-white font-mono text-xs placeholder:text-slate-600 outline-none focus:border-green-500/40 transition-all"
          />
          <button @click.prevent="removeSubstance('products', index)" class="p-1.5 text-slate-500 hover:text-red-400 opacity-0 group-hover/item:opacity-100 transition-all">
            <Trash2 class="w-3.5 h-3.5" />
          </button>
        </div>
        <button @click.prevent="addSubstance('products')" class="w-full py-2 bg-green-500/5 hover:bg-green-500/10 border border-dashed border-green-500/30 rounded-xl text-[9px] font-black text-green-500 flex items-center justify-center gap-1 uppercase tracking-widest transition-all">
          <Plus class="w-3 h-3" /> Product
        </button>
      </div>
    </div>

    <!-- 预览 -->
    <div v-if="generatedDisplay" class="p-3 bg-blue-500/5 border border-blue-500/10 rounded-xl mt-2">
      <span class="block text-[8px] font-bold text-blue-400 uppercase tracking-widest mb-1 opacity-60">Formula Preview</span>
      <span class="text-xs font-mono font-bold text-white break-all leading-tight">{{ generatedDisplay }}</span>
    </div>
  </div>
</template>

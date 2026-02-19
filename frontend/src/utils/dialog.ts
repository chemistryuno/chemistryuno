import { reactive, type Ref } from 'vue'

interface DialogState {
  show: boolean
  type: 'alert' | 'confirm' | 'prompt'
  title: string
  message: string
  confirmText: string
  cancelText: string
  inputValue: string
  inputPlaceholder: string
  closeDelay: number
  resolve: ((value: any) => void) | null
}

const state = reactive<DialogState>({
  show: false,
  type: 'alert',
  title: '',
  message: '',
  confirmText: '确定',
  cancelText: '取消',
  inputValue: '',
  inputPlaceholder: '',
  closeDelay: 0,
  resolve: null
})

// Toast 组件引用（需要在组件中设置）
let toastRef: Ref<any> | null = null

export const setToastRef = (ref: Ref<any>) => {
  toastRef = ref
}

export const useDialog = () => {
  const showAlert = (message: string, title = '提示', confirmText = '确定', closeDelay = 0) => {
    return new Promise<void>((resolve) => {
      state.show = true
      state.type = 'alert'
      state.title = title
      state.message = message
      state.confirmText = confirmText
      state.closeDelay = closeDelay
      state.resolve = resolve
    })
  }

  const showConfirm = (message: string, title = '确认', confirmText = '确定', cancelText = '取消') => {
    return new Promise<boolean>((resolve) => {
      state.show = true
      state.type = 'confirm'
      state.title = title
      state.message = message
      state.confirmText = confirmText
      state.cancelText = cancelText
      state.resolve = resolve
    })
  }

  const showPrompt = (message: string, placeholder = '', title = '输入', confirmText = '确定', cancelText = '取消') => {
    return new Promise<string | null>((resolve) => {
      state.show = true
      state.type = 'prompt'
      state.title = title
      state.message = message
      state.inputPlaceholder = placeholder
      state.inputValue = ''
      state.confirmText = confirmText
      state.cancelText = cancelText
      state.resolve = resolve
    })
  }

  const handleConfirm = () => {
    if (state.type === 'prompt') {
      state.resolve?.(state.inputValue)
    } else if (state.type === 'confirm') {
      state.resolve?.(true)
    } else {
      state.resolve?.(undefined)
    }
    state.show = false
  }

  const handleCancel = () => {
    if (state.type === 'confirm' || state.type === 'prompt') {
      state.resolve?.(state.type === 'confirm' ? false : null)
    }
    state.show = false
  }

  const closeDialog = () => {
    if (state.show) {
      handleCancel()
    }
  }

  const showToast = (
    message: string,
    title?: string,
    type: 'info' | 'success' | 'warning' | 'error' = 'info',
    duration: number = 4000
  ) => {
    if (toastRef?.value) {
      toastRef.value.showToast(message, title, type, duration)
    } else {
      console.warn('Toast component not initialized. Falling back to alert.')
      showAlert(message, title || '提示')
    }
  }

  return {
    state,
    showAlert,
    showConfirm,
    showPrompt,
    showToast,
    handleConfirm,
    handleCancel,
    closeDialog
  }
}

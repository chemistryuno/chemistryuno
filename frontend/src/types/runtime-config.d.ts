export {}

declare global {
  interface Window {
    __CHEM_RUNTIME_CONFIG?: {
      apiOrigin?: string
    }
  }
}

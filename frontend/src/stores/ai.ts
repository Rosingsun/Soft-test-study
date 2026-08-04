import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AiApiConfig } from '@/types/ai'

const STORAGE_KEY = 'ai_api_config'

function loadFromStorage(): AiApiConfig {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch {
    // ignore
  }
  return defaultConfig()
}

function saveToStorage(config: AiApiConfig) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
}

function defaultConfig(): AiApiConfig {
  return {
    provider: 'deepseek',
    api_key: '',
    base_url: 'https://api.deepseek.com',
    model: 'deepseek-chat',
  }
}

export const useAiStore = defineStore('ai', () => {
  const config = ref<AiApiConfig>(loadFromStorage())

  const hasConfig = computed(() => !!config.value.api_key && !!config.value.base_url && !!config.value.model)

  function saveConfig(c: AiApiConfig) {
    config.value = { ...c }
    saveToStorage(config.value)
  }

  function updateProvider(provider: string, baseUrl: string, model: string) {
    config.value.provider = provider
    config.value.base_url = baseUrl
    config.value.model = model
    saveToStorage(config.value)
  }

  function clearConfig() {
    config.value = defaultConfig()
    localStorage.removeItem(STORAGE_KEY)
  }

  function getConfig(): AiApiConfig {
    return { ...config.value }
  }

  return { config, hasConfig, saveConfig, updateProvider, clearConfig, getConfig }
})

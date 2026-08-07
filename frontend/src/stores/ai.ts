import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AiApiConfig } from '@/types/ai'

const STORAGE_KEY = 'ai_api_config'

function emptyConfig(): AiApiConfig {
  return { provider: '', api_key: '', base_url: '', model: '' }
}

function loadFromStorage(): AiApiConfig {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      // 校验字段完整，避免旧版本/脏数据带出默认值
      if (parsed && typeof parsed === 'object' && parsed.api_key && parsed.base_url && parsed.model) {
        return parsed
      }
    }
  } catch {
    // ignore
  }
  return emptyConfig()
}

function saveToStorage(config: AiApiConfig) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(config))
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
    config.value = emptyConfig()
    localStorage.removeItem(STORAGE_KEY)
  }

  function getConfig(): AiApiConfig {
    return { ...config.value }
  }

  return { config, hasConfig, saveConfig, updateProvider, clearConfig, getConfig }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AiApiConfig } from '@/types/ai'
import {
  getSessionItem,
  setSessionItem,
  removeSessionItem,
  migrateAiKeyFromLocalStorage,
} from '@/utils/storage'

const LEGACY_STORAGE_KEY = 'ai_api_config'
const AI_KEY = 'ai_key'
// 其它非敏感配置（provider / base_url / model）继续存在 localStorage
const NON_SENSITIVE_KEY = 'ai_api_meta'

function loadMeta(): Pick<AiApiConfig, 'provider' | 'base_url' | 'model'> {
  try {
    const raw = localStorage.getItem(NON_SENSITIVE_KEY)
    if (raw) {
      const parsed = JSON.parse(raw)
      if (parsed && typeof parsed === 'object') {
        return {
          provider: typeof parsed.provider === 'string' ? parsed.provider : '',
          base_url: typeof parsed.base_url === 'string' ? parsed.base_url : '',
          model: typeof parsed.model === 'string' ? parsed.model : '',
        }
      }
    }
  } catch {
    // ignore
  }
  return { provider: '', base_url: '', model: '' }
}

function saveMeta(meta: Pick<AiApiConfig, 'provider' | 'base_url' | 'model'>) {
  localStorage.setItem(NON_SENSITIVE_KEY, JSON.stringify(meta))
}

export const useAiStore = defineStore('ai', () => {
  // 启动时做一次 localStorage.ai_key → sessionStorage.ai_key 的迁移
  const migrated = migrateAiKeyFromLocalStorage()
  if (migrated) {
    // 旧 localStorage.ai_api_config（如有）也清掉，避免残留 api_key
    try { localStorage.removeItem(LEGACY_STORAGE_KEY) } catch { /* ignore */ }
  }
  // 同步 meta（如果老版本存在 ai_api_config，尝试从中读出非敏感字段后转储）
  if (!localStorage.getItem(NON_SENSITIVE_KEY)) {
    try {
      const legacy = localStorage.getItem(LEGACY_STORAGE_KEY)
      if (legacy) {
        const parsed = JSON.parse(legacy)
        saveMeta({
          provider: parsed?.provider || '',
          base_url: parsed?.base_url || '',
          model: parsed?.model || '',
        })
        // api_key 已在 sessionStorage 中（迁移后），删除原复合键
        localStorage.removeItem(LEGACY_STORAGE_KEY)
      }
    } catch { /* ignore */ }
  }

  const meta = ref(loadMeta())
  // aiKey 仅在内存中占位（不持久），实际值每次从 sessionStorage 实时读
  const aiKey = ref<string>(getSessionItem(AI_KEY) || '')

  const config = computed<AiApiConfig>(() => ({
    provider: meta.value.provider,
    api_key: aiKey.value,
    base_url: meta.value.base_url,
    model: meta.value.model,
  }))

  const hasConfig = computed(
    () => !!getSessionItem(AI_KEY) && !!meta.value.base_url && !!meta.value.model,
  )

  function syncConfig() {
    aiKey.value = getSessionItem(AI_KEY) || ''
  }

  function saveConfig(c: AiApiConfig) {
    meta.value = { provider: c.provider, base_url: c.base_url, model: c.model }
    saveMeta(meta.value)
    setAiKey(c.api_key)
  }

  function setAiKey(key: string) {
    if (key && key.trim() !== '') {
      setSessionItem(AI_KEY, key)
    } else {
      removeSessionItem(AI_KEY)
    }
    aiKey.value = key
  }

  function getAiKey(): string {
    // 实时从 sessionStorage 取，确保多 tab / 跨调用实时一致
    const v = getSessionItem(AI_KEY) || ''
    aiKey.value = v
    return v
  }

  function clearAiKey() {
    removeSessionItem(AI_KEY)
    aiKey.value = ''
  }

  function updateProvider(provider: string, baseUrl: string, model: string) {
    meta.value = { provider, base_url: baseUrl, model }
    saveMeta(meta.value)
  }

  function clearConfig() {
    meta.value = { provider: '', base_url: '', model: '' }
    clearAiKey()
    try { localStorage.removeItem(NON_SENSITIVE_KEY) } catch { /* ignore */ }
  }

  function getConfig(): AiApiConfig {
    return {
      provider: meta.value.provider,
      api_key: getAiKey(),
      base_url: meta.value.base_url,
      model: meta.value.model,
    }
  }

  // OPT-05: 监听 auth store 的 logout action，自动清除 AI Key
  try {
    // 动态 import 避免循环依赖（stores/auth 也未引用 ai store）
    import('./auth').then(({ useAuthStore }) => {
      const auth = useAuthStore()
      auth.$onAction(({ name, after }) => {
        if (name === 'logout') {
          after(() => {
            clearAiKey()
          })
        }
      })
    }).catch(() => { /* ignore */ })
  } catch { /* ignore */ }

  return {
    meta,
    aiKey,
    config,
    hasConfig,
    syncConfig,
    saveConfig,
    setAiKey,
    getAiKey,
    clearAiKey,
    updateProvider,
    clearConfig,
    getConfig,
  }
})

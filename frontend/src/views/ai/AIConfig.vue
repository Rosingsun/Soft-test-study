<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAiStore } from '@/stores/ai'
import { getAiProviders } from '@/api/ai'
import type { AiProvider } from '@/types/ai'

const router = useRouter()
const auth = useAuthStore()
const aiStore = useAiStore()

const providers = ref<AiProvider[]>([])
const loading = ref(true)
const testing = ref(false)
const testResult = ref<{ ok: boolean; msg: string } | null>(null)

const form = ref({
  provider: aiStore.config.provider,
  api_key: aiStore.config.api_key,
  base_url: aiStore.config.base_url,
  model: aiStore.config.model,
})

// 「记住 API Key」开关：勾选后 Key 明文落盘到本机 localStorage
const remember = computed({
  get: () => aiStore.remember,
  set: (v: boolean) => aiStore.setRemember(v),
})

onMounted(async () => {
  // 先同步一次（例如勾选「记住」后重新打开页面，需要从落盘值恢复）
  aiStore.syncConfig()
  form.value = {
    provider: aiStore.config.provider,
    api_key: aiStore.config.api_key,
    base_url: aiStore.config.base_url,
    model: aiStore.config.model,
  }
  try {
    const res = await getAiProviders()
    providers.value = res.providers
  } catch {
    // 静默处理
  } finally {
    loading.value = false
  }
})

function onProviderChange() {
  if (!form.value.provider) {
    form.value.base_url = ''
    form.value.model = ''
    return
  }
  const p = providers.value.find(x => x.provider === form.value.provider)
  if (p) {
    form.value.base_url = p.base_url
    form.value.model = p.models[0] || ''
  }
}

function save() {
  if (!form.value.api_key) {
    testResult.value = { ok: false, msg: '请输入 API Key' }
    return
  }
  if (!form.value.base_url) {
    testResult.value = { ok: false, msg: '请输入 API 地址' }
    return
  }
  aiStore.saveConfig({
    provider: form.value.provider,
    api_key: form.value.api_key,
    base_url: form.value.base_url,
    model: form.value.model,
  })
  testResult.value = { ok: true, msg: '配置已保存' }
}

async function testConnection() {
  if (!form.value.api_key) {
    testResult.value = { ok: false, msg: '请先输入 API Key' }
    return
  }
  if (!form.value.base_url) {
    testResult.value = { ok: false, msg: '请先设置 API 地址' }
    return
  }
  save()
  testing.value = true
  testResult.value = null
  try {
    const { generateQuestions } = await import('@/api/ai')
    await generateQuestions({
      api_config: aiStore.getConfig(),
      subject_id: auth.selectedSubjectId,
      chapter_id: 0,
      types: ['single'],
      difficulty: 'easy',
      count: 1,
    })
    testResult.value = { ok: true, msg: '连接成功！AI 接口可用' }
  } catch (e) {
    testResult.value = { ok: false, msg: `连接失败: ${(e as Error).message}` }
  } finally {
    testing.value = false
  }
}

function clearConfig() {
  aiStore.clearConfig()
  form.value = {
    provider: '',
    api_key: '',
    base_url: '',
    model: '',
  }
  testResult.value = null
}

function goPractice() {
  router.push('/ai/practice')
}
</script>

<template>
  <div class="mx-auto max-w-2xl">
    <div class="mb-6">
      <h1 class="text-2xl font-bold text-gray-900">AI 学习配置</h1>
      <p class="mt-1 text-sm text-gray-500">配置你的 AI API 后即可使用 AI 智能出题功能</p>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <div class="h-8 w-8 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent" />
    </div>

    <div v-else class="rounded-lg bg-white p-6 shadow-sm">
      <div class="mb-5 flex items-start gap-3">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50">
          <svg class="h-5 w-5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9.75 3.104v5.714a2.25 2.25 0 01-.659 1.591L5 14.5M9.75 3.104c-.251.023-.501.05-.75.082m.75-.082a24.301 24.301 0 014.5 0m0 0v5.714c0 .597.237 1.17.659 1.591L19.8 15.3M14.25 3.104c.251.023.501.05.75.082M19.8 15.3l-1.57.393A9.065 9.065 0 0112 15a9.065 9.065 0 00-6.23.693L5 14.5m14.8.8l1.402 1.402c1.232 1.232.65 3.318-1.067 3.611A48.309 48.309 0 0112 21c-2.773 0-5.491-.235-8.135-.687-1.718-.293-2.3-2.379-1.067-3.61L5 14.5" />
          </svg>
        </div>
        <div>
          <h2 class="text-base font-semibold text-gray-900">API 配置</h2>
          <p class="mt-0.5 text-sm text-gray-500">
            API 地址与模型长期保存在本机；API Key 默认只在当前标签页有效，可按需勾选「记住」
          </p>
        </div>
      </div>

      <div class="space-y-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">AI 提供商</label>
          <select
            v-model="form.provider"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            @change="onProviderChange"
          >
            <option value="">请选择提供商（或选「自定义」自行填写）</option>
            <option v-for="p in providers" :key="p.provider" :value="p.provider">{{ p.name }}</option>
          </select>
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">API 地址</label>
          <input
            v-model="form.base_url"
            type="text"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            placeholder="https://api.example.com/v1"
          />
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">模型</label>
          <input
            v-model="form.model"
            type="text"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            placeholder="模型名称，如 my-model"
          />
        </div>

        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">API Key</label>
          <input
            v-model="form.api_key"
            type="password"
            class="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none focus:ring-1 focus:ring-indigo-500"
            placeholder="sk-xxxxxxxxxxxxxxxx"
          />
          <p class="mt-1 text-xs text-gray-400">Key 不会上传到服务器，仅保存在你的浏览器中</p>
          <label class="mt-3 flex cursor-pointer items-start gap-2">
            <input
              v-model="remember"
              type="checkbox"
              class="mt-0.5 h-4 w-4 cursor-pointer rounded border-gray-300 text-indigo-600 focus:ring-indigo-500"
            />
            <span class="text-sm text-gray-700">
              记住 API Key（下次打开自动带入）
              <span class="block text-xs text-gray-400">
                不勾选时 Key 只在当前标签页有效，关闭标签页后需要重新填写
              </span>
            </span>
          </label>
          <p v-if="remember" class="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-700">
            已开启记住：Key 会以明文保存在本机浏览器，公用/共享电脑请勿勾选。
          </p>
        </div>
      </div>

      <div v-if="testResult" class="mt-4 rounded-lg px-4 py-2 text-sm" :class="testResult.ok ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-600'">
        {{ testResult.msg }}
      </div>

      <div class="mt-6 flex flex-wrap justify-between gap-2">
        <div class="flex gap-2">
          <button
            class="cursor-pointer rounded-lg border border-gray-300 bg-white px-4 py-2 text-sm text-gray-600 hover:bg-gray-50"
            @click="clearConfig"
          >
            清除配置
          </button>
          <button
            class="cursor-pointer rounded-lg border border-indigo-300 bg-indigo-50 px-4 py-2 text-sm text-indigo-600 hover:bg-indigo-100 disabled:opacity-50"
            :disabled="testing"
            @click="testConnection"
          >
            {{ testing ? '测试中...' : '测试连接' }}
          </button>
        </div>
        <div class="flex gap-2">
          <button
            class="cursor-pointer rounded-lg bg-indigo-600 px-4 py-2 text-sm text-white hover:bg-indigo-700"
            @click="save"
          >
            保存配置
          </button>
          <button
            v-if="aiStore.hasConfig"
            class="cursor-pointer rounded-lg bg-emerald-600 px-4 py-2 text-sm text-white hover:bg-emerald-700"
            @click="goPractice"
          >
            开始 AI 练习 →
          </button>
        </div>
      </div>
    </div>

    <div class="mt-6 rounded-lg bg-white p-6 shadow-sm">
      <h3 class="mb-3 text-sm font-semibold text-gray-700">使用说明</h3>
      <ol class="list-decimal space-y-1.5 pl-5 text-sm text-gray-500">
        <li>选择提供商，或选「自定义」自行填写任意 OpenAI 兼容接口</li>
        <li>前往对应平台获取 API Key（如 DeepSeek: platform.deepseek.com，MiniMax: platform.minimaxi.com）</li>
        <li>确认 API 地址与模型无误（可手动修改），填入 API Key</li>
        <li>点击"保存配置"；API 地址与模型会长期保存，API Key 默认仅当前标签页有效</li>
        <li>若希望关掉浏览器后仍保留 Key，勾选「记住 API Key」（会明文存在本机）</li>
        <li>点击"测试连接"确认可用，再前往 AI 练习页面让 AI 出题</li>
      </ol>
    </div>
  </div>
</template>

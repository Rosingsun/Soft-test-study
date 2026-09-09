<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { createAdminInvitationCode, getAdminUserDetail, listAdminInvitationCodes } from '@/api/admin'
import type { AdminUserDetailResp, InvitationCodeResp, InvitationCodeState } from '@/types/admin'
import { buildInviteUrl } from '@/utils/site'
import { copyText } from '@/utils/clipboard'
import { formatDuration, formatPercent } from '@/utils/format'
import { showToast } from '@/utils/toast'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseProgressBar from '@/components/common/BaseProgressBar.vue'

type StateFilter = 'all' | InvitationCodeState
type DetailTab = 'user' | 'study' | 'answer'

const list = ref<InvitationCodeResp[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const error = ref('')

const stateFilter = ref<StateFilter>('all')
const keyword = ref('')
const keywordApplied = ref('')

const generateOpen = ref(false)
const generateNote = ref('')
const generating = ref(false)
// 生成结果：展示新码与邀请链接，便于立即分发
const generatedOpen = ref(false)
const generated = ref<InvitationCodeResp | null>(null)

const detailOpen = ref(false)
const detailLoading = ref(false)
const detailError = ref('')
const detail = ref<AdminUserDetailResp | null>(null)
const detailTab = ref<DetailTab>('user')

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const generatedLink = computed(() => (generated.value ? buildInviteUrl(generated.value.code) : ''))

const stateMeta: Record<InvitationCodeState, { label: string; type: 'success' | 'warning' | 'danger' | 'info' | 'default' }> = {
  unused: { label: '未使用', type: 'success' },
  used: { label: '已使用', type: 'info' },
  disabled: { label: '已停用', type: 'danger' },
}

const stateTabs = [
  { label: '全部', value: 'all' },
  { label: '未使用', value: 'unused' },
  { label: '已使用', value: 'used' },
  { label: '已停用', value: 'disabled' },
]

const detailTabs = [
  { label: '用户信息', value: 'user' },
  { label: '学习信息', value: 'study' },
  { label: '答题信息', value: 'answer' },
]

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await listAdminInvitationCodes({
      state: stateFilter.value === 'all' ? undefined : stateFilter.value,
      keyword: keywordApplied.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    list.value = data.list
    total.value = data.total
    page.value = data.page
    pageSize.value = data.page_size
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

function changeFilter(value: string | number) {
  stateFilter.value = value as StateFilter
  page.value = 1
  load()
}

function applyKeyword() {
  keywordApplied.value = keyword.value.trim()
  page.value = 1
  load()
}

function resetFilters() {
  stateFilter.value = 'all'
  keyword.value = ''
  keywordApplied.value = ''
  page.value = 1
  load()
}

function goPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  load()
}

async function copyCode(row: InvitationCodeResp) {
  const ok = await copyText(row.code)
  showToast(ok ? '邀请码已复制' : '复制失败，请手动复制', ok ? 'success' : 'error')
}

async function copyLink(row: InvitationCodeResp) {
  const ok = await copyText(buildInviteUrl(row.code))
  showToast(ok ? '邀请链接已复制' : '复制失败，请手动复制', ok ? 'success' : 'error')
}

async function submitGenerate() {
  generating.value = true
  try {
    const created = await createAdminInvitationCode({ count: 1, note: generateNote.value.trim() })
    showToast('邀请码生成成功', 'success')
    generateOpen.value = false
    generateNote.value = ''
    // 回到第一页并清空筛选，确保新码立即可见
    stateFilter.value = 'all'
    keyword.value = ''
    keywordApplied.value = ''
    page.value = 1
    await load()
    // 弹出生成结果，可直接复制邀请码 / 邀请链接
    if (created.length > 0) {
      generated.value = created[0]
      generatedOpen.value = true
    }
  } catch {
    // 请求层已统一 toast
  } finally {
    generating.value = false
  }
}

async function openDetail(row: InvitationCodeResp) {
  if (!row.used_by) return
  detailOpen.value = true
  detailTab.value = 'user'
  detail.value = null
  detailError.value = ''
  detailLoading.value = true
  try {
    detail.value = await getAdminUserDetail(row.used_by)
  } catch (e) {
    detailError.value = (e as Error).message
  } finally {
    detailLoading.value = false
  }
}

function retryDetail() {
  const userId = detail.value?.user.id
  if (!userId) return
  detailError.value = ''
  detailLoading.value = true
  getAdminUserDetail(userId)
    .then(data => {
      detail.value = data
    })
    .catch(e => {
      detailError.value = (e as Error).message
    })
    .finally(() => {
      detailLoading.value = false
    })
}

// ============ 详情展示辅助 ============

function dash(value: string | number | null | undefined) {
  if (value === null || value === undefined || value === '') return '-'
  return String(value)
}

function accuracyClass(a: number) {
  if (a >= 80) return 'text-emerald-600'
  if (a >= 60) return 'text-indigo-600'
  return 'text-red-500'
}

function accuracyBarColor(a: number): 'indigo' | 'emerald' | 'yellow' {
  if (a >= 80) return 'emerald'
  if (a >= 60) return 'indigo'
  return 'yellow'
}

function shortDate(s: string) {
  return s ? s.slice(5) : '-'
}

function isUnused(row: InvitationCodeResp) {
  return row.state === 'unused'
}

/** 近 14 日练习量柱状图数据（按比例自适应高度） */
const recentBars = computed(() => {
  const recent = detail.value?.study.recent ?? []
  const items = recent.slice(-14)
  const max = Math.max(1, ...items.map(d => d.total_count))
  return items.map(d => ({
    date: shortDate(d.date),
    count: d.total_count,
    height: Math.round((d.total_count / max) * 100),
  }))
})

const roleLabel = computed(() => (detail.value?.user.role === 'admin' ? '管理员' : '学员'))

const statusLabel = computed(() => {
  const s = detail.value?.user.status
  if (s === 1) return { label: '正常', type: 'success' as const }
  return { label: '已禁用', type: 'danger' as const }
})
</script>

<template>
  <div>
    <BasePageHeader title="邀请码管理" subtitle="生成与分发注册邀请码，追踪使用情况并查看用户档案">
      <template #actions>
        <BaseButton type="secondary" @click="load">刷新</BaseButton>
        <BaseButton @click="generateOpen = true">生成邀请码</BaseButton>
      </template>
    </BasePageHeader>

    <!-- 筛选区 -->
    <div class="mb-5 flex flex-col gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
      <BaseTabs v-model="stateFilter" variant="pill" size="sm" :tabs="stateTabs" @update:model-value="changeFilter" />
      <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
        <div class="w-full sm:w-64">
          <BaseInput v-model="keyword" placeholder="邀请码 / 备注 / 使用人" @keyup.enter="applyKeyword" />
        </div>
        <div class="flex items-center gap-2">
          <BaseButton size="sm" @click="applyKeyword">搜索</BaseButton>
          <BaseButton type="ghost" size="sm" @click="resetFilters">重置</BaseButton>
        </div>
      </div>
    </div>

    <BaseCard v-if="error">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </BaseCard>

    <BaseCard v-else-if="!loading && list.length === 0">
      <BaseEmpty
        title="暂无符合条件的邀请码"
        description="调整筛选条件，或生成一个新的邀请码"
      >
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="generateOpen = true">生成邀请码</BaseButton>
        </template>
      </BaseEmpty>
    </BaseCard>

    <template v-else>
      <div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
        <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead>
              <tr class="border-b border-gray-100 bg-gray-50/70 text-xs font-semibold uppercase tracking-wider text-gray-500">
                <th scope="col" class="whitespace-nowrap px-5 py-3">邀请码</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 lg:table-cell">备注</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3">状态</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 md:table-cell">使用人</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 xl:table-cell">使用时间</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 lg:table-cell">创建人</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 md:table-cell">创建时间</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3 text-right">操作</th>
              </tr>
            </thead>

            <tbody class="divide-y divide-gray-50">
              <!-- 骨架屏：列宽与真实行一致，避免数据到达前后跳动 -->
              <template v-if="loading">
                <tr v-for="i in 5" :key="i">
                  <td class="px-5 py-4"><div class="skeleton h-4 w-36 rounded" /></td>
                  <td class="hidden px-5 py-4 lg:table-cell"><div class="skeleton h-4 w-28 rounded" /></td>
                  <td class="px-5 py-4"><div class="skeleton h-5 w-16 rounded-full" /></td>
                  <td class="hidden px-5 py-4 md:table-cell"><div class="skeleton h-4 w-20 rounded" /></td>
                  <td class="hidden px-5 py-4 xl:table-cell"><div class="skeleton h-4 w-28 rounded" /></td>
                  <td class="hidden px-5 py-4 lg:table-cell"><div class="skeleton h-4 w-20 rounded" /></td>
                  <td class="hidden px-5 py-4 md:table-cell"><div class="skeleton h-4 w-28 rounded" /></td>
                  <td class="px-5 py-4">
                    <div class="ml-auto flex items-center justify-end gap-2">
                      <div class="skeleton h-8 w-24 rounded-lg" />
                      <div class="skeleton h-8 w-24 rounded-lg" />
                    </div>
                  </td>
                </tr>
              </template>

              <template v-else>
                <tr
                  v-for="row in list"
                  :key="row.id"
                  class="group transition-colors duration-200 hover:bg-indigo-50/30"
                >
                  <!-- 邀请码 -->
                  <td class="px-5 py-3.5">
                    <div class="flex items-center gap-1.5">
                      <span class="rounded-lg bg-gray-50 px-2.5 py-1 font-mono text-xs font-semibold tracking-wide text-gray-800 ring-1 ring-inset ring-gray-200">
                        {{ row.code }}
                      </span>
                      <button
                        class="cursor-pointer rounded-lg p-1.5 text-gray-400 transition-colors hover:bg-indigo-50 hover:text-indigo-600"
                        title="复制邀请码"
                        @click="copyCode(row)"
                      >
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                          <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v10.5A2.25 2.25 0 007.5 18h6a2.25 2.25 0 002.25-2.25V6.75" />
                          <path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6.75H18A2.25 2.25 0 0120.25 9v10.5A2.25 2.25 0 0118 21.75H9A2.25 2.25 0 016.75 19.5V9A2.25 2.25 0 019 6.75" />
                        </svg>
                      </button>
                    </div>
                  </td>

                  <!-- 备注 -->
                  <td class="hidden max-w-[14rem] px-5 py-3.5 lg:table-cell">
                    <span class="block truncate text-gray-600" :title="row.note">{{ dash(row.note) }}</span>
                  </td>

                  <!-- 状态 -->
                  <td class="whitespace-nowrap px-5 py-3.5">
                    <BaseBadge :type="stateMeta[row.state].type" dot>{{ stateMeta[row.state].label }}</BaseBadge>
                  </td>

                  <!-- 使用人 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 md:table-cell">
                    <button
                      v-if="row.used_by"
                      class="cursor-pointer font-medium text-indigo-600 transition-colors hover:text-indigo-700"
                      @click="openDetail(row)"
                    >{{ row.used_by_name }}</button>
                    <span v-else class="text-gray-400">
                      <template v-if="isUnused(row)">尚未使用</template>
                      <template v-else>—</template>
                    </span>
                  </td>

                  <!-- 使用时间 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 xl:table-cell">{{ dash(row.used_at) }}</td>

                  <!-- 创建人 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 lg:table-cell">{{ dash(row.created_by_name) }}</td>

                  <!-- 创建时间 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 md:table-cell">{{ dash(row.created_at) }}</td>

                  <!-- 操作 -->
                  <td class="px-5 py-3.5">
                    <div class="flex items-center justify-end gap-2 whitespace-nowrap">
                      <BaseButton size="sm" type="secondary" @click="copyLink(row)">复制链接</BaseButton>
                      <BaseButton
                        v-if="row.used_by"
                        size="sm"
                        type="ghost"
                        @click="openDetail(row)"
                      >详情</BaseButton>
                      <span v-else class="text-xs text-gray-400">—</span>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 分页 -->
      <div class="mt-5 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-gray-100 bg-white px-4 py-3 shadow-sm">
        <p class="text-xs text-gray-500">共 {{ total }} 条邀请码</p>
        <div class="flex items-center gap-2">
          <BaseButton size="sm" type="secondary" :disabled="page <= 1" @click="goPage(page - 1)">上一页</BaseButton>
          <span class="text-xs font-medium text-gray-600">第 {{ page }} / {{ totalPages }} 页</span>
          <BaseButton size="sm" type="secondary" :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</BaseButton>
        </div>
      </div>
    </template>

    <!-- 生成邀请码 -->
    <BaseModal v-model="generateOpen" title="生成邀请码" width="max-w-md">
      <div class="space-y-4">
        <p class="text-sm text-gray-500">生成后将立即出现在列表顶部，可复制链接分发给新用户。</p>
        <BaseInput v-model="generateNote" label="备注（可选）" placeholder="例如：2026 春季批次" />
      </div>
      <div class="mt-6 flex justify-end gap-2 border-t border-gray-50 pt-4">
        <BaseButton type="ghost" size="sm" @click="generateOpen = false">取消</BaseButton>
        <BaseButton size="sm" :loading="generating" @click="submitGenerate">确认生成</BaseButton>
      </div>
    </BaseModal>

    <!-- 生成结果：可直接复制分发 -->
    <BaseModal v-model="generatedOpen" title="邀请码已生成" width="max-w-lg">
      <div v-if="generated" class="space-y-4">
        <div class="rounded-xl bg-gradient-to-r from-indigo-50 via-white to-violet-50 p-4 ring-1 ring-inset ring-indigo-100/60">
          <p class="text-xs text-gray-500">邀请码</p>
          <div class="mt-1.5 flex flex-wrap items-center gap-2">
            <span class="rounded-lg bg-white px-2.5 py-1 font-mono text-base font-semibold tracking-wide text-gray-900 ring-1 ring-inset ring-gray-200">
              {{ generated.code }}
            </span>
            <BaseButton size="sm" type="secondary" @click="copyCode(generated)">复制邀请码</BaseButton>
          </div>
          <p v-if="generated.note" class="mt-2 text-xs text-gray-500">备注：{{ generated.note }}</p>
        </div>

        <div>
          <p class="mb-1.5 text-xs text-gray-500">邀请链接（注册时自动带入邀请码）</p>
          <div class="rounded-lg border border-gray-200 bg-gray-50 px-3 py-2.5">
            <p class="break-all font-mono text-xs leading-relaxed text-gray-700">{{ generatedLink }}</p>
          </div>
          <BaseButton size="sm" class="mt-2" @click="copyLink(generated)">复制邀请链接</BaseButton>
        </div>
      </div>
      <div class="mt-6 flex justify-end border-t border-gray-50 pt-4">
        <BaseButton type="ghost" size="sm" @click="generatedOpen = false">关闭</BaseButton>
      </div>
    </BaseModal>

    <!-- 用户详情 -->
    <BaseModal v-model="detailOpen" title="用户详情" width="max-w-3xl">
      <BaseSkeleton v-if="detailLoading" variant="detail" />

      <div v-else-if="detailError" class="py-6">
        <BaseEmpty title="加载失败" :description="detailError">
          <template #action>
            <BaseButton type="secondary" size="sm" class="mt-4" @click="retryDetail">重新加载</BaseButton>
          </template>
        </BaseEmpty>
      </div>

      <div v-else-if="detail" class="max-h-[68vh] space-y-4 overflow-y-auto pr-1">
        <!-- 用户概览 -->
        <div class="flex flex-wrap items-center gap-3 rounded-xl bg-gradient-to-r from-indigo-50 via-white to-violet-50 p-4 ring-1 ring-inset ring-indigo-100/60">
          <div class="bg-brand-gradient flex h-11 w-11 shrink-0 items-center justify-center rounded-xl text-base font-semibold text-white shadow-sm">
            {{ (detail.user.nickname || detail.user.username || '?').slice(0, 1).toUpperCase() }}
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-base font-semibold tracking-tight text-gray-900">
              {{ detail.user.nickname || detail.user.username }}
            </p>
            <p class="truncate text-xs text-gray-500">@{{ detail.user.username }} · {{ detail.user.email }}</p>
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <BaseBadge :type="detail.user.role === 'admin' ? 'info' : 'default'">{{ roleLabel }}</BaseBadge>
            <BaseBadge :type="statusLabel.type" dot>{{ statusLabel.label }}</BaseBadge>
          </div>
        </div>

        <BaseTabs v-model="detailTab" variant="pill" size="sm" :tabs="detailTabs" />

        <!-- 用户信息 -->
        <div v-if="detailTab === 'user'" class="grid gap-3 sm:grid-cols-2">
          <div class="rounded-xl border border-gray-100 bg-gray-50/60 p-4">
            <p class="text-xs text-gray-400">账号信息</p>
            <dl class="mt-2 space-y-2 text-sm">
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">用户 ID</dt>
                <dd class="font-medium text-gray-800">{{ detail.user.id }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">用户名</dt>
                <dd class="truncate font-medium text-gray-800">{{ detail.user.username }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">昵称</dt>
                <dd class="truncate font-medium text-gray-800">{{ dash(detail.user.nickname) }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">注册时间</dt>
                <dd class="font-medium text-gray-800">{{ dash(detail.user.created_at) }}</dd>
              </div>
            </dl>
          </div>

          <div class="rounded-xl border border-gray-100 bg-gray-50/60 p-4">
            <p class="text-xs text-gray-400">报考与验证</p>
            <dl class="mt-2 space-y-2 text-sm">
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">邮箱</dt>
                <dd class="truncate font-medium text-gray-800">{{ detail.user.email }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">邮箱验证</dt>
                <dd>
                  <BaseBadge :type="detail.user.email_verified ? 'success' : 'warning'" dot>
                    {{ detail.user.email_verified ? '已验证' : '未验证' }}
                  </BaseBadge>
                </dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">报考等级</dt>
                <dd class="truncate font-medium text-gray-800">{{ dash(detail.user.level_name) }}</dd>
              </div>
              <div class="flex justify-between gap-3">
                <dt class="text-gray-500">报考科目</dt>
                <dd class="truncate font-medium text-gray-800">{{ dash(detail.user.subject_name) }}</dd>
              </div>
            </dl>
          </div>

          <div class="rounded-xl border border-gray-100 bg-gray-50/60 p-4 sm:col-span-2">
            <p class="text-xs text-gray-400">注册来源</p>
            <template v-if="detail.invite_code">
              <div class="mt-2 flex flex-wrap items-center gap-2">
                <span class="rounded-lg bg-white px-2.5 py-1 font-mono text-sm font-semibold text-gray-800 ring-1 ring-inset ring-gray-200">
                  {{ detail.invite_code.code }}
                </span>
                <span class="text-xs text-gray-500">
                  备注：{{ dash(detail.invite_code.note) }} · 使用于 {{ dash(detail.invite_code.used_at) }}
                </span>
              </div>
            </template>
            <p v-else class="mt-2 text-sm text-gray-600">该用户通过管理员白名单注册，未消耗邀请码。</p>
          </div>
        </div>

        <!-- 学习信息 -->
        <div v-else-if="detailTab === 'study'" class="space-y-4">
          <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">练习总量</p>
              <p class="mt-1 text-2xl font-bold tracking-tight text-gray-900">{{ detail.study.overview.total_practiced }}</p>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">练习正确率</p>
              <p class="mt-1 text-2xl font-bold tracking-tight" :class="accuracyClass(detail.study.overview.accuracy)">
                {{ formatPercent(detail.study.overview.accuracy) }}
              </p>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">学习天数</p>
              <p class="mt-1 text-2xl font-bold tracking-tight text-gray-900">{{ detail.study.overview.study_days }}</p>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">错题数</p>
              <p class="mt-1 text-2xl font-bold tracking-tight text-red-500">{{ detail.study.overview.wrong_count }}</p>
            </div>
          </div>

          <BaseCard>
            <h3 class="mb-3 text-sm font-semibold tracking-tight text-gray-900">近 14 日活跃</h3>
            <div v-if="recentBars.length === 0" class="py-6 text-center text-xs text-gray-400">近期暂无练习记录</div>
            <div v-else class="flex h-24 items-end gap-1.5">
              <div v-for="bar in recentBars" :key="bar.date" class="group flex flex-1 flex-col items-center gap-1">
                <span class="text-[10px] font-medium text-gray-400 opacity-0 transition-opacity group-hover:opacity-100">{{ bar.count }}</span>
                <div class="flex w-full flex-1 items-end">
                  <div
                    class="w-full rounded-t bg-gradient-to-t from-indigo-500 to-violet-500 transition-all duration-300"
                    :style="{ height: Math.max(bar.height, 4) + '%' }"
                  />
                </div>
                <span class="text-[10px] text-gray-400">{{ bar.date }}</span>
              </div>
            </div>
          </BaseCard>

          <BaseCard>
            <h3 class="mb-3 text-sm font-semibold tracking-tight text-gray-900">科目掌握度</h3>
            <div v-if="detail.study.subjects.length === 0" class="py-6 text-center text-xs text-gray-400">暂无练习数据</div>
            <div v-else class="space-y-3">
              <div v-for="s in detail.study.subjects" :key="s.subject_id">
                <div class="mb-1 flex items-center justify-between gap-3 text-xs">
                  <span class="truncate text-gray-700">{{ s.subject_name || `科目 ${s.subject_id}` }}</span>
                  <span class="shrink-0 font-medium" :class="accuracyClass(s.accuracy)">
                    {{ formatPercent(s.accuracy) }} · {{ s.correct_count }}/{{ s.total_count }}
                  </span>
                </div>
                <BaseProgressBar :value="s.accuracy" :color="accuracyBarColor(s.accuracy)" size="sm" />
              </div>
            </div>
          </BaseCard>

          <BaseCard>
            <h3 class="mb-3 text-sm font-semibold tracking-tight text-gray-900">章节练习 TOP 5</h3>
            <div v-if="detail.study.chapters.length === 0" class="py-6 text-center text-xs text-gray-400">暂无章节练习数据</div>
            <div v-else class="space-y-2">
              <div
                v-for="(c, idx) in detail.study.chapters.slice(0, 5)"
                :key="c.chapter_id"
                class="flex items-center justify-between gap-3 rounded-lg border border-gray-100 px-3 py-2"
              >
                <div class="flex min-w-0 items-center gap-2">
                  <span class="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-indigo-50 text-[10px] font-bold text-indigo-600">
                    {{ idx + 1 }}
                  </span>
                  <span class="truncate text-sm text-gray-700">{{ c.chapter_name }}</span>
                </div>
                <span class="shrink-0 text-xs font-medium" :class="accuracyClass(c.accuracy)">
                  {{ formatPercent(c.accuracy) }} · {{ c.total_count }} 题
                </span>
              </div>
            </div>
          </BaseCard>
        </div>

        <!-- 答题信息 -->
        <div v-else class="space-y-4">
          <div class="grid grid-cols-2 gap-3">
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">考试次数</p>
              <p class="mt-1 text-2xl font-bold tracking-tight text-gray-900">{{ detail.answer.total_exams }}</p>
            </div>
            <div class="rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
              <p class="text-[10px] font-semibold uppercase tracking-wider text-gray-400">平均得分</p>
              <p class="mt-1 text-2xl font-bold tracking-tight text-indigo-600">{{ detail.answer.avg_exam_score.toFixed(1) }}</p>
            </div>
          </div>

          <BaseCard>
            <h3 class="mb-3 text-sm font-semibold tracking-tight text-gray-900">最近考试记录</h3>
            <div v-if="detail.answer.records.length === 0" class="py-6 text-center text-xs text-gray-400">暂无考试记录</div>
            <div v-else class="space-y-2">
              <div
                v-for="r in detail.answer.records"
                :key="r.id"
                class="rounded-lg border border-gray-100 px-3 py-2.5"
              >
                <div class="flex flex-wrap items-center justify-between gap-2">
                  <p class="min-w-0 flex-1 truncate text-sm font-medium text-gray-800">{{ r.template_name || 'AI 智能组卷' }}</p>
                  <BaseBadge :type="r.status === 'finished' ? 'success' : 'warning'">
                    {{ r.status === 'finished' ? '已完成' : '进行中' }}
                  </BaseBadge>
                </div>
                <div class="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500">
                  <span>得分 <b class="text-gray-800">{{ r.score }}</b> / {{ r.total_score }}</span>
                  <span>正确率 <b :class="accuracyClass(r.accuracy)">{{ formatPercent(r.accuracy) }}</b></span>
                  <span>用时 {{ formatDuration(r.duration) }}</span>
                  <span>{{ dash(r.created_at) }}</span>
                </div>
              </div>
            </div>
          </BaseCard>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

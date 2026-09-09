<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  createAdminSubject,
  deleteAdminSubject,
  listAdminSubjects,
  setAdminSubjectStatus,
  updateAdminSubject,
} from '@/api/admin'
import { getExamLevels } from '@/api/subject'
import type { ExamLevel } from '@/types/subject'
import type { AdminSubjectResp, AdminSubjectUpsertReq } from '@/types/admin'
import { showToast } from '@/utils/toast'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextarea from '@/components/common/BaseTextarea.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

interface SubjectForm {
  id: number | null
  level_id: number
  name: string
  short_name: string
  description: string
  icon: string
  sort_order: number
}

const list = ref<AdminSubjectResp[]>([])
const levels = ref<ExamLevel[]>([])
const loading = ref(true)
const error = ref('')

const levelFilter = ref<number>(0)
const keyword = ref('')
const keywordApplied = ref('')

const modalOpen = ref(false)
const saving = ref(false)
const form = ref<SubjectForm>(emptyForm())
const formError = ref('')

const dialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)
const pendingDelete = ref<AdminSubjectResp | null>(null)
const pendingStatus = ref<{ subject: AdminSubjectResp; status: 0 | 1 } | null>(null)

function emptyForm(): SubjectForm {
  return {
    id: null,
    level_id: levels.value[0]?.id || 0,
    name: '',
    short_name: '',
    description: '',
    icon: '',
    sort_order: 0,
  }
}

const isEdit = computed(() => form.value.id !== null)
const modalTitle = computed(() => (isEdit.value ? '编辑科目' : '新增科目'))

async function load() {
  loading.value = true
  error.value = ''
  try {
    list.value = await listAdminSubjects({
      level_id: levelFilter.value || undefined,
      keyword: keywordApplied.value || undefined,
    })
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function init() {
  if (levels.value.length === 0) {
    try {
      levels.value = await getExamLevels()
      if (levels.value.length > 0) {
        form.value.level_id = levels.value[0].id
      }
    } catch {
      /* 等级加载失败时表单等级下拉为空，由用户手动选择 */
    }
  }
  await load()
}

onMounted(init)

async function applyKeyword() {
  keywordApplied.value = keyword.value.trim()
  await load()
}

async function resetFilters() {
  levelFilter.value = 0
  keyword.value = ''
  keywordApplied.value = ''
  await load()
}

async function changeLevelFilter(value: string | number) {
  levelFilter.value = Number(value)
  await load()
}

function openCreate() {
  form.value = emptyForm()
  formError.value = ''
  modalOpen.value = true
}

function openEdit(row: AdminSubjectResp) {
  form.value = {
    id: row.id,
    level_id: row.level_id,
    name: row.name,
    short_name: row.short_name,
    description: row.description,
    icon: row.icon,
    sort_order: row.sort_order,
  }
  formError.value = ''
  modalOpen.value = true
}

function validateForm(): string {
  if (!form.value.level_id) return '请选择所属等级'
  if (!form.value.name.trim()) return '科目名称不能为空'
  if (form.value.short_name.trim().length > 20) return '简称不能超过 20 个字符'
  return ''
}

async function submit() {
  const err = validateForm()
  if (err) {
    formError.value = err
    return
  }
  saving.value = true
  try {
    const body: AdminSubjectUpsertReq = {
      level_id: form.value.level_id,
      name: form.value.name.trim(),
      short_name: form.value.short_name.trim(),
      description: form.value.description,
      icon: form.value.icon,
      sort_order: form.value.sort_order,
    }
    if (isEdit.value && form.value.id !== null) {
      await updateAdminSubject(form.value.id, body)
      showToast('科目已更新', 'success')
    } else {
      await createAdminSubject(body)
      showToast('科目已创建', 'success')
    }
    modalOpen.value = false
    await load()
  } catch {
    // 请求层已统一 toast
  } finally {
    saving.value = false
  }
}

function confirmToggleStatus(subject: AdminSubjectResp) {
  pendingStatus.value = { subject, status: subject.status === 1 ? 0 : 1 }
  dialog.value?.open()
}

function onConfirmAction() {
  if (pendingDelete.value) {
    void doDelete(pendingDelete.value)
  } else if (pendingStatus.value) {
    void doToggleStatus(pendingStatus.value)
  }
}

async function doToggleStatus(p: { subject: AdminSubjectResp; status: 0 | 1 }) {
  try {
    await setAdminSubjectStatus(p.subject.id, p.status)
    showToast(p.status === 1 ? '科目已启用' : '科目已停用', 'success')
    await load()
  } catch {
    // 请求层已统一 toast
  }
}

function confirmDelete(row: AdminSubjectResp) {
  pendingStatus.value = null
  pendingDelete.value = row
  dialog.value?.open()
}

async function doDelete(row: AdminSubjectResp) {
  try {
    await deleteAdminSubject(row.id)
    showToast('科目已删除', 'success')
    await load()
  } catch {
    // 请求层已统一 toast
  } finally {
    pendingDelete.value = null
  }
}

const dialogTitle = computed(() => {
  if (pendingDelete.value) return '删除科目'
  if (pendingStatus.value) {
    return pendingStatus.value.status === 1 ? '启用科目' : '停用科目'
  }
  return ''
})

const dialogDescription = computed(() => {
  const subject = pendingDelete.value || pendingStatus.value?.subject
  if (!subject) return ''
  const display = subject.name + (subject.short_name ? `（${subject.short_name}）` : '')
  if (pendingDelete.value) {
    return `确定删除科目「${display}」吗？该操作不可恢复。若其下仍有子科目或题目，将无法删除。`
  }
  if (pendingStatus.value) {
    const verb = pendingStatus.value.status === 1 ? '启用' : '停用'
    return pendingStatus.value.status === 1
      ? `确定${verb}科目「${display}」吗？启用后用户可在科目导航中看到它。`
      : `确定${verb}科目「${display}」吗？停用后用户将无法在科目导航中看到它，历史数据保留。`
  }
  return ''
})
</script>

<template>
  <div>
    <BasePageHeader title="科目管理" subtitle="维护考试等级下的科目，支持新增、编辑、停用与删除">
      <template #actions>
        <BaseButton type="secondary" @click="load">刷新</BaseButton>
        <BaseButton @click="openCreate">新增科目</BaseButton>
      </template>
    </BasePageHeader>

    <!-- 筛选区 -->
    <div class="mb-5 flex flex-col gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
      <BaseSelect v-model="levelFilter" class="sm:w-56" @change="changeLevelFilter(levelFilter)">
        <option :value="0">全部等级</option>
        <option v-for="level in levels" :key="level.id" :value="level.id">
          {{ level.name }}
        </option>
      </BaseSelect>
      <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
        <div class="w-full sm:w-64">
          <BaseInput v-model="keyword" placeholder="搜索科目名称 / 简称" @keyup.enter="applyKeyword" />
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
        title="暂无科目"
        description="点击右上角「新增科目」创建一个考试科目"
      >
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="openCreate">新增科目</BaseButton>
        </template>
      </BaseEmpty>
    </BaseCard>

    <template v-else>
      <div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
        <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead>
              <tr class="border-b border-gray-100 bg-gray-50/70 text-xs font-semibold uppercase tracking-wider text-gray-500">
                <th scope="col" class="whitespace-nowrap px-5 py-3">科目</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3">所属等级</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 lg:table-cell">子科目 / 题量</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3">状态</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 md:table-cell">排序</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <!-- 骨架屏 -->
              <template v-if="loading">
                <tr v-for="i in 5" :key="i">
                  <td class="px-5 py-4">
                    <div class="flex items-center gap-3">
                      <div class="skeleton h-10 w-10 rounded-lg" />
                      <div class="space-y-1.5">
                        <div class="skeleton h-4 w-28 rounded" />
                        <div class="skeleton h-3 w-20 rounded" />
                      </div>
                    </div>
                  </td>
                  <td class="px-5 py-4"><div class="skeleton h-4 w-16 rounded" /></td>
                  <td class="hidden px-5 py-4 lg:table-cell"><div class="skeleton h-4 w-20 rounded" /></td>
                  <td class="px-5 py-4"><div class="skeleton h-5 w-16 rounded-full" /></td>
                  <td class="hidden px-5 py-4 md:table-cell"><div class="skeleton h-4 w-8 rounded" /></td>
                  <td class="px-5 py-4">
                    <div class="flex items-center justify-end gap-2">
                      <div class="skeleton h-8 w-16 rounded-lg" />
                      <div class="skeleton h-8 w-16 rounded-lg" />
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
                  <!-- 科目 -->
                  <td class="px-5 py-3.5">
                    <div class="flex items-center gap-3">
                      <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-base text-indigo-600">
                        {{ row.icon || row.name.charAt(0) }}
                      </div>
                      <div class="min-w-0">
                        <div class="flex flex-wrap items-center gap-1.5">
                          <span class="font-medium text-gray-900">{{ row.name }}</span>
                          <BaseBadge v-if="row.short_name" type="info">{{ row.short_name }}</BaseBadge>
                        </div>
                        <p v-if="row.description" class="max-w-xs truncate text-xs text-gray-400" :title="row.description">
                          {{ row.description }}
                        </p>
                      </div>
                    </div>
                  </td>

                  <!-- 所属等级 -->
                  <td class="whitespace-nowrap px-5 py-3.5 text-gray-600">{{ row.level_name || `等级 ${row.level_id}` }}</td>

                  <!-- 子科目 / 题量 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 lg:table-cell">
                    {{ row.sub_subject_count }} 个子科目 · {{ row.question_count }} 道题
                  </td>

                  <!-- 状态 -->
                  <td class="whitespace-nowrap px-5 py-3.5">
                    <BaseBadge :type="row.status === 1 ? 'success' : 'default'" dot>
                      {{ row.status === 1 ? '启用' : '停用' }}
                    </BaseBadge>
                  </td>

                  <!-- 排序 -->
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 md:table-cell">{{ row.sort_order }}</td>

                  <!-- 操作 -->
                  <td class="px-5 py-3.5">
                    <div class="flex items-center justify-end gap-1.5 whitespace-nowrap">
                      <BaseButton size="sm" type="ghost" @click="openEdit(row)">
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" /></svg>
                        编辑
                      </BaseButton>
                      <BaseButton
                        v-if="row.status === 1"
                        size="sm"
                        type="ghost"
                        class="text-red-500 hover:bg-red-50 hover:text-red-600"
                        @click="confirmToggleStatus(row)"
                      >
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 10.5V6.75a4.5 4.5 0 119 0v3.75M3.75 21.75h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H3.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
                        停用
                      </BaseButton>
                      <BaseButton
                        v-else
                        size="sm"
                        type="ghost"
                        class="text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700"
                        @click="confirmToggleStatus(row)"
                      >
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 10.5V6.75a4.5 4.5 0 119 0v3.75M3.75 21.75h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H3.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
                        启用
                      </BaseButton>
                      <BaseButton size="sm" type="danger" @click="confirmDelete(row)">
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" /></svg>
                        删除
                      </BaseButton>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 列表底部统计条 -->
      <div v-if="!loading && list.length > 0" class="mt-5 flex items-center justify-between gap-3 rounded-xl border border-gray-100 bg-white px-4 py-3 shadow-sm">
        <p class="text-xs text-gray-500">共 {{ list.length }} 个科目</p>
      </div>
    </template>

    <!-- 新增 / 编辑弹窗 -->
    <BaseModal v-model="modalOpen" :title="modalTitle" width="max-w-lg">
      <div class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <BaseSelect v-model="form.level_id" label="所属等级" :disabled="isEdit">
            <option v-for="level in levels" :key="level.id" :value="level.id">
              {{ level.name }}
            </option>
          </BaseSelect>
          <BaseInput v-model="form.sort_order" label="排序（数字越小越靠前）" type="number" />
        </div>

        <div class="flex items-end gap-3">
          <div class="flex-1">
            <BaseInput v-model="form.name" label="科目名称" placeholder="例如：系统分析师" />
          </div>
          <BaseInput v-model="form.short_name" label="简称" placeholder="例如：系分" class="w-28" />
        </div>

        <div>
          <BaseInput v-model="form.icon" label="图标文字（可选）" placeholder="留空则显示科目首字" />
        </div>

        <BaseTextarea v-model="form.description" label="科目描述" :rows="3" placeholder="一句话介绍该科目" />

        <p v-if="formError" class="text-xs text-red-500">{{ formError }}</p>
      </div>
      <div class="mt-6 flex justify-end gap-2 border-t border-gray-50 pt-4">
        <BaseButton type="ghost" size="sm" @click="modalOpen = false">取消</BaseButton>
        <BaseButton size="sm" :loading="saving" @click="submit">
          {{ isEdit ? '保存修改' : '确认新增' }}
        </BaseButton>
      </div>
    </BaseModal>

    <!-- 通用确认弹窗：删除 / 启停 -->
    <ConfirmDialog
      ref="dialog"
      :title="dialogTitle"
      :description="dialogDescription"
      :danger="!!pendingDelete"
      :confirm-text="pendingDelete ? '确认删除' : '确认'"
      @confirm="onConfirmAction"
    />
  </div>
</template>

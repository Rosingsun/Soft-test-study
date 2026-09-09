<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  createAdminUser,
  deleteAdminUser,
  listAdminUsers,
  resetAdminUserPassword,
  setAdminUserStatus,
  updateAdminUser,
} from '@/api/admin'
import { getExamLevels, getSubjectsByLevel } from '@/api/subject'
import type { AdminUserListItem, CreateAdminUserReq, UpdateAdminUserReq } from '@/types/admin'
import type { ExamLevel, Subject } from '@/types/subject'
import { useAuthStore } from '@/stores/auth'
import { showToast } from '@/utils/toast'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

type RoleFilter = 'all' | 'student' | 'admin'
type StatusFilter = '' | '1' | '0'
type FormMode = 'create' | 'edit'

const list = ref<AdminUserListItem[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const error = ref('')

// 筛选
const roleFilter = ref<RoleFilter>('all')
const statusFilter = ref<StatusFilter>('')
const keyword = ref('')
const keywordApplied = ref('')

// 当前登录管理员（用于自我保护：不能禁用/删除自己、不能改自己的角色）
const auth = useAuthStore()
const currentUserId = computed(() => auth.user?.id ?? 0)

// 等级 / 科目级联数据
const levels = ref<ExamLevel[]>([])
const formSubjects = ref<Subject[]>([])

// 表单状态
const formOpen = ref(false)
const formMode = ref<FormMode>('create')
const submitting = ref(false)
const editTarget = ref<AdminUserListItem | null>(null)

const emptyForm = () => ({
  username: '',
  email: '',
  password: '',
  nickname: '',
  role: 'student',
  status: 1,
  level_id: 0,
  subject_id: 0,
})
const form = reactive<{ username: string; email: string; password: string; nickname: string; role: string; status: number; level_id: number; subject_id: number }>(emptyForm())

// 重置密码
const pwdOpen = ref(false)
const pwdTarget = ref<AdminUserListItem | null>(null)
const newPassword = ref('')
const pwdSubmitting = ref(false)

// 删除确认
const deleteTarget = ref<AdminUserListItem | null>(null)
const deleteDialog = ref<InstanceType<typeof ConfirmDialog> | null>(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

const roleTabs = [
  { label: '全部', value: 'all' },
  { label: '学员', value: 'student' },
  { label: '管理员', value: 'admin' },
]

const statusLabel = (s: number) =>
  s === 1 ? { label: '正常', type: 'success' as const } : { label: '已禁用', type: 'danger' as const }

const isSelf = (id: number) => id === currentUserId.value

// ============ 数据加载 ============

async function load() {
  loading.value = true
  error.value = ''
  try {
    const data = await listAdminUsers({
      role: roleFilter.value === 'all' ? undefined : roleFilter.value,
      status: statusFilter.value === '' ? undefined : Number(statusFilter.value),
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

async function loadLevels() {
  try {
    levels.value = await getExamLevels()
  } catch {
    levels.value = []
  }
}

async function loadSubjectsForLevel(levelId: number) {
  if (!levelId) {
    formSubjects.value = []
    form.subject_id = 0
    return
  }
  try {
    formSubjects.value = await getSubjectsByLevel(levelId)
  } catch {
    formSubjects.value = []
  }
  // 若当前已选科目不属于新等级，清空
  const cur = form.subject_id
  if (cur && !formSubjects.value.some(s => s.id === cur)) {
    form.subject_id = 0
  }
}

onMounted(() => {
  load()
  loadLevels()
})

// ============ 筛选 ============

function changeRole(v: string | number) {
  roleFilter.value = v as RoleFilter
  page.value = 1
  load()
}

function changeStatus() {
  page.value = 1
  load()
}

function applyKeyword() {
  keywordApplied.value = keyword.value.trim()
  page.value = 1
  load()
}

function resetFilters() {
  roleFilter.value = 'all'
  statusFilter.value = ''
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

// ============ 新建 / 编辑 ============

function openCreate() {
  formMode.value = 'create'
  editTarget.value = null
  Object.assign(form, emptyForm())
  formSubjects.value = []
  formOpen.value = true
}

function openEdit(row: AdminUserListItem) {
  formMode.value = 'edit'
  editTarget.value = row
  Object.assign(form, {
    username: row.username,
    email: row.email,
    password: '',
    nickname: row.nickname,
    role: row.role,
    status: row.status,
    level_id: row.level_id,
    subject_id: row.subject_id,
  })
  if (row.level_id) loadSubjectsForLevel(row.level_id)
  formOpen.value = true
}

watch(
  () => form.level_id,
  v => {
    if (formMode.value === 'create' || formMode.value === 'edit') {
      loadSubjectsForLevel(Number(v) || 0)
    }
  },
)

async function submitForm() {
  submitting.value = true
  try {
    if (formMode.value === 'create') {
      const body: CreateAdminUserReq = {
        username: form.username.trim(),
        email: form.email.trim(),
        password: form.password,
        nickname: form.nickname.trim(),
        role: form.role,
        status: form.status,
        level_id: form.level_id || undefined,
        subject_id: form.subject_id || undefined,
      }
      await createAdminUser(body)
      showToast('用户创建成功', 'success')
    } else if (editTarget.value) {
      const body: UpdateAdminUserReq = {}
      if (form.nickname !== editTarget.value.nickname) body.nickname = form.nickname.trim()
      if (form.email.trim() !== editTarget.value.email) body.email = form.email.trim()
      if (form.role !== editTarget.value.role) body.role = form.role
      if (form.status !== editTarget.value.status) body.status = form.status
      if (form.level_id !== editTarget.value.level_id) body.level_id = form.level_id
      if (form.subject_id !== editTarget.value.subject_id) body.subject_id = form.subject_id
      await updateAdminUser(editTarget.value.id, body)
      showToast('用户更新成功', 'success')
    }
    formOpen.value = false
    await load()
  } catch {
    // 请求层已统一 toast
  } finally {
    submitting.value = false
  }
}

// ============ 启 / 禁用 ============

async function toggleStatus(row: AdminUserListItem) {
  const target = row.status === 1 ? 0 : 1
  try {
    await setAdminUserStatus(row.id, target)
    showToast(target === 1 ? '账号已启用' : '账号已禁用', 'success')
    await load()
  } catch {
    // 请求层已统一 toast
  }
}

// ============ 重置密码 ============

function openReset(row: AdminUserListItem) {
  pwdTarget.value = row
  newPassword.value = ''
  pwdOpen.value = true
}

async function submitReset() {
  if (!pwdTarget.value) return
  pwdSubmitting.value = true
  try {
    await resetAdminUserPassword(pwdTarget.value.id, newPassword.value)
    showToast('密码重置成功', 'success')
    pwdOpen.value = false
  } catch {
    // 请求层已统一 toast
  } finally {
    pwdSubmitting.value = false
  }
}

// ============ 删除 ============

function openDelete(row: AdminUserListItem) {
  deleteTarget.value = row
  deleteDialog.value?.open()
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  try {
    await deleteAdminUser(deleteTarget.value.id)
    showToast('用户已删除', 'success')
    deleteTarget.value = null
    // 若当前页删空则回退一页
    if (list.value.length === 1 && page.value > 1) page.value -= 1
    await load()
  } catch {
    // 请求层已统一 toast
  }
}

// ============ 展示辅助 ============

function displayName(row: AdminUserListItem) {
  return row.nickname || row.username
}

function initial(row: AdminUserListItem) {
  return displayName(row).slice(0, 1).toUpperCase()
}
</script>

<template>
  <div>
    <BasePageHeader title="用户管理" subtitle="查看与管理平台注册用户，支持新建、编辑、启禁、重置密码与删除">
      <template #actions>
        <BaseButton type="secondary" @click="load">刷新</BaseButton>
        <BaseButton @click="openCreate">新建用户</BaseButton>
      </template>
    </BasePageHeader>

    <!-- 筛选区 -->
    <div class="mb-5 flex flex-col gap-3 rounded-xl border border-gray-100 bg-white p-4 shadow-sm sm:flex-row sm:items-center sm:justify-between">
      <BaseTabs v-model="roleFilter" variant="pill" size="sm" :tabs="roleTabs" @update:model-value="changeRole" />
      <div class="flex w-full flex-col gap-2 sm:w-auto sm:flex-row sm:items-center">
        <div class="flex items-center gap-2">
          <div class="w-28">
            <BaseSelect v-model="statusFilter" @change="changeStatus">
              <option value="">全部状态</option>
              <option value="1">正常</option>
              <option value="0">已禁用</option>
            </BaseSelect>
          </div>
          <div class="w-52">
            <BaseInput v-model="keyword" placeholder="用户名 / 昵称 / 邮箱" @keyup.enter="applyKeyword" />
          </div>
          <BaseButton size="sm" @click="applyKeyword">搜索</BaseButton>
          <BaseButton type="ghost" size="sm" @click="resetFilters">重置</BaseButton>
        </div>
      </div>
    </div>

    <!-- 加载失败 -->
    <BaseCard v-if="error">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </BaseCard>

    <!-- 空列表 -->
    <BaseCard v-else-if="!loading && list.length === 0">
      <BaseEmpty title="暂无符合条件的用户" description="调整筛选条件，或新建一个用户">
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="openCreate">新建用户</BaseButton>
        </template>
      </BaseEmpty>
    </BaseCard>

    <!-- 列表 -->
    <template v-else>
      <div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
        <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead>
              <tr class="border-b border-gray-100 bg-gray-50/70 text-xs font-semibold uppercase tracking-wider text-gray-500">
                <th scope="col" class="whitespace-nowrap px-5 py-3">用户</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 md:table-cell">角色</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3">状态</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 lg:table-cell">报考科目</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 xl:table-cell">邮箱验证</th>
                <th scope="col" class="hidden whitespace-nowrap px-5 py-3 md:table-cell">注册时间</th>
                <th scope="col" class="whitespace-nowrap px-5 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <template v-if="loading">
                <tr v-for="i in 5" :key="i">
                  <td class="px-5 py-4">
                    <div class="flex items-center gap-3">
                      <div class="skeleton h-9 w-9 rounded-full" />
                      <div class="space-y-1.5"><div class="skeleton h-3.5 w-28" /><div class="skeleton h-3 w-20" /></div>
                    </div>
                  </td>
                  <td class="hidden px-5 py-4 md:table-cell"><div class="skeleton h-5 w-14 rounded-full" /></td>
                  <td class="px-5 py-4"><div class="skeleton h-5 w-16 rounded-full" /></td>
                  <td class="hidden px-5 py-4 lg:table-cell"><div class="skeleton h-3.5 w-32" /></td>
                  <td class="hidden px-5 py-4 xl:table-cell"><div class="skeleton h-5 w-16 rounded-full" /></td>
                  <td class="hidden px-5 py-4 md:table-cell"><div class="skeleton h-3.5 w-24" /></td>
                  <td class="px-5 py-4">
                    <div class="ml-auto flex items-center justify-end gap-1.5">
                      <div class="skeleton h-8 w-12 rounded-lg" />
                      <div class="skeleton h-8 w-12 rounded-lg" />
                    </div>
                  </td>
                </tr>
              </template>
              <template v-else>
                <tr v-for="row in list" :key="row.id" class="group transition-colors duration-200 hover:bg-indigo-50/30">
                  <td class="px-5 py-3.5">
                    <div class="flex items-center gap-3">
                      <div class="bg-brand-gradient flex h-9 w-9 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white shadow-sm">
                        {{ initial(row) }}
                      </div>
                      <div class="min-w-0">
                        <p class="truncate font-medium text-gray-900">
                          {{ displayName(row) }}
                          <span v-if="isSelf(row.id)" class="ml-1 text-xs font-normal text-indigo-500">(我)</span>
                        </p>
                        <p class="truncate text-xs text-gray-500">@{{ row.username }}</p>
                      </div>
                    </div>
                  </td>
                  <td class="hidden whitespace-nowrap px-5 py-3.5 md:table-cell">
                    <BaseBadge :type="row.role === 'admin' ? 'info' : 'default'">
                      {{ row.role === 'admin' ? '管理员' : '学员' }}
                    </BaseBadge>
                  </td>
                  <td class="whitespace-nowrap px-5 py-3.5">
                    <BaseBadge :type="statusLabel(row.status).type" dot>{{ statusLabel(row.status).label }}</BaseBadge>
                  </td>
                  <td class="hidden max-w-[14rem] px-5 py-3.5 lg:table-cell">
                    <p class="truncate text-gray-700">
                      <template v-if="row.subject_name">{{ row.level_name }} · {{ row.subject_name }}</template>
                      <template v-else-if="row.level_name">{{ row.level_name }}</template>
                      <span v-else class="text-gray-400">未报考</span>
                    </p>
                  </td>
                  <td class="hidden whitespace-nowrap px-5 py-3.5 xl:table-cell">
                    <BaseBadge :type="row.email_verified ? 'success' : 'warning'">
                      {{ row.email_verified ? '已验证' : '未验证' }}
                    </BaseBadge>
                  </td>
                  <td class="hidden whitespace-nowrap px-5 py-3.5 text-gray-500 md:table-cell">{{ row.created_at }}</td>
                  <td class="px-5 py-3.5">
                    <div class="flex items-center justify-end gap-1.5 whitespace-nowrap">
                      <BaseButton size="sm" type="ghost" @click="openEdit(row)">编辑</BaseButton>
                      <BaseButton v-if="!isSelf(row.id)" size="sm" type="ghost" @click="openReset(row)">重置密码</BaseButton>
                      <BaseButton
                        v-if="!isSelf(row.id)"
                        size="sm"
                        :type="row.status === 1 ? 'secondary' : 'success'"
                        @click="toggleStatus(row)"
                      >
                        {{ row.status === 1 ? '禁用' : '启用' }}
                      </BaseButton>
                      <BaseButton v-if="!isSelf(row.id)" size="sm" type="danger" @click="openDelete(row)">删除</BaseButton>
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
        <p class="text-xs text-gray-500">共 {{ total }} 位用户</p>
        <div class="flex items-center gap-2">
          <BaseButton size="sm" type="secondary" :disabled="page <= 1" @click="goPage(page - 1)">上一页</BaseButton>
          <span class="text-xs font-medium text-gray-600">第 {{ page }} / {{ totalPages }} 页</span>
          <BaseButton size="sm" type="secondary" :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</BaseButton>
        </div>
      </div>
    </template>

    <!-- 新建 / 编辑 用户 -->
    <BaseModal v-model="formOpen" :title="formMode === 'create' ? '新建用户' : '编辑用户'" width="max-w-lg">
      <form class="space-y-4" @submit.prevent="submitForm">
        <div v-if="formMode === 'create'">
          <BaseInput v-model="form.username" label="用户名 *" placeholder="字母/数字/下划线/中文，3-50 位" />
        </div>
        <div v-else class="rounded-lg border border-gray-100 bg-gray-50/60 px-3 py-2 text-sm text-gray-600">
          正在编辑用户：<b class="text-gray-800">{{ displayName(editTarget!) }}</b>（@{{ form.username }}）
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <BaseInput v-model="form.email" label="邮箱 *" placeholder="user@example.com" />
          <div v-if="formMode === 'create'">
            <BaseInput v-model="form.password" label="初始密码 *" type="password" toggleable placeholder="至少 8 位且含数字" />
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <BaseInput v-model="form.nickname" label="昵称" placeholder="选填" />
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">角色</label>
            <BaseSelect v-model="form.role" :disabled="formMode === 'edit' && !!editTarget && isSelf(editTarget.id)">
              <option value="student">学员</option>
              <option value="admin">管理员</option>
            </BaseSelect>
          </div>
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">报考等级</label>
            <BaseSelect v-model="form.level_id">
              <option :value="0">未选择</option>
              <option v-for="l in levels" :key="l.id" :value="l.id">{{ l.name }}</option>
            </BaseSelect>
          </div>
          <div>
            <label class="mb-1.5 block text-sm font-medium text-gray-700">报考科目</label>
            <BaseSelect v-model="form.subject_id" :disabled="!form.level_id">
              <option :value="0">{{ form.level_id ? '选择科目' : '请先选择等级' }}</option>
              <option v-for="s in formSubjects" :key="s.id" :value="s.id">{{ s.name }}</option>
            </BaseSelect>
          </div>
        </div>

        <p v-if="formMode === 'edit' && isSelf(editTarget!.id)" class="text-xs text-gray-400">
          当前登录的账号不允许被禁用或改为非管理员。
        </p>

        <div class="flex justify-end gap-2 border-t border-gray-50 pt-4">
          <BaseButton type="ghost" size="sm" @click="formOpen = false">取消</BaseButton>
          <BaseButton size="sm" type="primary" :loading="submitting" @click="submitForm">
            {{ formMode === 'create' ? '创建用户' : '保存修改' }}
          </BaseButton>
        </div>
      </form>
    </BaseModal>

    <!-- 重置密码 -->
    <BaseModal v-model="pwdOpen" title="重置密码" width="max-w-sm">
      <p v-if="pwdTarget" class="mb-4 text-sm text-gray-500">
        为 <b class="text-gray-800">{{ displayName(pwdTarget) }}</b>（@{{ pwdTarget.username }}）设置新密码，用户下次登录需使用新密码。
      </p>
      <BaseInput v-model="newPassword" label="新密码 *" type="password" toggleable placeholder="至少 8 位且含数字" @keyup.enter="submitReset" />
      <div class="mt-6 flex justify-end gap-2 border-t border-gray-50 pt-4">
        <BaseButton type="ghost" size="sm" @click="pwdOpen = false">取消</BaseButton>
        <BaseButton size="sm" :loading="pwdSubmitting" @click="submitReset">确认重置</BaseButton>
      </div>
    </BaseModal>

    <!-- 删除确认 -->
    <ConfirmDialog
      ref="deleteDialog"
      title="删除用户"
      :description="deleteTarget ? `确定删除用户「${displayName(deleteTarget)}」吗？该用户的全部学习、答题数据将一并清除，且不可恢复。` : ''"
      danger
      confirm-text="删除"
      @confirm="confirmDelete"
    />
  </div>
</template>

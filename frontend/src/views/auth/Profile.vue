<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { changePassword } from '@/api/auth'
import { getStatsOverview, getDailyStats } from '@/api/stats'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseLoading from '@/components/common/BaseLoading.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { StatsOverviewResp, DailyStatsResp } from '@/types/stats'

const auth = useAuthStore()
const overview = ref<StatsOverviewResp | null>(null)
const dailyStats = ref<DailyStatsResp[]>([])
const loading = ref(true)
const error = ref('')

const showEdit = ref(false)
const nickname = ref('')
const avatar = ref('')
const difficulty = ref('')
const saving = ref(false)

const showPwd = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const pwdSaving = ref(false)
const pwdError = ref('')

const sortedDaily = computed(() => [...dailyStats.value].sort((a, b) => b.date.localeCompare(a.date)))

function openEdit() {
  nickname.value = auth.user?.nickname || ''
  avatar.value = auth.user?.avatar || ''
  difficulty.value = auth.user?.difficulty || ''
  showEdit.value = true
}

async function saveProfile() {
  if (!nickname.value.trim()) return
  saving.value = true
  try {
    await auth.updateProfile({
      nickname: nickname.value.trim(),
      avatar: avatar.value.trim() || undefined,
      difficulty: difficulty.value || null,
    })
    showEdit.value = false
  } catch {
    // toast 由 request 层提示
  } finally {
    saving.value = false
  }
}

async function savePassword() {
  pwdError.value = ''
  if (!oldPassword.value || !newPassword.value) {
    pwdError.value = '请输入原密码和新密码'
    return
  }
  if (newPassword.value.length < 6) {
    pwdError.value = '新密码至少 6 位'
    return
  }
  if (newPassword.value !== confirmPassword.value) {
    pwdError.value = '两次输入的新密码不一致'
    return
  }
  pwdSaving.value = true
  try {
    await changePassword({ old_password: oldPassword.value, new_password: newPassword.value })
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    showPwd.value = false
  } catch {
    // toast 由 request 层提示
  } finally {
    pwdSaving.value = false
  }
}

const difficultyLabels: Record<string, string> = { easy: '简单', medium: '中等', hard: '困难' }

onMounted(async () => {
  try {
    const [ov, daily] = await Promise.all([getStatsOverview(), getDailyStats(30)])
    overview.value = ov
    dailyStats.value = daily || []
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="mx-auto max-w-4xl">
    <BasePageHeader title="个人中心" subtitle="管理个人信息与账号安全" />

    <div v-if="loading">
      <BaseLoading />
    </div>

    <div v-else-if="error" class="rounded-lg bg-white p-6 shadow-sm">
      <p class="text-red-500">{{ error }}</p>
    </div>

    <template v-else-if="auth.user">
      <!-- 个人信息卡片 -->
      <div class="mb-6 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm">
        <div class="h-20 bg-gradient-to-r from-indigo-600 to-indigo-400" />
        <div class="flex flex-col items-center gap-4 px-6 pb-6 sm:flex-row sm:items-end">
          <div class="-mt-10 flex h-20 w-20 shrink-0 items-center justify-center overflow-hidden rounded-full border-4 border-white bg-indigo-100 text-2xl font-bold text-indigo-600 shadow-sm">
            <img v-if="auth.user.avatar" :src="auth.user.avatar" :alt="auth.user.nickname" class="h-full w-full object-cover" />
            <template v-else>{{ (auth.user.nickname || auth.user.username).charAt(0).toUpperCase() }}</template>
          </div>
          <div class="flex-1 text-center sm:text-left">
            <h1 class="text-xl font-bold text-gray-900">{{ auth.user.nickname || auth.user.username }}</h1>
            <p class="text-sm text-gray-500">{{ auth.user.email }}</p>
            <div class="mt-2 flex flex-wrap justify-center gap-2 text-xs sm:justify-start">
              <span class="rounded-md bg-indigo-50 px-2 py-0.5 text-indigo-600">{{ auth.user.level_name || '-' }}</span>
              <span class="rounded-md bg-emerald-50 px-2 py-0.5 text-emerald-600">{{ auth.user.subject_name || '-' }}</span>
              <span class="rounded-md bg-gray-100 px-2 py-0.5 text-gray-600">{{ auth.user.role === 'admin' ? '管理员' : '学生' }}</span>
              <span v-if="auth.user.difficulty" class="rounded-md bg-gray-100 px-2 py-0.5 text-gray-600">
                难度偏好：{{ difficultyLabels[auth.user.difficulty] || auth.user.difficulty }}
              </span>
            </div>
          </div>
          <div class="flex gap-2">
            <BaseButton type="secondary" size="sm" @click="openEdit">编辑资料</BaseButton>
            <BaseButton type="secondary" size="sm" @click="showPwd = true">修改密码</BaseButton>
          </div>
        </div>
      </div>

      <!-- 答题概览 -->
      <div v-if="overview" class="mb-6 grid grid-cols-2 gap-4 sm:grid-cols-4">
        <div class="rounded-lg border border-gray-200 bg-white p-4 text-center shadow-sm">
          <p class="text-2xl font-bold text-indigo-600">{{ overview.total_practiced }}</p>
          <p class="text-xs text-gray-500">总答题数</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 text-center shadow-sm">
          <p class="text-2xl font-bold" :class="overview.accuracy >= 60 ? 'text-emerald-500' : 'text-yellow-500'">
            {{ overview.accuracy.toFixed(1) }}%
          </p>
          <p class="text-xs text-gray-500">正确率</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 text-center shadow-sm">
          <p class="text-2xl font-bold text-emerald-600">{{ overview.total_correct }}</p>
          <p class="text-xs text-gray-500">正确数量</p>
        </div>
        <div class="rounded-lg border border-gray-200 bg-white p-4 text-center shadow-sm">
          <p class="text-2xl font-bold text-red-500">{{ overview.total_practiced - overview.total_correct }}</p>
          <p class="text-xs text-gray-500">错误数量</p>
        </div>
      </div>

      <!-- 每日学习记录 -->
      <div class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
        <h2 class="mb-4 text-base font-semibold text-gray-900">每日学习记录</h2>
        <div v-if="sortedDaily.length === 0" class="py-8 text-center text-sm text-gray-400">
          暂无记录，快去练习吧
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-xs text-gray-500">
                <th class="pb-2 font-medium">日期</th>
                <th class="pb-2 font-medium">答题总数</th>
                <th class="pb-2 font-medium">正确数量</th>
                <th class="pb-2 font-medium">错误数量</th>
                <th class="pb-2 font-medium">正确率</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in sortedDaily.slice(0, 30)" :key="row.date" class="border-b border-gray-50">
                <td class="py-2.5 text-gray-700">{{ row.date }}</td>
                <td class="py-2.5 text-gray-700">{{ row.total_count }}</td>
                <td class="py-2.5 text-emerald-600">{{ row.correct_count }}</td>
                <td class="py-2.5 text-red-500">{{ row.incorrect_count }}</td>
                <td class="py-2.5">
                  <span
                    class="rounded-md px-2 py-0.5 text-xs font-medium"
                    :class="row.total_count > 0
                      ? (row.correct_count / row.total_count * 100 >= 60 ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-600')
                      : 'text-gray-400'"
                  >
                    {{ row.total_count > 0 ? (row.correct_count / row.total_count * 100).toFixed(0) + '%' : '-' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <!-- 编辑资料弹窗 -->
    <BaseModal v-model="showEdit" title="编辑资料">
      <div class="space-y-4">
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">昵称</label>
          <BaseInput v-model="nickname" placeholder="请输入昵称" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">头像 URL</label>
          <BaseInput v-model="avatar" placeholder="请输入头像图片链接（可选）" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">难度偏好</label>
          <div class="flex gap-2">
            <button
              v-for="(label, key) in difficultyLabels"
              :key="key"
              class="flex-1 cursor-pointer rounded-lg border px-3 py-2 text-sm transition-colors"
              :class="difficulty === key
                ? 'border-indigo-500 bg-indigo-50 text-indigo-600'
                : 'border-gray-300 bg-white text-gray-600 hover:bg-gray-50'"
              @click="difficulty = key"
            >
              {{ label }}
            </button>
          </div>
        </div>
        <div class="flex justify-end gap-2">
          <BaseButton type="secondary" @click="showEdit = false">取消</BaseButton>
          <BaseButton :loading="saving" @click="saveProfile">保存</BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- 修改密码弹窗 -->
    <BaseModal v-model="showPwd" title="修改密码">
      <div class="space-y-4">
        <p v-if="pwdError" class="rounded-lg bg-red-50 px-4 py-2 text-sm text-red-600">{{ pwdError }}</p>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">原密码</label>
          <BaseInput v-model="oldPassword" type="password" placeholder="请输入原密码" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">新密码</label>
          <BaseInput v-model="newPassword" type="password" placeholder="至少 6 位" />
        </div>
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">确认新密码</label>
          <BaseInput v-model="confirmPassword" type="password" placeholder="请再次输入新密码" />
        </div>
        <div class="flex justify-end gap-2">
          <BaseButton type="secondary" @click="showPwd = false">取消</BaseButton>
          <BaseButton :loading="pwdSaving" @click="savePassword">确认修改</BaseButton>
        </div>
      </div>
    </BaseModal>
  </div>
</template>

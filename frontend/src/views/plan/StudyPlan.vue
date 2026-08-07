<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { listStudyPlans, createStudyPlan, deleteStudyPlan } from '@/api/studyPlan'
import { useSubjectStore } from '@/stores/subject'
import type { StudyPlanResp } from '@/types/studyPlan'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCard from '@/components/common/BaseCard.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import { showToast } from '@/utils/toast'

const router = useRouter()
const subjectStore = useSubjectStore()

const list = ref<StudyPlanResp[]>([])
const loading = ref(true)
const error = ref('')
const showCreate = ref(false)
const creating = ref(false)
const form = ref({ title: '', subject_id: 0, daily_goal: 20, start_date: '', end_date: '' })
const formError = ref('')
const deleting = ref<StudyPlanResp | null>(null)
const confirmRef = ref<InstanceType<typeof ConfirmDialog> | null>(null)

const subjects = computed(() => subjectStore.subjects)

const activeCount = computed(() => list.value.filter(p => p.status === 1).length)

function fmtDate(d: string) {
  const parts = d.split('-')
  return parts.length === 3 ? `${parts[0]}.${parts[1]}.${parts[2]}` : d
}

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    list.value = await listStudyPlans()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(load)

function openCreate() {
  form.value = { title: '', subject_id: 0, daily_goal: 20, start_date: todayStr(), end_date: todayStr() }
  formError.value = ''
  showCreate.value = true
}

async function handleCreate() {
  formError.value = ''
  if (!form.value.title.trim()) {
    formError.value = '请输入计划名称'
    return
  }
  if (!form.value.daily_goal || form.value.daily_goal < 1) {
    formError.value = '每日目标题数至少为 1'
    return
  }
  if (!form.value.start_date || !form.value.end_date) {
    formError.value = '请选择起止日期'
    return
  }
  if (form.value.end_date < form.value.start_date) {
    formError.value = '结束日期不能早于开始日期'
    return
  }
  creating.value = true
  try {
    await createStudyPlan({
      title: form.value.title.trim(),
      subject_id: form.value.subject_id || undefined,
      daily_goal: form.value.daily_goal,
      start_date: form.value.start_date,
      end_date: form.value.end_date,
    })
    showToast('计划创建成功', 'success')
    showCreate.value = false
    await load()
  } catch {
    // toast 由请求层提示
  } finally {
    creating.value = false
  }
}

function confirmDelete(p: StudyPlanResp) {
  deleting.value = p
  confirmRef.value?.open()
}

async function handleDelete() {
  if (!deleting.value) return
  try {
    await deleteStudyPlan(deleting.value.id)
    showToast('计划已删除', 'success')
    list.value = list.value.filter(p => p.id !== deleting.value!.id)
  } catch {
    // toast 由请求层提示
  } finally {
    deleting.value = null
  }
}

function subjectName(id: number) {
  if (!id) return '全科目'
  return subjects.value.find(s => s.id === id)?.short_name || subjects.value.find(s => s.id === id)?.name || '科目'
}

function statusBadge(p: StudyPlanResp) {
  if (p.status === 2) return { label: '已完成', type: 'success' as const }
  if (p.status === 0) return { label: '已放弃', type: 'default' as const }
  return { label: '进行中', type: 'info' as const }
}

function pctText(p: StudyPlanResp) {
  return `${Math.min(100, p.completion_pct).toFixed(0)}%`
}

function pctClass(p: StudyPlanResp) {
  if (p.completion_pct >= 100) return 'bg-emerald-500'
  if (p.completion_pct >= 60) return 'bg-indigo-500'
  return 'bg-amber-500'
}
</script>

<template>
  <div>
    <BasePageHeader title="学习计划" :subtitle="activeCount ? `进行中计划 ${activeCount} 个` : '制定备考计划，稳步提升'">
      <template #actions>
        <BaseButton @click="openCreate">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
          </svg>
          新建计划
        </BaseButton>
      </template>
    </BasePageHeader>

    <BaseSkeleton v-if="loading" variant="list" :count="3" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else-if="list.length === 0" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="还没有学习计划" description="设定每日目标，打卡与练习都会计入进度">
        <template #action>
          <BaseButton size="sm" class="mt-4" @click="openCreate">立即创建</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <div v-else class="grid gap-4 lg:grid-cols-2">
      <BaseCard v-for="p in list" :key="p.id" hover class="flex flex-col gap-4 p-6">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h3 class="truncate text-base font-semibold tracking-tight text-gray-900">{{ p.title }}</h3>
              <BaseBadge :type="statusBadge(p).type">{{ statusBadge(p).label }}</BaseBadge>
            </div>
            <p class="mt-1 text-xs text-gray-400">
              {{ subjectName(p.subject_id) }} · {{ fmtDate(p.start_date) }} ~ {{ fmtDate(p.end_date) }}
              <template v-if="p.status === 1"> · 剩余 {{ p.remaining_days }} 天</template>
            </p>
          </div>
          <div class="flex shrink-0 gap-1.5">
            <BaseButton type="secondary" size="sm" @click="confirmDelete(p)">删除</BaseButton>
          </div>
        </div>

        <!-- 今日进度 -->
        <div>
          <div class="mb-1.5 flex items-center justify-between text-xs">
            <span class="text-gray-500">今日进度</span>
            <span class="font-medium text-gray-700">{{ Math.min(p.today_done, p.today_goal) }}/{{ p.today_goal }} 题</span>
          </div>
          <div class="h-2 overflow-hidden rounded-full bg-gray-100">
            <div
              class="h-full rounded-full bg-gradient-to-r from-indigo-500 to-violet-500 transition-all duration-500"
              :style="{ width: Math.min(100, (p.today_done / p.today_goal) * 100) + '%' }"
            />
          </div>
        </div>

        <!-- 总体进度 -->
        <div class="flex items-center justify-between rounded-xl bg-gray-50 px-4 py-3">
          <div class="min-w-0 flex-1">
            <div class="mb-1 flex items-center justify-between text-xs">
              <span class="text-gray-500">总体进度</span>
              <span class="font-medium text-gray-700">{{ p.overall_done }}/{{ p.overall_goal }} 题 · {{ pctText(p) }}</span>
            </div>
            <div class="h-1.5 overflow-hidden rounded-full bg-gray-200">
              <div class="h-full rounded-full transition-all duration-500" :class="pctClass(p)" :style="{ width: Math.min(100, p.completion_pct) + '%' }" />
            </div>
          </div>
          <BaseButton type="secondary" size="sm" class="ml-4" @click="router.push('/practice/random')">
            去练习
          </BaseButton>
        </div>
      </BaseCard>
    </div>

    <!-- 新建计划弹窗 -->
    <BaseModal v-model="showCreate" title="新建学习计划">
      <div class="space-y-4">
        <BaseInput v-model="form.title" label="计划名称" placeholder="如：软设冲刺 45 天" />
        <div>
          <label class="mb-1.5 block text-sm font-medium text-gray-700">目标科目</label>
          <BaseSelect v-model="form.subject_id">
            <option :value="0">全科目（不区分）</option>
            <option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.short_name || s.name }}</option>
          </BaseSelect>
        </div>
        <BaseInput v-model.number="form.daily_goal" label="每日目标题数" type="number" min="1" placeholder="如：20" />
        <div class="grid grid-cols-2 gap-3">
          <BaseInput v-model="form.start_date" label="开始日期" type="date" />
          <BaseInput v-model="form.end_date" label="结束日期" type="date" />
        </div>
        <p v-if="formError" class="text-xs text-red-500">{{ formError }}</p>
        <div class="flex justify-end gap-2 border-t border-gray-50 pt-4">
          <BaseButton type="secondary" @click="showCreate = false">取消</BaseButton>
          <BaseButton :loading="creating" @click="handleCreate">创建计划</BaseButton>
        </div>
      </div>
    </BaseModal>

    <ConfirmDialog
      ref="confirmRef"
      title="删除学习计划"
      :description="deleting ? `确定删除「${deleting.title}」吗？` : ''"
      confirm-text="删除"
      danger
      @confirm="handleDelete"
    />
  </div>
</template>

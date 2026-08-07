<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { getRanking } from '@/api/ranking'
import { useSubjectStore } from '@/stores/subject'
import type { RankingResp, RankingCategory, RankingMetric, DistributionItem } from '@/types/ranking'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseTabs from '@/components/common/BaseTabs.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseSkeleton from '@/components/common/BaseSkeleton.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'

const subjectStore = useSubjectStore()

const category = ref<RankingCategory>('practice')
const metric = ref<RankingMetric>('accuracy')
const subjectId = ref<number>(0)
const data = ref<RankingResp | null>(null)
const loading = ref(true)
const error = ref('')

const categoryTabs = [
  { label: '训练排名', value: 'practice' as RankingCategory },
  { label: '模拟考试排名', value: 'exam' as RankingCategory },
  { label: '打卡排名', value: 'checkin' as RankingCategory },
]

const metricTabs = computed(() => {
  if (category.value === 'checkin') {
    return [
      { label: '最高连续天数', value: 'streak' as RankingMetric },
      { label: '打卡正确率', value: 'accuracy' as RankingMetric },
    ]
  }
  return [
    { label: '答题正确率', value: 'accuracy' as RankingMetric },
    { label: '预估分数', value: 'score' as RankingMetric },
  ]
})

// 切榜时修正维度默认值
watch(category, (c) => {
  if (c === 'checkin') {
    if (metric.value !== 'streak' && metric.value !== 'accuracy') metric.value = 'streak'
  } else if (metric.value !== 'accuracy' && metric.value !== 'score') {
    metric.value = 'accuracy'
  }
})

const showSubjectFilter = computed(() => category.value !== 'checkin')

function categoryLabel() {
  if (category.value === 'checkin') return '打卡排名'
  if (category.value === 'exam') return '模考排名'
  return '训练排名'
}

function metricLabel() {
  if (metric.value === 'streak') return '最高连续天数'
  if (metric.value === 'score') return '预估分数'
  return '答题正确率'
}

// 按 metric 格式化数值（用于均值/Top10 等指标）
function formatValue(v: number) {
  if (metric.value === 'streak') return `${v.toFixed(0)} 天`
  if (metric.value === 'score') return `${v.toFixed(1)} 分`
  return `${v.toFixed(1)}%`
}

function valueText(r: RankingResp) {
  if (!r.my) return '—'
  if (r.metric === 'streak') return `${Math.round(r.my.value)} 天`
  if (r.metric === 'score') return `${r.my.value.toFixed(1)} 分`
  return `${r.my.value.toFixed(1)}%`
}

function rankClass(rank: number, total: number) {
  if (total === 0) return 'text-gray-300'
  if (rank === 1) return 'text-amber-300'
  if (rank <= Math.max(1, Math.ceil(total * 0.1))) return 'text-emerald-300'
  return 'text-white'
}

// 击败比例（= 100 - percentile，percentile 是「被超过」的百分比口径，需确认；这里反向使用 = 排名靠前比例）
function beatRatio(): number {
  if (!data.value?.my) return 0
  return Math.max(0, 100 - data.value.my.percentile)
}

// 直方图相关
const maxCount = computed(() => {
  if (!data.value?.distribution?.length) return 0
  return Math.max(...data.value.distribution.map(d => d.count), 1)
})

function barHeight(item: DistributionItem) {
  if (item.count === 0) return 4
  return Math.max(10, (item.count / maxCount.value) * 88)
}

// 柱体填充色：当前用户用品牌渐变；其他用同色色阶按人数集中度区分
function barFillClass(item: DistributionItem) {
  if (item.is_mine) {
    return 'bg-gradient-to-b from-indigo-500 via-indigo-500 to-violet-500 ring-2 ring-white/70 shadow-lg shadow-indigo-500/40'
  }
  if (item.count === 0) return 'bg-gray-100'
  const ratio = item.count / maxCount.value
  if (ratio > 0.7) return 'bg-indigo-400'
  if (ratio > 0.4) return 'bg-indigo-300'
  if (ratio > 0.2) return 'bg-indigo-200'
  return 'bg-indigo-100'
}

function barTextClass(item: DistributionItem) {
  if (item.is_mine) return 'text-indigo-700'
  if (item.count === 0) return 'text-gray-300'
  const ratio = item.count / maxCount.value
  if (ratio > 0.7) return 'text-indigo-700'
  if (ratio > 0.4) return 'text-indigo-600'
  if (ratio > 0.2) return 'text-indigo-500'
  return 'text-indigo-400'
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    data.value = await getRanking(category.value, metric.value, showSubjectFilter.value ? subjectId.value : 0)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  if (!subjectStore.subjects.length) {
    subjectStore.fetchSubjectsByLevel(1).catch(() => {})
  }
  load()
})

watch([category, metric, subjectId], load)
</script>

<template>
  <div>
    <BasePageHeader title="成绩排行" subtitle="仅展示你的成绩与全平台名次，保护个人隐私" />

    <!-- 切换 + 筛选 -->
    <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <BaseTabs v-model="category" :tabs="categoryTabs" class="max-w-xl" />
      <div v-if="showSubjectFilter" class="w-full sm:w-56">
        <BaseSelect v-model="subjectId">
          <option :value="0">全部科目</option>
          <option v-for="s in subjectStore.subjects" :key="s.id" :value="s.id">{{ s.short_name || s.name }}</option>
        </BaseSelect>
      </div>
    </div>

    <BaseTabs v-model="metric" :tabs="metricTabs" class="mb-6 max-w-lg" />

    <BaseSkeleton v-if="loading" variant="detail" />

    <div v-else-if="error" class="rounded-xl border border-gray-100 bg-white shadow-sm">
      <BaseEmpty title="加载失败" :description="error">
        <template #action>
          <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
        </template>
      </BaseEmpty>
    </div>

    <template v-else-if="data">
      <!-- 主区：名次大卡 + 分布直方图（12 栅格） -->
      <div class="mb-6 grid gap-5 lg:grid-cols-12">
        <!-- 我的名次（5/12） -->
        <div class="lg:col-span-5">
          <div class="bg-brand-gradient relative flex h-full flex-col overflow-hidden rounded-2xl p-6 text-white shadow-lg shadow-indigo-600/20 sm:p-7">
            <div class="pointer-events-none absolute -right-10 -top-10 h-44 w-44 rounded-full bg-white/10" />
            <div class="pointer-events-none absolute -bottom-20 right-16 h-56 w-56 rounded-full bg-white/5" />

            <div class="relative flex items-center gap-2 text-sm text-indigo-100">
              <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16.5 18.75h-9m9 0a3 3 0 013 3h-15a3 3 0 013-3m9 0v-4.875A3.375 3.375 0 0015.375 10.5h-6.75A3.375 3.375 0 005.25 13.875V18.75m9-12.75h.008v.008H15V6zm-3 0h.008v.008H12V6zm-3 0h.008v.008H9V6z" />
              </svg>
              <span>{{ categoryLabel() }} · {{ metricLabel() }}</span>
            </div>

            <div class="relative mt-6 flex flex-wrap items-end gap-x-4 gap-y-2">
              <div class="flex items-baseline gap-2">
                <span
                  :class="data.my ? rankClass(data.my.rank, data.my.total_participants) : 'text-gray-300'"
                  class="text-6xl font-bold tracking-tight tabular-nums"
                >
                  {{ data.my ? data.my.rank : '—' }}
                </span>
                <span class="text-base font-medium text-white/80">/ {{ data.my ? data.my.total_participants : 0 }} 人</span>
              </div>
              <BaseBadge v-if="data.my && data.my.rank === 1" type="warning">🥇 冠军</BaseBadge>
              <BaseBadge v-else-if="data.my && data.my.rank <= Math.max(1, Math.ceil(data.my.total_participants * 0.1))" type="success">前 10%</BaseBadge>
            </div>

            <div class="relative mt-3">
              <p class="text-xs text-indigo-100">你的{{ metricLabel() }}</p>
              <p class="mt-1 text-2xl font-bold tabular-nums sm:text-3xl">
                {{ data.my ? valueText(data) : '暂无数据' }}
              </p>
            </div>

            <div class="relative mt-auto pt-6">
              <template v-if="data.my">
                <div class="flex items-center justify-between text-xs">
                  <span class="text-indigo-100">击败比例</span>
                  <span class="font-semibold tabular-nums">{{ beatRatio() }}%</span>
                </div>
                <div class="mt-2 h-2 overflow-hidden rounded-full bg-white/15 ring-1 ring-inset ring-white/10">
                  <div
                    class="h-full rounded-full bg-white shadow-sm transition-all duration-700"
                    :style="{ width: beatRatio() + '%' }"
                  />
                </div>
                <p class="mt-2 text-[11px] text-indigo-100">超过前 {{ data.my.percentile }}% 的用户</p>
              </template>
              <div v-else class="rounded-xl bg-white/10 px-3 py-2.5 text-xs text-indigo-50 ring-1 ring-inset ring-white/10">
                暂未参与该维度排名，快去积累数据吧
              </div>
            </div>
          </div>
        </div>

        <!-- 分布直方图（7/12） -->
        <div class="lg:col-span-7">
          <div class="flex h-full flex-col rounded-2xl border border-gray-100 bg-white p-5 shadow-sm sm:p-6">
            <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
              <div>
                <h3 class="text-base font-semibold tracking-tight text-gray-900">各阶段人员分布</h3>
                <p class="mt-0.5 text-xs text-gray-400">参与者{{ metricLabel() }}分布，柱越高人数越多；高亮柱为你所在阶段</p>
              </div>
              <BaseBadge v-if="data.my" type="info" dot>{{ data.my.total_participants }} 人参与</BaseBadge>
            </div>

            <div v-if="data.distribution && data.distribution.length" class="flex min-h-[260px] flex-1 gap-2 sm:gap-3">
              <!-- Y 轴刻度 -->
              <div class="flex w-7 shrink-0 flex-col justify-between pb-9 text-right text-[10px] font-medium text-gray-400 sm:w-9">
                <span class="leading-none">100%</span>
                <span class="leading-none">50%</span>
                <span class="leading-none">0</span>
              </div>

              <!-- 柱状图区域 -->
              <div class="relative flex-1">
                <!-- 横向网格线 -->
                <div class="pointer-events-none absolute inset-0 flex flex-col justify-between pb-9">
                  <div class="border-t border-dashed border-gray-100" />
                  <div class="border-t border-dashed border-gray-100" />
                  <div class="border-t border-gray-100" />
                </div>

                <!-- 柱子 -->
                <div class="relative flex h-full items-end gap-1.5 pb-9 sm:gap-2">
                  <div
                    v-for="(item, idx) in data.distribution"
                    :key="item.label"
                    class="group flex flex-1 flex-col items-center justify-end"
                  >
                    <!-- 人数标签（柱顶） -->
                    <div
                      class="mb-1.5 text-xs font-bold tabular-nums transition-colors duration-300"
                      :class="barTextClass(item)"
                    >
                      {{ item.count > 0 ? item.count : '·' }}
                    </div>

                    <!-- 柱体 -->
                    <div
                      class="bar-anim relative w-full overflow-hidden rounded-t-md transition-all duration-500"
                      :class="[barFillClass(item), { 'bar-mine-anim': item.is_mine }]"
                      :style="{
                        height: barHeight(item) + '%',
                        animationDelay: (idx * 70) + 'ms',
                      }"
                    >
                      <!-- 柱内渐变光晕（仅高亮柱） -->
                      <div v-if="item.is_mine" class="pointer-events-none absolute inset-0 bg-gradient-to-b from-white/30 via-transparent to-transparent" />
                    </div>

                    <!-- 「你」徽章（柱上方） -->
                    <div v-if="item.is_mine" class="pointer-events-none absolute -top-1 left-1/2 -translate-x-1/2 -translate-y-full">
                      <span class="inline-flex items-center gap-1 rounded-full bg-indigo-600 px-2 py-0.5 text-[10px] font-bold text-white shadow-md ring-2 ring-white">
                        <svg class="h-2.5 w-2.5" fill="currentColor" viewBox="0 0 20 20">
                          <path d="M10 2.5l2.5 5 5.5.8-4 3.9.9 5.5L10 15.1 5.1 17.7l.9-5.5-4-3.9 5.5-.8L10 2.5z" />
                        </svg>
                        你
                      </span>
                      <!-- 向下小三角 -->
                      <span class="absolute left-1/2 top-full -translate-x-1/2 -translate-y-0.5 border-x-4 border-t-4 border-x-transparent border-t-indigo-600" />
                    </div>

                    <!-- 占比（柱底） -->
                    <div class="mt-1.5 text-[10px] font-medium tabular-nums text-gray-400">
                      {{ item.ratio > 0 ? (item.ratio * 100).toFixed(0) + '%' : '0%' }}
                    </div>
                  </div>
                </div>

                <!-- X 轴标签 -->
                <div class="absolute bottom-0 left-0 right-0 flex justify-between gap-1.5 sm:gap-2">
                  <div
                    v-for="item in data.distribution"
                    :key="item.label"
                    class="flex-1 text-center text-[11px] font-medium transition-colors"
                    :class="item.is_mine ? 'text-indigo-600' : 'text-gray-500'"
                  >
                    {{ item.label }}
                  </div>
                </div>
              </div>
            </div>

            <div v-else class="flex flex-1 items-center justify-center py-8 text-sm text-gray-400">
              暂无分布数据
            </div>
          </div>
        </div>
      </div>

      <!-- 关键指标行 -->
      <div class="mb-6 grid gap-4 sm:grid-cols-3">
        <div class="group rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-shadow hover:shadow-md sm:p-5">
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 transition-colors group-hover:bg-indigo-100">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 002.625.372 9.337 9.337 0 004.121-.952 4.125 4.125 0 00-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 018.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0111.964-3.07M12 6.375a3.375 3.375 0 11-6.75 0 3.375 3.375 0 016.75 0zm8.25 2.25a2.625 2.625 0 11-5.25 0 2.625 2.625 0 015.25 0z" />
              </svg>
            </div>
            <div>
              <p class="text-xs text-gray-400">同等级用户</p>
              <p class="mt-0.5 text-lg font-bold text-gray-900 tabular-nums">
                {{ data.scope.total_users }}<span class="ml-0.5 text-xs font-normal text-gray-400">人</span>
              </p>
            </div>
          </div>
        </div>

        <div class="group rounded-xl border border-gray-100 bg-white p-4 shadow-sm transition-shadow hover:shadow-md sm:p-5">
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-amber-50 text-amber-600 transition-colors group-hover:bg-amber-100">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
              </svg>
            </div>
            <div>
              <p class="text-xs text-gray-400">全平台均值</p>
              <p class="mt-0.5 text-lg font-bold text-gray-900 tabular-nums">
                {{ formatValue(data.reference.avg) }}
              </p>
            </div>
          </div>
        </div>

        <div class="group rounded-xl border border-indigo-100 bg-gradient-to-br from-indigo-50/80 to-violet-50/60 p-4 shadow-sm transition-shadow hover:shadow-md sm:p-5">
          <div class="flex items-center gap-3">
            <div class="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-600 text-white shadow-md shadow-indigo-600/30">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12c0 1.268-.63 2.39-1.593 3.068a3.745 3.745 0 01-1.043 3.296 3.745 3.745 0 01-3.296 1.043A3.745 3.745 0 0112 21c-1.268 0-2.39-.63-3.068-1.593a3.746 3.746 0 01-3.296-1.043 3.745 3.745 0 01-1.043-3.296A3.745 3.745 0 013 12c0-1.268.63-2.39 1.593-3.068a3.745 3.745 0 011.043-3.296 3.746 3.746 0 013.296-1.043A3.746 3.746 0 0112 3c1.268 0 2.39.63 3.068 1.593a3.746 3.746 0 013.296 1.043 3.746 3.746 0 011.043 3.296A3.745 3.745 0 0121 12z" />
              </svg>
            </div>
            <div>
              <p class="text-xs text-indigo-600/80">进入前 10% 需达到</p>
              <p class="mt-0.5 text-lg font-bold text-indigo-600 tabular-nums">
                {{ formatValue(data.reference.top10_threshold) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- 口径说明 -->
      <div class="rounded-xl border border-gray-100 bg-white p-5 shadow-sm">
        <h3 class="mb-2 text-sm font-semibold text-gray-800">口径说明</h3>
        <ul class="space-y-1.5 text-xs leading-relaxed text-gray-500">
          <li v-if="category === 'checkin' && metric === 'streak'">· 按最高连续打卡天数排名，并列时参考累计打卡天数。</li>
          <li v-if="category === 'checkin' && metric === 'accuracy'">· 按全部打卡的每日正确率平均值排名。</li>
          <li v-if="category === 'exam' && metric === 'accuracy'">· 按已交卷模拟考试的正确率平均值排名。</li>
          <li v-if="category === 'exam' && metric === 'score'">· 预估分数为历史模考得分的时间加权平均（越近权重越大），至少 1 次交卷才参与。</li>
          <li v-if="category === 'practice' && metric === 'accuracy'">· 按练习正确率排名（少量答题会向平均水平收缩，防止刷榜），展示的是原始正确率。</li>
          <li v-if="category === 'practice' && metric === 'score'">· 预估分数 = 平滑正确率 × 75，至少练习 10 题才参与。</li>
          <li>· 默认与同等级用户对比，可选择科目进一步筛选；打卡榜单按等级对比。</li>
          <li>· 出于隐私考虑，仅展示你的名次与匿名统计线，不显示其他用户信息。</li>
        </ul>
      </div>
    </template>
  </div>
</template>

<style scoped>
@keyframes barRise {
  from {
    transform: scaleY(0);
    opacity: 0;
  }
  to {
    transform: scaleY(1);
    opacity: 1;
  }
}

@keyframes minePulse {
  0%, 100% {
    box-shadow: 0 8px 20px -4px rgba(99, 102, 241, 0.35), 0 0 0 2px rgba(255, 255, 255, 0.7);
  }
  50% {
    box-shadow: 0 10px 28px -2px rgba(99, 102, 241, 0.55), 0 0 0 2px rgba(255, 255, 255, 0.7);
  }
}

.bar-anim {
  transform-origin: bottom;
  animation: barRise 0.7s cubic-bezier(0.16, 1, 0.3, 1) backwards;
}

.bar-mine-anim {
  animation: barRise 0.7s cubic-bezier(0.16, 1, 0.3, 1) backwards,
             minePulse 2.6s ease-in-out infinite 0.8s;
}
</style>

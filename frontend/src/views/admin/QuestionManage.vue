<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import {
  batchSetAdminQuestionStatus,
  createAdminQuestion,
  getAdminQuestionStats,
  listAdminQuestions,
  updateAdminQuestion,
} from '@/api/admin'
import { getChapters, getSubjectsByLevel, getSubSubjects, getExamLevels } from '@/api/subject'
import type {
  AdminQuestionResp,
  AdminQuestionStatsResp,
  AdminQuestionUpsertReq,
} from '@/types/admin'
import type { ExamLevel, Subject, SubSubject, Chapter } from '@/types/subject'
import type { Difficulty, QuestionType } from '@/types/question'
import { showToast } from '@/utils/toast'
import { typeLabel, difficultyLabel } from '@/utils/question'
import BasePageHeader from '@/components/common/BasePageHeader.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseBadge from '@/components/common/BaseBadge.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextarea from '@/components/common/BaseTextarea.vue'
import BaseModal from '@/components/common/BaseModal.vue'
import BaseEmpty from '@/components/common/BaseEmpty.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

// ===== 常量 =====
const QUESTION_TYPES: QuestionType[] = ['single', 'multi', 'judge', 'fill', 'multi_blank', 'short', 'comprehensive', 'essay', 'case_study']
const DIFFICULTIES: Difficulty[] = ['easy', 'medium', 'hard']

const typeMeta: Record<QuestionType, { label: string; badge: 'default' | 'info' | 'success' | 'warning' | 'danger' }> = {
  single: { label: '单选', badge: 'info' },
  multi: { label: '多选', badge: 'warning' },
  judge: { label: '判断', badge: 'success' },
  fill: { label: '填空', badge: 'default' },
  short: { label: '简答', badge: 'default' },
  comprehensive: { label: '综合', badge: 'default' },
  essay: { label: '论文', badge: 'danger' },
  case_study: { label: '案例', badge: 'danger' },
  multi_blank: { label: '多空', badge: 'warning' },
}

// ===== 筛选器 =====
const levels = ref<ExamLevel[]>([])
const subjects = ref<Subject[]>([])
const subSubjects = ref<SubSubject[]>([])
const chapters = ref<Chapter[]>([])

const filters = reactive({
  level_id: '' as string | number,
  subject_id: '' as string | number,
  sub_subject_id: '' as string | number,
  chapter_id: '' as string | number,
  type: '' as string,
  difficulty: '' as string,
  status: '' as string,
  keyword: '',
  year: '' as string,
})

// ===== 筛选芯片（已选条件展示 & 一键移除） =====
const filterChips = computed(() => {
  const chips: { key: string; label: string }[] = []
  if (filters.level_id) chips.push({ key: 'level_id', label: `等级：${levels.value.find(l => l.id === Number(filters.level_id))?.name ?? ''}` })
  if (filters.subject_id) chips.push({ key: 'subject_id', label: `科目：${subjects.value.find(s => s.id === Number(filters.subject_id))?.name ?? ''}` })
  if (filters.sub_subject_id) chips.push({ key: 'sub_subject_id', label: `子科目：${subSubjects.value.find(s => s.id === Number(filters.sub_subject_id))?.name ?? ''}` })
  if (filters.chapter_id) chips.push({ key: 'chapter_id', label: `章节：${chapters.value.find(c => c.id === Number(filters.chapter_id))?.name ?? ''}` })
  if (filters.type) chips.push({ key: 'type', label: `题型：${typeLabel(filters.type)}` })
  if (filters.difficulty) chips.push({ key: 'difficulty', label: `难度：${difficultyLabel(filters.difficulty)}` })
  if (filters.status !== '') chips.push({ key: 'status', label: `状态：${filters.status === '1' ? '启用' : '已停用'}` })
  if (filters.year) chips.push({ key: 'year', label: `年份：${filters.year}` })
  if (filters.keyword.trim()) chips.push({ key: 'keyword', label: `关键词：${filters.keyword.trim()}` })
  return chips
})

function removeChip(key: string) {
  if (key === 'level_id') filters.level_id = ''
  else if (key === 'subject_id') filters.subject_id = ''
  else if (key === 'sub_subject_id') filters.sub_subject_id = ''
  else if (key === 'chapter_id') filters.chapter_id = ''
  else if (key === 'type') filters.type = ''
  else if (key === 'difficulty') filters.difficulty = ''
  else if (key === 'status') filters.status = ''
  else if (key === 'year') filters.year = ''
  else if (key === 'keyword') filters.keyword = ''
}

const hasActiveFilters = computed(() => filterChips.value.length > 0)

// ===== 列表 =====
const list = ref<AdminQuestionResp[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(true)
const error = ref('')

const stats = ref<AdminQuestionStatsResp | null>(null)

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))

// 选择
const selected = ref<Set<number>>(new Set())
const allChecked = computed(() => list.value.length > 0 && list.value.every(q => selected.value.has(q.id)))

// ===== 编辑弹窗 =====
const editorOpen = ref(false)
const editingId = ref<number | null>(null)
const saving = ref(false)

// 选项编辑：数组中每一项 { letter, content, checked }；checked 仅作答答案标记用
interface OptionRow {
  letter: string
  content: string
  checked: boolean
}

const form = reactive({
  level_id: '' as string | number,
  subject_id: '' as string | number,
  sub_subject_id: '' as string | number,
  chapter_id: '' as string | number,
  type: 'single' as QuestionType,
  difficulty: 'easy' as Difficulty,
  content: '',
  case_material: '',
  options: [] as OptionRow[], // single/multi/judge
  blank_options: [] as { options: OptionRow[] }[], // multi_blank 每空选项
  blank_answers: [] as string[], // multi_blank 每空答案
  answer: '', // 选择题答案字母 / 填空题/主观题文本
  analysis: '',
  year: '' as string,
  source: '',
  status: 1,
})

const formErrors = reactive<Record<string, string>>({})

const confirmDialog = ref<InstanceType<typeof ConfirmDialog>>()
const pendingAction = ref<(() => void) | null>(null)

// 判断题答案：A=正确 / B=错误（与后端/前端判分约定一致）
const judgeAnswer = ref<'A' | 'B'>('A')

// ===== 加载科目树级联 =====
async function loadLevels() {
  levels.value = await getExamLevels()
}

async function loadSubjects() {
  const lv = Number(filters.level_id || 0)
  if (!lv) return
  subjects.value = await getSubjectsByLevel(lv)
}

async function loadSubSubjects() {
  const sid = Number(filters.subject_id || 0)
  if (!sid) return
  subSubjects.value = await getSubSubjects(sid)
}

async function loadChapters() {
  const ssid = Number(filters.sub_subject_id || 0)
  if (!ssid) return
  chapters.value = await getChapters(ssid)
}

watch(() => filters.level_id, () => {
  filters.subject_id = ''
  filters.sub_subject_id = ''
  filters.chapter_id = ''
  subjects.value = []
  subSubjects.value = []
  chapters.value = []
  loadSubjects().catch(() => {})
  page.value = 1
  load()
})

watch(() => filters.subject_id, () => {
  filters.sub_subject_id = ''
  filters.chapter_id = ''
  subSubjects.value = []
  chapters.value = []
  if (filters.subject_id) loadSubSubjects().catch(() => {})
  page.value = 1
  load()
})

watch(() => filters.sub_subject_id, () => {
  filters.chapter_id = ''
  chapters.value = []
  if (filters.sub_subject_id) loadChapters().catch(() => {})
  page.value = 1
  load()
})

watch(() => filters.chapter_id, () => {
  page.value = 1
  load()
})

watch(() => [filters.type, filters.difficulty, filters.status, filters.year], () => {
  page.value = 1
  load()
})

// 关键词输入防抖
let keywordTimer: ReturnType<typeof setTimeout> | null = null
watch(() => filters.keyword, () => {
  if (keywordTimer) clearTimeout(keywordTimer)
  keywordTimer = setTimeout(() => {
    page.value = 1
    load()
  }, 500)
})

// ===== 加载列表 =====
async function load() {
  loading.value = true
  error.value = ''
  try {
    const params = {
      subject_id: filters.subject_id ? Number(filters.subject_id) : undefined,
      sub_subject_id: filters.sub_subject_id ? Number(filters.sub_subject_id) : undefined,
      chapter_id: filters.chapter_id ? Number(filters.chapter_id) : undefined,
      type: filters.type || undefined,
      difficulty: filters.difficulty || undefined,
      status: filters.status === '' ? undefined : filters.status,
      keyword: filters.keyword.trim() || undefined,
      year: filters.year ? Number(filters.year) : undefined,
      page: page.value,
      page_size: pageSize.value,
    }
    const data = await listAdminQuestions(params)
    list.value = data.list
    total.value = data.total
    page.value = data.page
    pageSize.value = data.page_size
    // 清除已失效的行选择
    const ids = new Set(data.list.map(q => q.id))
    selected.value = new Set([...selected.value].filter(id => ids.has(id)))
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    stats.value = await getAdminQuestionStats()
  } catch {
    stats.value = null
  }
}

function resetFilters() {
  Object.assign(filters, { level_id: '', subject_id: '', sub_subject_id: '', chapter_id: '', type: '', difficulty: '', status: '', keyword: '', year: '' })
  subjects.value = []
  subSubjects.value = []
  chapters.value = []
  page.value = 1
  load()
}

function goPage(next: number) {
  if (next < 1 || next > totalPages.value || next === page.value) return
  page.value = next
  load()
}

onMounted(async () => {
  await loadLevels()
  await load()
  loadStats()
})

// ===== 选择 =====
function toggleAll() {
  if (allChecked.value) {
    list.value.forEach(q => selected.value.delete(q.id))
  } else {
    list.value.forEach(q => selected.value.add(q.id))
  }
  // 触发响应式
  selected.value = new Set(selected.value)
}

function toggleOne(id: number) {
  const s = new Set(selected.value)
  if (s.has(id)) s.delete(id)
  else s.add(id)
  selected.value = s
}

function batchStatus(status: number) {
  const ids = [...selected.value]
  if (ids.length === 0) {
    showToast('请先勾选要操作的题目', 'info')
    return
  }
  const label = status === 1 ? '启用' : '停用'
  pendingAction.value = () => doBatchStatus(ids, status)
  confirmDialog.value?.open()
  pendingActionLabel.value = `确定${label}选中的 ${ids.length} 道题目？`
}

const pendingActionLabel = ref('')

async function doBatchStatus(ids: number[], status: number) {
  try {
    await batchSetAdminQuestionStatus({ ids, status })
    showToast(`已${status === 1 ? '启用' : '停用'} ${ids.length} 道题目`, 'success')
    selected.value = new Set()
    await load()
    loadStats()
  } catch {
    // 请求层统一 toast
  }
}

// ===== 编辑 =====
function resetForm() {
  Object.assign(form, {
    level_id: '', subject_id: '', sub_subject_id: '', chapter_id: '',
    type: 'single', difficulty: 'easy',
    content: '', case_material: '',
    options: [], blank_options: [], blank_answers: [],
    answer: '', analysis: '', year: '', source: '', status: 1,
  })
  Object.keys(formErrors).forEach(k => delete formErrors[k])
}

// 需要自定义选项的客观题（单选 / 多选）
function isChoiceType(t: QuestionType) {
  return t === 'single' || t === 'multi'
}

// 判断题：答案为二元（正确 / 错误），选项固定为 A=正确 / B=错误
function isJudgeType(t: QuestionType) {
  return t === 'judge'
}

// 判断题选项（固定两行，供存储与展示）
function buildJudgeOptions() {
  return [
    { letter: 'A', content: '正确', checked: judgeAnswer.value === 'A' },
    { letter: 'B', content: '错误', checked: judgeAnswer.value === 'B' },
  ]
}

// 从题目响应解析选项字符串 → 选项行
function optionsFromString(options: string): OptionRow[] {
  if (!options) return []
  try {
    const arr = JSON.parse(options)
    if (Array.isArray(arr)) {
      return arr.map((o, i) => {
        const letter = String.fromCharCode(65 + i)
        if (typeof o === 'string') return { letter, content: o, checked: false }
        const content = typeof o.content === 'string' ? o.content : JSON.stringify(o.content ?? '')
        return { letter: letter, content, checked: false }
      })
    }
    if (typeof arr === 'object' && arr !== null) {
      // {"A":"...","B":"..."} 形态
      return Object.entries(arr).map(([letter, c]) => ({ letter, content: String(c), checked: false }))
    }
  } catch {
    /* 忽略解析失败，返回空 */
  }
  return []
}

// 多选题答案字符串拆成字母集合
function answerSet(answer: string): Set<string> {
  return new Set((answer || '').toUpperCase().split(',').map(s => s.trim()).filter(Boolean))
}

function fillAnswerState(row: AdminQuestionResp) {
  if (form.type === 'judge') {
    const ans = (row.answer || '').toUpperCase().trim()
    judgeAnswer.value = ans === 'B' ? 'B' : 'A'
    form.options = buildJudgeOptions()
    return
  }
  if (form.type === 'single') {
    const ans = (row.answer || '').toUpperCase().trim()
    form.options.forEach((_, i) => { form.options[i].checked = form.options[i].letter.toUpperCase() === ans })
  } else if (form.type === 'multi') {
    const set = answerSet(row.answer)
    form.options.forEach((_, i) => { form.options[i].checked = set.has(form.options[i].letter.toUpperCase()) })
  }
}

// 解析多空题的答案与选项
function fillBlankState(row: AdminQuestionResp) {
  const answers = parseBlankAnswerArray(row.answer)
  form.blank_answers = answers.map(a => a)
  form.blank_options = []
  try {
    const groups = JSON.parse(row.blank_options || '[]')
    if (Array.isArray(groups)) {
      form.blank_options = groups.map((g: { options?: unknown }) => {
        const opts = Array.isArray(g.options) ? g.options : []
        return {
          options: opts.map((o: any, i: number) => {
            const letter = typeof o.id === 'string' ? o.id : String.fromCharCode(65 + i)
            return { letter, content: typeof o.content === 'string' ? o.content : String(o.content ?? ''), checked: false }
          }),
        }
      })
      // 用答案反填勾选
      form.blank_options.forEach((blank, gi) => {
        const target = answers[gi]?.toUpperCase()
        blank.options.forEach(o => { o.checked = o.letter.toUpperCase() === target })
      })
    }
  } catch {
    form.blank_options = []
  }
  // 若无任何空选项组，按答案数量初始化空组
  if (form.blank_options.length === 0 && form.blank_answers.length > 0) {
    form.blank_options = form.blank_answers.map(() => ({ options: [] }))
  }
}

function parseBlankAnswerArray(answer: string): string[] {
  if (!answer) return []
  const s = answer.trim()
  if (s.startsWith('[')) {
    try {
      const arr = JSON.parse(s)
      if (Array.isArray(arr)) return arr.map(String)
    } catch { /* ignore */ }
  }
  return s.split(',').map(x => x.trim()).filter(Boolean)
}

async function openCreate() {
  editingId.value = null
  resetForm()
  form.status = 1
  editorOpen.value = true
  // 预填当前筛选的科目，方便连续录入
  if (filters.level_id) form.level_id = filters.level_id
  if (filters.subject_id) {
    form.subject_id = filters.subject_id
    await loadFormSubSubjects()
  }
  if (filters.sub_subject_id) {
    form.sub_subject_id = filters.sub_subject_id
    await loadFormChapters()
  }
  if (filters.chapter_id) form.chapter_id = filters.chapter_id
  if (filters.difficulty) form.difficulty = filters.difficulty as Difficulty
  // 默认加几个空选项位
  if (form.options.length === 0) {
    form.options = ['A', 'B', 'C', 'D'].map(l => ({ letter: l, content: '', checked: false }))
  }
}

async function openEdit(row: AdminQuestionResp) {
  editingId.value = row.id
  resetForm()
  editorOpen.value = true
  // 预填
  form.type = row.type
  form.difficulty = row.difficulty
  form.content = row.content
  form.case_material = row.case_material || ''
  form.analysis = row.analysis || ''
  form.answer = row.answer
  form.year = row.year ? String(row.year) : ''
  form.source = row.source || ''
  form.status = row.status
  // 载入科目树
  if (row.subject_id) {
    // 需要定位其所属等级：通过加载全部科目难以直接推断，此处允许仅保留 subject_id
    form.subject_id = row.subject_id
  }
  form.sub_subject_id = row.sub_subject_id
  form.chapter_id = row.chapter_id
  if (row.subject_id) {
    try {
      // 反查等级
      const all = await getAllSubjectsFlat()
      const s = all.find(x => x.id === row.subject_id)
      if (s) {
        form.level_id = s.level_id
        await loadFormSubjects()
      }
    } catch { /* ignore */ }
    await loadFormSubSubjects().catch(() => {})
  }
  if (row.sub_subject_id) {
    await loadFormChapters().catch(() => {})
  }
  // 选项
  if (isJudgeType(row.type)) {
    form.options = buildJudgeOptions()
    fillAnswerState(row)
  } else if (isChoiceType(row.type)) {
    form.options = optionsFromString(row.options)
    if (form.options.length === 0) {
      form.options = ['A', 'B', 'C', 'D'].map(l => ({ letter: l, content: '', checked: false }))
    }
    fillAnswerState(row)
  } else if (row.type === 'multi_blank') {
    fillBlankState(row)
  }
}

// 编辑弹窗里的级联数据（独立于筛选器，避免相互干扰）
const formSubjects = ref<Subject[]>([])
const formSubSubjects = ref<SubSubject[]>([])
const formChapters = ref<Chapter[]>([])

async function getAllSubjectsFlat(): Promise<Subject[]> {
  // 从三个等级各拉一遍拼起来（无 /subjects/all 用于学员组则用等级遍历）
  const out: Subject[] = []
  for (const lv of levels.value) {
    try {
      const arr = await getSubjectsByLevel(lv.id)
      out.push(...arr)
    } catch { /* ignore */ }
  }
  return out
}

async function loadFormSubjects() {
  const lv = Number(form.level_id || 0)
  if (!lv) return
  formSubjects.value = await getSubjectsByLevel(lv)
}

async function loadFormSubSubjects() {
  const sid = Number(form.subject_id || 0)
  if (!sid) return
  formSubSubjects.value = await getSubSubjects(sid)
}

async function loadFormChapters() {
  const ssid = Number(form.sub_subject_id || 0)
  if (!ssid) return
  formChapters.value = await getChapters(ssid)
}

// 编辑弹窗级联 watch
watch(() => form.level_id, async () => {
  form.subject_id = ''
  form.sub_subject_id = ''
  form.chapter_id = ''
  formSubjects.value = []
  formSubSubjects.value = []
  formChapters.value = []
  if (form.level_id) await loadFormSubjects().catch(() => {})
})
watch(() => form.subject_id, async () => {
  form.sub_subject_id = ''
  form.chapter_id = ''
  formSubSubjects.value = []
  formChapters.value = []
  if (form.subject_id) await loadFormSubSubjects().catch(() => {})
})
watch(() => form.sub_subject_id, async () => {
  form.chapter_id = ''
  formChapters.value = []
  if (form.sub_subject_id) await loadFormChapters().catch(() => {})
})
watch(() => form.type, (t) => {
  // 切换题型时重置选项结构
  Object.keys(formErrors).forEach(k => delete formErrors[k])
  judgeAnswer.value = 'A'
  if (isJudgeType(t)) {
    form.options = buildJudgeOptions()
    form.answer = judgeAnswer.value
  } else if (isChoiceType(t)) {
    // 保留已有选项，若无则补默认 A/B/C/D
    if (form.options.length === 0) {
      form.options = ['A', 'B', 'C', 'D'].map(l => ({ letter: l, content: '', checked: false }))
    }
    form.options.forEach(o => { o.checked = false })
    form.answer = ''
  } else if (t === 'multi_blank') {
    form.blank_options = []
    form.blank_answers = []
    form.answer = ''
  } else {
    form.answer = ''
  }
})

// ===== 表单校验 & 提交 =====
function validate(): boolean {
  Object.keys(formErrors).forEach(k => delete formErrors[k])
  if (!form.subject_id) formErrors.subject_id = '请选择科目'
  if (!form.type) formErrors.type = '请选择题型'
  if (!form.content.trim()) formErrors.content = '请输入题干'
  if (form.type === 'single') {
    const validOpts = form.options.filter(o => o.content.trim() !== '')
    if (validOpts.length < 2) {
      formErrors.options = '单选题至少需要两个选项'
    }
    const checked = validOpts.filter(o => o.checked)
    if (checked.length !== 1) formErrors.answer = '请勾选唯一正确答案'
  } else if (form.type === 'multi') {
    const validOpts = form.options.filter(o => o.content.trim() !== '')
    if (validOpts.length < 2) {
      formErrors.options = '多选题至少需要两个选项'
    }
    const checked = validOpts.filter(o => o.checked)
    if (checked.length < 2) formErrors.answer = '请勾选至少两个正确答案'
  } else if (form.type === 'judge') {
    // 判断题答案由 正确/错误 二元选择驱动，无需额外校验
  } else if (form.type === 'multi_blank') {
    if (form.blank_answers.length === 0) formErrors.answer = '请设置至少一个空'
    for (let i = 0; i < form.blank_answers.length; i++) {
      if (!form.blank_answers[i]) formErrors.answer = `第 ${i + 1} 空答案未设置`
    }
    for (let i = 0; i < form.blank_options.length; i++) {
      const opts = (form.blank_options[i]?.options || []).filter(o => o.content.trim() !== '')
      if (opts.length < 2) formErrors.answer = `第 ${i + 1} 空至少需要两个选项`
    }
  } else if (!form.answer.trim()) {
    // fill/short/comprehensive/essay/case 允许留空参考答案，但给出提示而非强校验
  }
  return Object.keys(formErrors).length === 0
}

// 组装选择题 options JSON
function buildChoiceOptionsJSON(): string {
  return JSON.stringify(
    form.options.filter(o => o.content.trim() !== '').map(o => ({ id: o.letter, content: o.content.trim() }))
  )
}

// 组装多空题 blank_options JSON 与 answer JSON
function buildBlankPayload(): { blank_options: string; answer: string } {
  const answers = form.blank_answers.map(a => a.trim().toUpperCase())
  const groups = form.blank_options.map((g, gi) => ({
    blank_index: gi + 1,
    options: g.options.filter(o => o.content.trim() !== '').map(o => ({ id: o.letter, content: o.content.trim() })),
  }))
  return {
    blank_options: JSON.stringify(groups),
    answer: JSON.stringify(answers),
  }
}

function buildPayload(): AdminQuestionUpsertReq {
  let options = ''
  let blank_options = ''
  let answer = ''

  if (isJudgeType(form.type)) {
    // 判断题固定选项：A=正确 / B=错误
    options = JSON.stringify([{ id: 'A', content: '正确' }, { id: 'B', content: '错误' }])
    answer = judgeAnswer.value
  } else if (isChoiceType(form.type)) {
    options = buildChoiceOptionsJSON()
    const checked = form.options.filter(o => o.content.trim() !== '' && o.checked)
    const letters = checked.map(o => o.letter.toUpperCase()).sort()
    answer = letters.join(',')
  } else if (form.type === 'multi_blank') {
    const blank = buildBlankPayload()
    blank_options = blank.blank_options
    answer = blank.answer
  } else {
    answer = form.answer.trim()
  }

  return {
    subject_id: Number(form.subject_id),
    sub_subject_id: form.sub_subject_id ? Number(form.sub_subject_id) : 0,
    chapter_id: form.chapter_id ? Number(form.chapter_id) : 0,
    type: form.type,
    difficulty: form.difficulty,
    content: form.content.trim(),
    case_material: form.case_material.trim(),
    options,
    blank_options,
    answer,
    analysis: form.analysis.trim(),
    year: form.year ? Number(form.year) : 0,
    source: form.source.trim(),
    status: form.status,
  }
}

async function submitForm() {
  if (!validate()) {
    showToast('请修正表单中的错误', 'error')
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await updateAdminQuestion(editingId.value, payload)
      showToast('题目已更新', 'success')
    } else {
      await createAdminQuestion(payload)
      showToast('题目已创建', 'success')
    }
    editorOpen.value = false
    page.value = 1
    await load()
    loadStats()
  } catch {
    // 请求层已统一 toast
  } finally {
    saving.value = false
  }
}

// 新增一个空的填空位
function addBlank() {
  form.blank_options.push({ options: ['A', 'B'].map(l => ({ letter: l, content: '', checked: false })) })
  form.blank_answers.push('')
}

// 单题启用 / 停用
function confirmToggleStatus(row: AdminQuestionResp) {
  const target = row.status === 1 ? 0 : 1
  pendingAction.value = () => doToggleStatus(row, target)
  pendingActionLabel.value =
    target === 1
      ? `确定启用题目「${shortContent(row.content)}」？启用后学员可正常抽到此题。`
      : `确定停用题目「${shortContent(row.content)}」？停用后学员将不再抽到此题（历史记录保留）。`
  confirmDialog.value?.open()
}

async function doToggleStatus(row: AdminQuestionResp, status: number) {
  try {
    await batchSetAdminQuestionStatus({ ids: [row.id], status })
    showToast(status === 1 ? '题目已启用' : '题目已停用', 'success')
    selected.value.delete(row.id)
    selected.value = new Set(selected.value)
    await load()
    loadStats()
  } catch {
    // 请求层统一 toast
  }
}

function runPending() {
  const fn = pendingAction.value
  pendingAction.value = null
  if (fn) fn()
}

function shortContent(c: string) {
  const plain = c.replace(/<[^>]+>/g, '').replace(/\s+/g, ' ')
  return plain.length > 28 ? plain.slice(0, 28) + '…' : plain
}

// ===== 展示辅助 =====
function statusBadge(status: number) {
  return status === 1 ? { label: '启用', type: 'success' as const } : { label: '已停用', type: 'danger' as const }
}
function difficultyBadge(d: string) {
  return d === 'easy' ? { label: '简单', type: 'success' as const } : d === 'hard' ? { label: '困难', type: 'danger' as const } : { label: '中等', type: 'warning' as const }
}
</script>

<template>
  <div>
    <BasePageHeader title="题库管理" subtitle="检索、新增、编辑与维护全部考试题目">
      <template #actions>
        <BaseButton type="secondary" @click="load">刷新</BaseButton>
        <BaseButton @click="openCreate">新增题目</BaseButton>
      </template>
    </BasePageHeader>

    <!-- 概览统计 -->
    <div v-if="stats" class="mb-5 grid grid-cols-2 gap-3 lg:grid-cols-4">
      <div class="flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6.75h12M8.25 12h12m-12 5.25h12M3.75 6.75h.007v.008H3.75V6.75zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zM3.75 12h.007v.008H3.75V12zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm-.375 5.25h.007v.008H3.75v-.008zm.375 0a.375.375 0 11-.75 0 .375.375 0 01.75 0z" /></svg>
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-400">题目总量</p>
          <p class="mt-0.5 text-2xl font-bold tracking-tight text-gray-900">{{ stats.total }}</p>
        </div>
      </div>
      <div class="flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-emerald-50 text-emerald-600">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-400">启用</p>
          <p class="mt-0.5 text-2xl font-bold tracking-tight text-emerald-600">{{ stats.enabled }}</p>
        </div>
      </div>
      <div class="flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-red-50 text-red-500">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" /></svg>
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-400">已停用</p>
          <p class="mt-0.5 text-2xl font-bold tracking-tight text-red-500">{{ stats.disabled }}</p>
        </div>
      </div>
      <div class="flex items-center gap-4 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
        <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-violet-50 text-violet-600">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z" /></svg>
        </div>
        <div class="min-w-0">
          <p class="text-xs font-medium text-gray-400">题型数</p>
          <p class="mt-0.5 text-2xl font-bold tracking-tight text-violet-600">{{ stats.by_type.length }}</p>
        </div>
      </div>
    </div>

    <!-- 题型分布 -->
    <div v-if="stats && stats.by_type.length" class="mb-5 rounded-xl border border-gray-100 bg-white p-4 shadow-sm">
      <div class="mb-3 flex items-center gap-2">
        <span class="flex h-6 w-6 items-center justify-center rounded-md bg-violet-50 text-violet-600">
          <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M7.5 14.25v2.25m3-4.5v4.5m3-6.75v6.75m3-9v9M6 20.25h12A2.25 2.25 0 0020.25 18V6A2.25 2.25 0 0018 3.75H6A2.25 2.25 0 003.75 6v12A2.25 2.25 0 006 20.25z" /></svg>
        </span>
        <span class="text-sm font-semibold text-gray-700">题型分布</span>
      </div>
      <div class="flex flex-wrap items-center gap-2">
        <BaseBadge v-for="t in stats.by_type" :key="t.type" :type="typeMeta[t.type as QuestionType]?.badge || 'default'">
          {{ typeMeta[t.type as QuestionType]?.label || t.type }} · {{ t.count }}
        </BaseBadge>
      </div>
    </div>

    <!-- 筛选区 -->
    <div class="mb-5 rounded-xl border border-gray-100 bg-white shadow-sm">
      <!-- 第一行：级联归属（等级 → 科目 → 子科目 → 章节） -->
      <div class="border-b border-gray-50 p-4 pb-3">
        <p class="mb-2 flex items-center gap-1.5 text-xs font-medium text-gray-400">
          <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" /></svg>
          归属
        </p>
        <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <BaseSelect v-model="filters.level_id">
            <option value="">全部等级</option>
            <option v-for="l in levels" :key="l.id" :value="l.id">{{ l.name }}</option>
          </BaseSelect>
          <BaseSelect v-model="filters.subject_id" :disabled="!filters.level_id">
            <option value="">{{ filters.level_id ? '全部科目' : '请先选等级' }}</option>
            <option v-for="s in subjects" :key="s.id" :value="s.id">{{ s.name }}</option>
          </BaseSelect>
          <BaseSelect v-model="filters.sub_subject_id" :disabled="!filters.subject_id">
            <option value="">{{ filters.subject_id ? '全部子科目' : '请先选科目' }}</option>
            <option v-for="s in subSubjects" :key="s.id" :value="s.id">{{ s.name }}</option>
          </BaseSelect>
          <BaseSelect v-model="filters.chapter_id" :disabled="!filters.sub_subject_id">
            <option value="">{{ filters.sub_subject_id ? '全部章节' : '请先选子科目' }}</option>
            <option v-for="c in chapters" :key="c.id" :value="c.id">{{ c.name }}</option>
          </BaseSelect>
        </div>
      </div>

      <!-- 第二行：题目属性 + 关键词搜索 -->
      <div class="flex flex-col gap-2 border-b border-gray-50 p-4 pb-3 sm:flex-row sm:flex-wrap sm:items-center">
        <BaseSelect v-model="filters.type" class="sm:w-32">
          <option value="">全部题型</option>
          <option v-for="t in QUESTION_TYPES" :key="t" :value="t">{{ typeLabel(t) }}</option>
        </BaseSelect>
        <BaseSelect v-model="filters.difficulty" class="sm:w-28">
          <option value="">全部难度</option>
          <option v-for="d in DIFFICULTIES" :key="d" :value="d">{{ difficultyLabel(d) }}</option>
        </BaseSelect>
        <BaseSelect v-model="filters.status" class="sm:w-28">
          <option value="">全部状态</option>
          <option value="1">启用</option>
          <option value="0">已停用</option>
        </BaseSelect>
        <BaseInput v-model.number="filters.year" type="number" placeholder="年份，如 2023" class="sm:w-36" />
        <div class="relative min-w-0 flex-1 sm:min-w-[12rem]">
          <svg class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" /></svg>
          <input
            v-model="filters.keyword"
            type="text"
            placeholder="题干 / 解析 / 题源关键词（自动搜索）"
            class="w-full rounded-lg border border-gray-300 bg-white py-2 pl-9 pr-3 text-sm text-gray-800 shadow-sm transition-colors placeholder:text-gray-400 hover:border-gray-400 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
          />
        </div>
      </div>

      <!-- 第三行：已选筛选条件芯片 + 批量操作 -->
      <div class="flex flex-wrap items-center justify-between gap-3 p-4 pt-3">
        <div class="flex min-w-0 flex-wrap items-center gap-1.5">
          <transition-group name="fade">
            <template v-if="hasActiveFilters" key="chips">
              <span
                v-for="chip in filterChips"
                :key="chip.key"
                class="inline-flex cursor-default items-center gap-1 rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20"
              >
                {{ chip.label }}
                <button
                  type="button"
                  class="cursor-pointer rounded-full p-0.5 text-indigo-400 transition-colors hover:bg-indigo-100 hover:text-indigo-600"
                  :aria-label="`移除${chip.label}`"
                  @click="removeChip(chip.key)"
                >
                  <svg class="h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
                </button>
              </span>
              <button
                key="clear-all"
                type="button"
                class="cursor-pointer rounded-full px-2 py-0.5 text-xs font-medium text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600"
                @click="resetFilters"
              >
                清空全部
              </button>
            </template>
            <span v-else key="no-filter" class="text-xs text-gray-400">全部题目，选择上方条件自动筛选</span>
          </transition-group>
        </div>
        <transition name="fade">
          <div v-if="selected.size > 0" class="flex shrink-0 flex-wrap items-center gap-2">
            <span class="inline-flex items-center gap-1.5 rounded-full bg-indigo-50 px-2.5 py-0.5 text-xs font-medium text-indigo-700 ring-1 ring-inset ring-indigo-600/20">
              已选 {{ selected.size }} 题
            </span>
            <BaseButton size="sm" type="success" @click="batchStatus(1)">批量启用</BaseButton>
            <BaseButton size="sm" type="danger" @click="batchStatus(0)">批量停用</BaseButton>
          </div>
        </transition>
      </div>
    </div>

    <!-- 错误 -->
    <div v-if="error" class="mb-5">
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
        <BaseEmpty title="加载失败" :description="error">
          <template #action>
            <BaseButton type="secondary" size="sm" class="mt-4" @click="load">重新加载</BaseButton>
          </template>
        </BaseEmpty>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else-if="!loading && list.length === 0">
      <div class="rounded-xl border border-gray-100 bg-white p-6 shadow-sm">
        <BaseEmpty title="暂无符合条件的题目" description="调整筛选条件，或点击右上角「新增题目」">
          <template #action>
            <BaseButton size="sm" class="mt-4" @click="openCreate">新增题目</BaseButton>
          </template>
        </BaseEmpty>
      </div>
    </div>

    <!-- 列表 -->
    <template v-else>
      <div class="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
        <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm">
            <thead>
              <tr class="border-b border-gray-100 bg-gray-50/70 text-xs font-semibold uppercase tracking-wider text-gray-500">
                <th scope="col" class="w-10 px-3 py-3">
                  <input type="checkbox" class="cursor-pointer accent-indigo-600" :checked="allChecked" @change="toggleAll" />
                </th>
                <th scope="col" class="whitespace-nowrap px-4 py-3">ID</th>
                <th scope="col" class="whitespace-nowrap px-4 py-3">题目</th>
                <th scope="col" class="whitespace-nowrap px-4 py-3">类型</th>
                <th scope="col" class="whitespace-nowrap px-4 py-3">难度</th>
                <th scope="col" class="hidden whitespace-nowrap px-4 py-3 lg:table-cell">所属科目</th>
                <th scope="col" class="hidden whitespace-nowrap px-4 py-3 md:table-cell">年份</th>
                <th scope="col" class="whitespace-nowrap px-4 py-3">状态</th>
                <th scope="col" class="whitespace-nowrap px-4 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-50">
              <!-- 骨架屏 -->
              <template v-if="loading">
                <tr v-for="i in 5" :key="i">
                  <td class="px-3 py-4"><div class="skeleton h-4 w-4 rounded" /></td>
                  <td class="px-4 py-4"><div class="skeleton h-4 w-8" /></td>
                  <td class="px-4 py-4"><div class="skeleton h-4 w-72 max-w-full" /></td>
                  <td class="px-4 py-4"><div class="skeleton h-5 w-14 rounded-full" /></td>
                  <td class="px-4 py-4"><div class="skeleton h-5 w-14 rounded-full" /></td>
                  <td class="hidden px-4 py-4 lg:table-cell"><div class="skeleton h-4 w-24" /></td>
                  <td class="hidden px-4 py-4 md:table-cell"><div class="skeleton h-4 w-10" /></td>
                  <td class="px-4 py-4"><div class="skeleton h-5 w-14 rounded-full" /></td>
                  <td class="px-4 py-4"><div class="ml-auto flex justify-end gap-2"><div class="skeleton h-8 w-20 rounded-lg" /></div></td>
                </tr>
              </template>
              <!-- 真实行 -->
              <template v-else>
                <tr v-for="row in list" :key="row.id" class="group transition-colors duration-200 hover:bg-indigo-50/30">
                  <td class="px-3 py-3">
                    <input type="checkbox" class="cursor-pointer accent-indigo-600" :checked="selected.has(row.id)" @change="toggleOne(row.id)" />
                  </td>
                  <td class="whitespace-nowrap px-4 py-3 font-mono text-xs text-gray-400">#{{ row.id }}</td>
                  <td class="max-w-[26rem] px-4 py-3">
                    <p class="line-clamp-2 cursor-pointer text-gray-800 transition-colors hover:text-indigo-600" :title="row.content" @click="openEdit(row)">
                      {{ shortContent(row.content) }}
                    </p>
                    <p v-if="row.source" class="mt-0.5 text-xs text-gray-400">题源：{{ row.source }}</p>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3">
                    <BaseBadge :type="typeMeta[row.type as QuestionType]?.badge || 'default'">{{ typeLabel(row.type) }}</BaseBadge>
                  </td>
                  <td class="whitespace-nowrap px-4 py-3">
                    <BaseBadge :type="difficultyBadge(row.difficulty).type">{{ difficultyBadge(row.difficulty).label }}</BaseBadge>
                  </td>
                  <td class="hidden max-w-[16rem] px-4 py-3 lg:table-cell">
                    <p class="truncate text-xs text-gray-600" :title="row.subject_name">{{ row.subject_name }}</p>
                    <p class="truncate text-xs text-gray-400">{{ [row.sub_subject_name, row.chapter_name].filter(Boolean).join(' / ') }}</p>
                  </td>
                  <td class="hidden whitespace-nowrap px-4 py-3 text-gray-500 md:table-cell">{{ row.year || '-' }}</td>
                  <td class="whitespace-nowrap px-4 py-3">
                    <BaseBadge :type="statusBadge(row.status).type" dot>{{ statusBadge(row.status).label }}</BaseBadge>
                  </td>
                  <td class="px-4 py-3">
                    <div class="flex items-center justify-end gap-2 whitespace-nowrap">
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
                      <BaseButton v-else size="sm" type="ghost" class="text-emerald-600 hover:bg-emerald-50 hover:text-emerald-700" @click="confirmToggleStatus(row)">
                        <svg class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 10.5V6.75a4.5 4.5 0 119 0v3.75M3.75 21.75h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H3.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
                        启用
                      </BaseButton>
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
        <p class="text-xs text-gray-500">共 {{ total }} 道题目</p>
        <div class="flex items-center gap-2">
          <BaseButton size="sm" type="secondary" :disabled="page <= 1" @click="goPage(page - 1)">上一页</BaseButton>
          <span class="text-xs font-medium text-gray-600">第 {{ page }} / {{ totalPages }} 页</span>
          <BaseButton size="sm" type="secondary" :disabled="page >= totalPages" @click="goPage(page + 1)">下一页</BaseButton>
        </div>
      </div>
    </template>

    <!-- 编辑 / 新增弹窗 -->
    <BaseModal v-model="editorOpen" :title="editingId ? `编辑题目 #${editingId}` : '新增题目'" width="max-w-3xl">
      <div class="max-h-[72vh] space-y-4 overflow-y-auto pr-1">
        <!-- 基础归属 -->
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <BaseSelect v-model="form.type" label="题型" :disabled="!!editingId">
            <option v-for="t in QUESTION_TYPES" :key="t" :value="t">{{ typeLabel(t) }}</option>
          </BaseSelect>
          <BaseSelect v-model="form.difficulty" label="难度">
            <option v-for="d in DIFFICULTIES" :key="d" :value="d">{{ difficultyLabel(d) }}</option>
          </BaseSelect>
          <BaseInput v-model="form.year" label="年份" type="number" placeholder="如 2023" />
        </div>

        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <BaseSelect v-model="form.level_id" label="等级" :error="formErrors.subject_id ? '请选择科目' : undefined">
            <option value="">请选择等级</option>
            <option v-for="l in levels" :key="l.id" :value="l.id">{{ l.name }}</option>
          </BaseSelect>
          <BaseSelect v-model="form.subject_id" label="科目" :disabled="!form.level_id">
            <option value="">请选择科目</option>
            <option v-for="s in formSubjects" :key="s.id" :value="s.id">{{ s.name }}</option>
          </BaseSelect>
          <BaseSelect v-model="form.sub_subject_id" label="子科目" :disabled="!form.subject_id">
            <option value="">不归属子科目</option>
            <option v-for="s in formSubSubjects" :key="s.id" :value="s.id">{{ s.name }}</option>
          </BaseSelect>
        </div>
        <BaseSelect v-model="form.chapter_id" label="章节" :disabled="!form.sub_subject_id">
          <option value="">不归属章节</option>
          <option v-for="c in formChapters" :key="c.id" :value="c.id">{{ c.name }}</option>
        </BaseSelect>

        <!-- 题源/状态 -->
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <BaseInput v-model="form.source" label="题源" placeholder="如 ai / seed / sa-2023" />
          <BaseSelect v-model="form.status" label="状态">
            <option :value="1">启用</option>
            <option :value="0">停用</option>
          </BaseSelect>
        </div>

        <!-- 题干 -->
        <div>
          <BaseTextarea v-model="form.content" label="题干" :rows="3" placeholder="请输入题目内容，支持 Markdown" />
          <p v-if="formErrors.content" class="mt-1 text-xs text-red-500">{{ formErrors.content }}</p>
        </div>

        <!-- 案例分析材料 -->
        <div v-if="form.type === 'case_study'">
          <BaseTextarea v-model="form.case_material" label="案例材料" :rows="3" placeholder="案例分析题的大段案例背景" />
        </div>

        <!-- 选择题选项编辑 -->
        <div v-if="isChoiceType(form.type)">
          <div class="mb-2 flex items-center justify-between">
            <label class="text-sm font-medium text-gray-700">
              选项<span v-if="form.type === 'multi'" class="ml-1 text-xs text-gray-400">（多选，勾选正确答案）</span>
            </label>
            <BaseButton size="sm" type="ghost" @click="form.options.push({ letter: String.fromCharCode(65 + form.options.length), content: '', checked: false })">+ 添加选项</BaseButton>
          </div>
          <div class="space-y-2">
            <div v-for="(opt, i) in form.options" :key="i" class="flex items-center gap-2">
              <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border-2 border-gray-200 text-xs font-semibold text-gray-500">{{ opt.letter }}</span>
              <input v-model="opt.content" class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-800 focus:border-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500/20" placeholder="选项内容" />
              <label class="flex shrink-0 cursor-pointer items-center gap-1 text-xs text-gray-600">
                <input type="checkbox" v-model="opt.checked" class="cursor-pointer accent-indigo-600" /> 答案
              </label>
              <BaseButton v-if="form.options.length > 2" size="sm" type="ghost" class="shrink-0" @click="form.options.splice(i, 1)">删</BaseButton>
            </div>
          </div>
          <p v-if="formErrors.options" class="mt-1 text-xs text-red-500">{{ formErrors.options }}</p>
          <p v-if="formErrors.answer" class="mt-1 text-xs text-red-500">{{ formErrors.answer }}</p>
        </div>

        <!-- 判断题：正确答案 正确 / 错误 -->
        <div v-else-if="isJudgeType(form.type)">
          <div class="mb-2">
            <label class="mb-1.5 block text-sm font-medium text-gray-700">正确答案</label>
            <div class="grid grid-cols-2 gap-3">
              <button
                type="button"
                :class="[
                  'flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 px-4 py-3 text-sm font-medium transition-all duration-200 active:scale-[0.98]',
                  judgeAnswer === 'A'
                    ? 'border-emerald-500 bg-emerald-50/70 text-emerald-700 shadow-sm'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-emerald-300 hover:bg-emerald-50/40',
                ]"
                @click="judgeAnswer = 'A'"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                正确
              </button>
              <button
                type="button"
                :class="[
                  'flex cursor-pointer items-center justify-center gap-2 rounded-xl border-2 px-4 py-3 text-sm font-medium transition-all duration-200 active:scale-[0.98]',
                  judgeAnswer === 'B'
                    ? 'border-red-500 bg-red-50/70 text-red-700 shadow-sm'
                    : 'border-gray-200 bg-white text-gray-600 hover:border-red-300 hover:bg-red-50/40',
                ]"
                @click="judgeAnswer = 'B'"
              >
                <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                错误
              </button>
            </div>
          </div>
        </div>

        <!-- 多空题选项 -->
        <div v-else-if="form.type === 'multi_blank'">
          <div class="mb-2 flex items-center justify-between">
            <label class="text-sm font-medium text-gray-700">各空选项与答案</label>
            <BaseButton size="sm" type="ghost" @click="addBlank">+ 添加空</BaseButton>
          </div>
          <p class="mb-2 text-xs text-gray-400">题干中请用 __ 或 （） 标出每个填空位置，按顺序对应下列各空。</p>
          <div v-for="(blank, gi) in form.blank_options" :key="gi" class="mb-3 rounded-xl border border-gray-100 bg-gray-50/50 p-3">
            <div class="mb-2 flex items-center justify-between">
              <span class="text-xs font-semibold text-gray-700">第 {{ gi + 1 }} 空</span>
              <BaseButton v-if="form.blank_options.length > 1" size="sm" type="ghost" class="text-red-500" @click="form.blank_options.splice(gi, 1); form.blank_answers.splice(gi, 1)">删除此空</BaseButton>
            </div>
            <div class="mb-2 flex items-center gap-2">
              <span class="shrink-0 text-xs text-gray-500">正确答案</span>
              <BaseSelect v-model="form.blank_answers[gi]" :disabled="!blank.options.length">
                <option value="">请选择</option>
                <option v-for="o in blank.options" :key="o.letter" :value="o.letter.toUpperCase()">{{ o.letter.toUpperCase() }}</option>
              </BaseSelect>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <template v-for="(opt, oi) in blank.options" :key="oi">
                <div class="flex items-center gap-1 rounded-lg border border-gray-200 bg-white px-2 py-1">
                  <span class="text-xs font-semibold text-gray-500">{{ opt.letter }}</span>
                  <input v-model="opt.content" class="w-32 text-sm text-gray-800 focus:outline-none" placeholder="选项内容" />
                  <button type="button" class="cursor-pointer text-gray-300 hover:text-red-500" @click="blank.options.splice(oi, 1)">×</button>
                </div>
              </template>
              <BaseButton size="sm" type="ghost" @click="blank.options.push({ letter: String.fromCharCode(65 + blank.options.length), content: '', checked: false })">+</BaseButton>
            </div>
          </div>
          <p v-if="formErrors.answer" class="mt-1 text-xs text-red-500">{{ formErrors.answer }}</p>
        </div>

        <!-- 主观题参考答案 -->
        <div v-else>
          <BaseTextarea v-model="form.answer" label="参考答案" :rows="3" placeholder="参考答案（评分/复习参考，简答题可留空）" />
        </div>

        <!-- 解析 -->
        <div>
          <BaseTextarea v-model="form.analysis" label="解析" :rows="3" placeholder="题目解析，用于练习后查看" />
        </div>
      </div>

      <div class="mt-5 flex items-center justify-between gap-2 border-t border-gray-100 pt-4">
        <p class="text-xs text-gray-400">{{ editingId ? '保存后将立即生效' : '新题目默认启用' }}</p>
        <div class="flex items-center gap-2">
          <BaseButton type="ghost" size="sm" @click="editorOpen = false">取消</BaseButton>
          <BaseButton size="sm" :loading="saving" @click="submitForm">{{ editingId ? '保存修改' : '创建题目' }}</BaseButton>
        </div>
      </div>
    </BaseModal>

    <!-- 批量确认 -->
    <ConfirmDialog ref="confirmDialog" title="确认操作" :description="pendingActionLabel" danger confirm-text="确认执行" @confirm="runPending" />
  </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
/* transition-group 列表移动/移除动画 */
.fade-move {
  transition: transform 0.18s ease;
}
.fade-leave-active {
  position: absolute;
}
</style>

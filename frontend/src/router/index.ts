import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    meta: { guest: true },
    component: () => import('@/views/auth/Login.vue'),
  },
  {
    path: '/register',
    name: 'Register',
    meta: { guest: true },
    component: () => import('@/views/auth/Register.vue'),
  },
  {
    path: '/',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/Dashboard.vue'),
      },
      {
        path: 'subjects',
        name: 'SubjectList',
        component: () => import('@/views/subjects/SubjectList.vue'),
      },
      {
        path: 'subjects/:id',
        name: 'SubSubjectList',
        component: () => import('@/views/subjects/SubSubjectList.vue'),
      },
      {
        path: 'sub-subjects/:id/chapters',
        name: 'ChapterList',
        component: () => import('@/views/subjects/SubjectDetail.vue'),
      },
      {
        path: 'practice/chapter/:chapterId',
        name: 'ChapterPractice',
        component: () => import('@/views/practice/ChapterPractice.vue'),
      },
      {
        path: 'practice/random',
        name: 'RandomPractice',
        component: () => import('@/views/practice/RandomPractice.vue'),
      },
      {
        path: 'practice/special',
        name: 'SpecialPractice',
        component: () => import('@/views/practice/SpecialPractice.vue'),
      },
      {
        path: 'practice/wrong',
        name: 'WrongPractice',
        component: () => import('@/views/practice/WrongPractice.vue'),
      },
      {
        path: 'wrong-questions',
        name: 'WrongQuestions',
        component: () => import('@/views/practice/WrongQuestions.vue'),
      },
      {
        path: 'favorites',
        name: 'Favorites',
        component: () => import('@/views/practice/Favorites.vue'),
      },
      {
        path: 'exam',
        name: 'ExamList',
        component: () => import('@/views/exam/ExamList.vue'),
      },
      {
        path: 'exam-records',
        name: 'ExamRecords',
        component: () => import('@/views/exam/ExamRecords.vue'),
      },
      {
        path: 'exam/:id',
        name: 'ExamTaking',
        component: () => import('@/views/exam/ExamTaking.vue'),
      },
      {
        path: 'exam/:id/result',
        name: 'ExamResult',
        component: () => import('@/views/exam/ExamResult.vue'),
      },
      {
        path: 'ai/config',
        name: 'AIConfig',
        component: () => import('@/views/ai/AIConfig.vue'),
      },
      {
        path: 'ai/practice',
        name: 'AIPractice',
        component: () => import('@/views/ai/AIPractice.vue'),
      },
      {
        path: 'progress',
        name: 'Progress',
        component: () => import('@/views/progress/StudyProgress.vue'),
      },
      {
        path: 'profile',
        name: 'Profile',
        component: () => import('@/views/auth/Profile.vue'),
      },
      {
        path: 'community',
        name: 'Community',
        component: () => import('@/views/community/PostList.vue'),
      },
      {
        path: 'admin/users',
        name: 'AdminUsers',
        meta: { role: 'admin' },
        component: () => import('@/views/admin/UserManage.vue'),
      },
      {
        path: 'admin/subjects',
        name: 'AdminSubjects',
        meta: { role: 'admin' },
        component: () => import('@/views/admin/SubjectManage.vue'),
      },
      {
        path: 'admin/questions',
        name: 'AdminQuestions',
        meta: { role: 'admin' },
        component: () => import('@/views/admin/QuestionManage.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const token = localStorage.getItem('access_token')

  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }

  if (to.meta.guest && token) {
    const auth = useAuthStore()
    if (!auth.initialized) {
      await auth.checkAuth()
    }
    next(auth.getHomeRoute())
    return
  }

  if (to.meta.requiresAuth && token) {
    const auth = useAuthStore()
    if (!auth.initialized) {
      await auth.checkAuth()
    }
    if (!auth.isLoggedIn) {
      next('/login')
      return
    }
  }

  if (to.meta.role) {
    const auth = useAuthStore()
    if (!auth.user || auth.user.role !== to.meta.role) {
      next('/')
      return
    }
  }

  next()
})

export default router

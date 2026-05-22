import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/issues' },
  { path: '/issues', name: 'issues', component: () => import('./views/IssuesView.vue') },
  { path: '/epics', name: 'epics', component: () => import('./views/EpicsView.vue') },
  { path: '/board', name: 'board', component: () => import('./views/BoardView.vue') },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router

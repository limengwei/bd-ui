<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div class="header-left">
        <h1 class="app-title">{{ t('app.title') }}</h1>
        <el-select
          v-model="currentWorkspace"
          :placeholder="t('header.selectWorkspace')"
          size="small"
          style="width: 200px"
          @change="onWorkspaceChange"
        >
          <el-option
            v-for="ws in workspaceStore.workspaces"
            :key="ws.path"
            :label="workspaceStore.workspaceName(ws.path)"
            :value="ws.path"
          />
        </el-select>
        <el-menu
          :default-active="activeMenu"
          mode="horizontal"
          :ellipsis="false"
          class="header-menu"
          @select="onMenuSelect"
        >
          <el-menu-item index="/issues">{{ t('nav.issues') }}</el-menu-item>
          <el-menu-item index="/epics">{{ t('nav.epics') }}</el-menu-item>
          <el-menu-item index="/board">{{ t('nav.board') }}</el-menu-item>
        </el-menu>
      </div>
      <div class="header-right">
        <el-tag v-if="!wsConnected" type="danger" size="small">{{ t('header.disconnected') }}</el-tag>
        <el-dropdown @command="onLocaleChange" style="cursor: pointer;">
          <span class="locale-switch">
            {{ locale === 'zh-CN' ? '中文' : 'EN' }}
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="zh-CN">中文</el-dropdown-item>
              <el-dropdown-item command="en">English</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-switch
          v-model="isDark"
          :active-text="t('header.dark')"
          :inactive-text="t('header.light')"
          size="small"
          @change="toggleDark"
        />
        <el-button type="primary" size="small" @click="showNewIssue = true">
          {{ t('header.newIssue') }}
        </el-button>
      </div>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>

    <NewIssueDialog v-model="showNewIssue" />
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowDown } from '@element-plus/icons-vue'
import { useWs } from './composables/useWs'
import { useWorkspaceStore } from './stores/workspace'
import { useIssueStore } from './stores/issues'
import NewIssueDialog from './components/NewIssueDialog.vue'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const { connected: wsConnected } = useWs()
const workspaceStore = useWorkspaceStore()
const issueStore = useIssueStore()

const isDark = ref(false)
const showNewIssue = ref(false)
const currentWorkspace = ref('')

const activeMenu = computed(() => route.path)

function onMenuSelect(index) {
  router.push(index)
}

function toggleDark(val) {
  document.documentElement.classList.toggle('dark', val)
}

function onLocaleChange(lang) {
  locale.value = lang
  localStorage.setItem('beads-ui.locale', lang)
  window.location.reload()
}

async function onWorkspaceChange(path) {
  await workspaceStore.switchWorkspace(path)
  issueStore.fetchIssues()
}

onMounted(() => {
  workspaceStore.loadWorkspaces().then(() => {
    if (workspaceStore.current) {
      currentWorkspace.value = workspaceStore.current.path
    }
  })
  issueStore.fetchIssues()
})
</script>

<style>
html, body, #app {
  margin: 0;
  padding: 0;
  height: 100%;
  font-family: -apple-system, system-ui, 'Segoe UI', Roboto, 'Noto Sans SC', sans-serif;
}

.app-container {
  height: 100%;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--el-border-color);
  padding: 0 20px;
  height: 56px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.app-title {
  font-size: 20px;
  margin: 0;
  letter-spacing: 0.02em;
  color: var(--el-text-color-primary);
}

.header-menu {
  border-bottom: none !important;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.locale-switch {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  cursor: pointer;
}

.app-main {
  padding: 16px 20px;
  overflow: auto;
}
</style>

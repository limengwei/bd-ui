<template>
  <el-container class="app-container">
    <el-header class="app-header">
      <div class="header-left">
        <h1 class="app-title">{{ t('app.title') }}</h1>
        <el-dropdown trigger="click" @command="onWorkspaceCommand">
          <el-button size="small" style="max-width: 220px;">
            <span style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">
              {{ currentWorkspace ? workspaceStore.workspaceName(currentWorkspace) : t('header.selectWorkspace') }}
            </span>
            <el-icon style="margin-left: 4px;"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="ws in workspaceStore.workspaces"
                :key="ws.path"
                :command="{ action: 'switch', path: ws.path }"
                :class="{ 'is-active': currentWorkspace === ws.path }"
              >
                <div style="display: flex; align-items: center; justify-content: space-between; min-width: 200px;">
                  <span>{{ workspaceStore.workspaceName(ws.path) }}</span>
                  <el-icon
                    class="ws-remove-btn"
                    @click.stop="onRemoveWorkspace(ws.path)"
                  >
                    <Close />
                  </el-icon>
                </div>
              </el-dropdown-item>
              <el-dropdown-item divided :command="{ action: 'add' }">
                <el-icon style="margin-right: 4px;"><Plus /></el-icon>
                {{ t('workspace.addProject') }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
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
        <el-button size="small" @click="showSettings = true">
          <el-icon><Setting /></el-icon>
        </el-button>
      </div>
    </el-header>
    <el-main class="app-main">
      <router-view />
    </el-main>

    <NewIssueDialog v-model="showNewIssue" />
    <BdBinSettings v-model="showSettings" />

    <el-dialog
      v-model="showAddWorkspace"
      :title="t('workspace.addProject')"
      width="460px"
    >
      <el-input
        v-model="addWorkspacePath"
        :placeholder="t('workspace.pathPlaceholder')"
        :disabled="adding"
        @keyup.enter="onAddWorkspace"
      />
      <template #footer>
        <el-button @click="showAddWorkspace = false">{{ t('workspace.cancelBtn') }}</el-button>
        <el-button type="primary" :disabled="!addWorkspacePath.trim() || adding" @click="onAddWorkspace">
          {{ adding ? t('workspace.adding') : t('workspace.addBtn') }}
        </el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ArrowDown, Setting, Plus, Close } from '@element-plus/icons-vue'
import { useWs } from './composables/useWs'
import { useWorkspaceStore } from './stores/workspace'
import { useIssueStore } from './stores/issues'
import NewIssueDialog from './components/NewIssueDialog.vue'
import BdBinSettings from './components/BdBinSettings.vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const { connected: wsConnected } = useWs()
const workspaceStore = useWorkspaceStore()
const issueStore = useIssueStore()

const isDark = ref(true)
const showNewIssue = ref(false)
const showSettings = ref(false)
const showAddWorkspace = ref(false)
const addWorkspacePath = ref('')
const adding = ref(false)
const currentWorkspace = ref('')

const activeMenu = computed(() => route.path)

function onMenuSelect(index) {
  router.push(index)
}

function toggleDark(val) {
  document.documentElement.classList.toggle('dark', val)
  localStorage.setItem('beads-ui.dark-mode', val ? '1' : '0')
}

function onLocaleChange(lang) {
  locale.value = lang
  localStorage.setItem('beads-ui.locale', lang)
  window.location.reload()
}

function saveLastWorkspace(path) {
  if (path) {
    localStorage.setItem('beads-ui.last-workspace', path)
  } else {
    localStorage.removeItem('beads-ui.last-workspace')
  }
}

async function onWorkspaceCommand(cmd) {
  if (!cmd) return
  if (cmd.action === 'switch') {
    currentWorkspace.value = cmd.path
    saveLastWorkspace(cmd.path)
    await workspaceStore.switchWorkspace(cmd.path)
    issueStore.fetchIssues()
  } else if (cmd.action === 'add') {
    showAddWorkspace.value = true
  }
}

async function onRemoveWorkspace(path) {
  try {
    await ElMessageBox.confirm(
      t('workspace.removeConfirm', { name: workspaceStore.workspaceName(path) }),
      t('workspace.removeTitle'),
      { confirmButtonText: t('workspace.removeBtn'), cancelButtonText: t('workspace.cancelBtn'), type: 'warning' }
    )
    await workspaceStore.removeWorkspace(path)
    if (currentWorkspace.value === path) {
      currentWorkspace.value = ''
      if (workspaceStore.workspaces.length > 0) {
        const next = workspaceStore.workspaces[0].path
        currentWorkspace.value = next
        saveLastWorkspace(next)
        await workspaceStore.switchWorkspace(next)
        issueStore.fetchIssues()
      } else {
        saveLastWorkspace('')
      }
    }
  } catch {}
}

async function onAddWorkspace() {
  if (!addWorkspacePath.value.trim()) return
  adding.value = true
  try {
    await workspaceStore.addWorkspace(addWorkspacePath.value.trim())
    showAddWorkspace.value = false
    addWorkspacePath.value = ''
    if (workspaceStore.workspaces.length > 0 && !currentWorkspace.value) {
      const ws = workspaceStore.workspaces[workspaceStore.workspaces.length - 1]
      currentWorkspace.value = ws.path
      saveLastWorkspace(ws.path)
      await workspaceStore.switchWorkspace(ws.path)
      issueStore.fetchIssues()
    }
  } catch (e) {
    ElMessage.error((e && e.message) || t('workspace.addFail'))
  } finally {
    adding.value = false
  }
}

onMounted(async () => {
  const stored = localStorage.getItem('beads-ui.dark-mode')
  const dark = stored === null ? true : stored === '1'
  isDark.value = dark
  document.documentElement.classList.toggle('dark', dark)
})

watch(wsConnected, async (val) => {
  if (val) {
    await workspaceStore.loadWorkspaces()
    const lastWorkspace = localStorage.getItem('beads-ui.last-workspace')
    if (lastWorkspace && workspaceStore.workspaces.some(ws => ws.path === lastWorkspace)) {
      currentWorkspace.value = lastWorkspace
      if (workspaceStore.current && workspaceStore.current.path !== lastWorkspace) {
        await workspaceStore.switchWorkspace(lastWorkspace)
      }
    } else if (workspaceStore.current) {
      currentWorkspace.value = workspaceStore.current.path
      saveLastWorkspace(workspaceStore.current.path)
    }
    issueStore.fetchIssues()
  }
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

.ws-remove-btn {
  visibility: hidden;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  font-size: 12px;
}

.el-dropdown-menu__item:hover .ws-remove-btn {
  visibility: visible;
}

.ws-remove-btn:hover {
  color: var(--el-color-danger);
}
</style>

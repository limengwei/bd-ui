<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="header-left">
        <h1 class="app-title">{{ t('app.title') }}</h1>
        <el-dropdown trigger="click" @command="onWorkspaceCommand" class="workspace-picker">
          <span class="workspace-picker__trigger">
            <span class="workspace-picker__text">
              {{ currentWorkspace ? workspaceStore.workspaceName(currentWorkspace) : t('header.selectWorkspace') }}
            </span>
            <el-icon class="workspace-picker__arrow"><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item
                v-for="ws in workspaceStore.workspaces"
                :key="ws.path"
                :command="{ action: 'switch', path: ws.path }"
                :class="{ 'is-active': currentWorkspace === ws.path }"
              >
                <div class="ws-item">
                  <span>{{ workspaceStore.workspaceName(ws.path) }}</span>
                  <el-icon class="ws-remove-btn" @click.stop="onRemoveWorkspace(ws.path)">
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
        <nav class="header-nav">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="nav-tab"
            :class="{ active: route.path === item.path }"
          >
            {{ item.label }}
          </router-link>
        </nav>
      </div>
      <div class="header-right">
        <span v-if="!wsConnected" class="disconnect-badge">{{ t('header.disconnected') }}</span>
        <el-dropdown @command="onLocaleChange" class="locale-dropdown">
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
        <label class="theme-toggle">
          <input type="checkbox" :checked="isDark" @change="toggleDark($event.target.checked)" />
        </label>
        <button class="btn-new-issue" @click="showNewIssue = true">
          + {{ t('header.newIssue') }}
        </button>
        <button class="btn-icon" @click="showSettings = true">
          <el-icon><Setting /></el-icon>
        </button>
      </div>
    </header>

    <main class="app-main">
      <router-view />
    </main>

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
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch, onUnmounted } from 'vue'
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

const navItems = computed(() => [
  { path: '/issues', label: t('nav.issues') },
  { path: '/epics', label: t('nav.epics') },
  { path: '/board', label: t('nav.board') },
  { path: '/graph', label: t('nav.graph') },
])

function onMenuSelect(index) {
  router.push(index)
}

function toggleDark(val) {
  isDark.value = val
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

  document.addEventListener('keydown', onGlobalKeydown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onGlobalKeydown)
})

function onGlobalKeydown(e) {
  if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT' || e.target.isContentEditable) {
    return
  }

  if (e.key === 'n' || e.key === 'N') {
    e.preventDefault()
    showNewIssue.value = true
  } else if (e.key === '/') {
    e.preventDefault()
    const searchInput = document.querySelector('.search-input input')
    if (searchInput) searchInput.focus()
  } else if (e.key === '1') {
    router.push('/issues')
  } else if (e.key === '2') {
    router.push('/epics')
  } else if (e.key === '3') {
    router.push('/board')
  } else if (e.key === '4') {
    router.push('/graph')
  }
}

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
  background-color: var(--bg);
  color: var(--fg);
  font-family: -apple-system, system-ui, 'Segoe UI', Roboto, 'Noto Sans SC', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-rendering: optimizeLegibility;
  letter-spacing: 0.01em;
}

.app-shell {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.app-header {
  position: sticky;
  top: 0;
  z-index: 50;
  padding: 10px 18px 0;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: start;
  justify-content: space-between;
  background: linear-gradient(
    180deg,
    color-mix(in srgb, var(--panel-bg) 80%, transparent),
    var(--panel-bg)
  );
  backdrop-filter: saturate(150%) blur(6px);
}

.header-left {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.app-title {
  font-size: 18px;
  margin: 0;
  letter-spacing: 0.02em;
  color: var(--fg);
}

.workspace-picker__trigger {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  font-weight: 500;
  color: var(--muted);
  padding: 4px 10px;
  background: color-mix(in srgb, var(--panel-bg) 60%, transparent);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  max-width: 180px;
  transition: border-color 160ms ease;
}

.workspace-picker__trigger:hover {
  border-color: color-mix(in srgb, var(--link) 50%, var(--border));
}

.workspace-picker__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.workspace-picker__arrow {
  font-size: 12px;
  flex-shrink: 0;
}

.header-nav {
  display: flex;
  align-items: flex-end;
  gap: 4px;
}

.nav-tab {
  display: inline-block;
  padding: 8px 14px 10px;
  color: var(--muted);
  text-decoration: none;
  font-weight: 600;
  font-size: 13px;
  border: 1px solid transparent;
  border-top-left-radius: 8px;
  border-top-right-radius: 8px;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  line-height: 1.2;
  transition: color 160ms ease, background 160ms ease;
}

.nav-tab:hover,
.nav-tab:focus {
  color: var(--fg);
}

.nav-tab.active {
  color: var(--fg);
  background: var(--bg);
  border-color: var(--border);
  border-bottom-color: var(--bg);
  margin-bottom: -1px;
  padding-bottom: 11px;
  position: relative;
  z-index: 1;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.disconnect-badge {
  display: inline-block;
  padding: 0 8px;
  line-height: 20px;
  height: 20px;
  border-radius: var(--badge-radius);
  font-size: 11px;
  font-weight: 600;
  background: color-mix(in srgb, #ef4444 15%, transparent);
  color: #ef4444;
  border: 1px solid color-mix(in srgb, #ef4444 30%, transparent);
}

.locale-dropdown {
  cursor: pointer;
}

.locale-switch {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--muted);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 160ms ease;
}

.locale-switch:hover {
  background: color-mix(in srgb, var(--fg) 6%, transparent);
}

.theme-toggle {
  display: inline-flex;
  align-items: center;
  cursor: pointer;
}

.theme-toggle input[type='checkbox'] {
  --switch-h: 22px;
  appearance: none;
  position: relative;
  width: 40px;
  height: var(--switch-h);
  border-radius: var(--switch-h);
  border: 1px solid var(--border);
  background: var(--panel-bg);
  transition: background 160ms ease, border-color 160ms ease;
  cursor: pointer;
}

.theme-toggle input[type='checkbox']::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 2px;
  width: calc(var(--switch-h) - 6px);
  height: calc(var(--switch-h) - 6px);
  border-radius: 999px;
  background: var(--fg);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  transform: translate(0, -50%);
  transition: transform 180ms ease, background 180ms ease;
}

.theme-toggle input[type='checkbox']:checked::after {
  left: auto;
  right: 2px;
  transform: translate(0, -50%);
}

.btn-new-issue {
  background: var(--link);
  color: #fff;
  border: none;
  padding: 5px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 140ms ease, transform 60ms ease;
  white-space: nowrap;
}

.btn-new-issue:hover {
  background: var(--link-hover);
}

.btn-new-issue:active {
  transform: translateY(1px);
}

.btn-icon {
  background: transparent;
  color: var(--muted);
  border: 1px solid var(--border);
  padding: 4px 8px;
  border-radius: 6px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  transition: background 140ms ease, color 140ms ease, border-color 140ms ease;
}

.btn-icon:hover {
  color: var(--fg);
  border-color: color-mix(in srgb, var(--link) 50%, var(--border));
  background: color-mix(in srgb, var(--fg) 4%, transparent);
}

.app-main {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 16px 18px;
}

.ws-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-width: 200px;
}

.ws-remove-btn {
  visibility: hidden;
  color: var(--muted);
  cursor: pointer;
  font-size: 12px;
}

.el-dropdown-menu__item:hover .ws-remove-btn {
  visibility: visible;
}

.ws-remove-btn:hover {
  color: #ef4444;
}
</style>

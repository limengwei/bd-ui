<template>
  <div class="epics-view">
    <div class="epics-toolbar">
      <div class="epics-toolbar-left">
        <span class="epics-title">{{ t('epics.title') }}</span>
        <el-select v-model="statusFilter" size="small" class="filter-select" clearable
          :placeholder="t('issue.statusFilter')">
          <el-option :label="t('status.all')" value="all" />
          <el-option :label="t('status.open')" value="open" />
          <el-option :label="t('status.inProgress')" value="in_progress" />
          <el-option :label="t('status.closed')" value="closed" />
        </el-select>
      </div>
      <div class="epics-toolbar-right">
        <button class="toolbar-btn" @click="onCloseEligible" :disabled="closingEligible">
          {{ closingEligible ? t('epics.closing') : t('epics.closeEligible') }}
        </button>
        <button class="toolbar-btn" @click="issueStore.fetchIssues()" :disabled="issueStore.loading">
          {{ t('issue.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="!issueStore.loading && filteredEpics.length === 0" class="empty-state">
      <div class="empty-icon">⛰️</div>
      <div class="empty-text">{{ t('epics.empty') }}</div>
    </div>

    <div v-else class="epics-list">
      <div v-for="epic in filteredEpics" :key="epic.id" class="epic-card">
        <div class="epic-header" @click="toggleExpand(epic.id)">
          <div class="epic-header-left">
            <span class="expand-icon" :class="{ expanded: expandedEpics.has(epic.id) }">▶</span>
            <span class="beads-badge beads-badge--id">{{ epic.id }}</span>
            <span class="epic-name">{{ epic.title }}</span>
          </div>
          <div class="epic-header-right">
            <span class="beads-badge" :class="statusBadgeClass(epic.status)">{{ statusLabel(epic.status) }}</span>
            <span class="children-count">{{ epic.closed_children || 0 }} / {{ epic.total_children || 0 }}</span>
            <el-progress :percentage="epicProgress(epic)" :status="epicProgress(epic) >= 100 ? 'success' : ''"
              :stroke-width="10" :show-text="true" class="epic-progress" />
          </div>
        </div>

        <div v-if="expandedEpics.has(epic.id)" class="epic-children">
          <div v-if="getChildren(epic.id).length === 0" class="children-empty">
            {{ t('epics.noChildren') }}
          </div>
          <div v-for="child in getChildren(epic.id)" :key="child.id" class="child-row" @click="onChildClick(child)">
            <span class="beads-badge beads-badge--id">{{ child.id }}</span>
            <span class="child-title">{{ child.title }}</span>
            <span class="beads-badge" :class="statusBadgeClass(child.status)">{{ statusLabel(child.status) }}</span>
            <span v-if="child.priority != null" class="beads-badge" :class="'beads-badge--p' + child.priority">P{{
              child.priority }}</span>
            <span v-if="child.issue_type" class="beads-badge" :class="'beads-badge--' + child.issue_type">{{
              typeLabel(child.issue_type) }}</span>
          </div>
        </div>
      </div>
    </div>

    <IssueDetail v-model="showDetail" :issue-id="selectedIssueId" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import { useWs } from '../composables/useWs'
import IssueDetail from '../components/IssueDetail.vue'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const issueStore = useIssueStore()
const { send } = useWs()

const statusFilter = ref('all')
const expandedEpics = ref(new Set())
const showDetail = ref(false)
const selectedIssueId = ref(null)
const closingEligible = ref(false)

const epics = computed(() => issueStore.issues.filter(i => i.issue_type === 'epic'))

const filteredEpics = computed(() => {
  if (statusFilter.value === 'all' || !statusFilter.value) return epics.value
  return epics.value.filter(e => e.status === statusFilter.value)
})

function getChildren(epicId) {
  return issueStore.parentChildMap.get(epicId) || []
}

function epicProgress(row) {
  const total = row.total_children || 0
  if (total === 0) {
    if (row.status === 'closed') {
      return 100
    }
    return 0
  }
  return Math.round(((row.closed_children || 0) / total) * 100)
}

function toggleExpand(epicId) {
  const s = new Set(expandedEpics.value)
  if (s.has(epicId)) {
    s.delete(epicId)
  } else {
    s.add(epicId)
  }
  expandedEpics.value = s
}

function onChildClick(child) {
  selectedIssueId.value = child.id
  showDetail.value = true
}

async function onCloseEligible() {
  closingEligible.value = true
  try {
    await send('bd-command', { args: ['epic', 'close-eligible'] })
    ElMessage.success(t('epics.closeEligibleSuccess'))
    await issueStore.fetchIssues()
  } catch (e) {
    ElMessage.error(t('epics.closeEligibleFail'))
  } finally {
    closingEligible.value = false
  }
}

function statusBadgeClass(status) {
  return {
    open: 'beads-badge--open',
    in_progress: 'beads-badge--in-progress',
    closed: 'beads-badge--closed',
  }[status] || ''
}

function statusLabel(status) {
  return { open: t('status.open'), in_progress: t('status.inProgress'), closed: t('status.closed') }[status] || status
}

function typeLabel(type) {
  return { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }[type] || type
}

onMounted(() => {
  if (issueStore.issues.length === 0) issueStore.fetchIssues()
})
</script>

<style scoped>
.epics-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.epics-toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.epics-toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.epics-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
}

.filter-select {
  width: 120px;
}

.toolbar-btn {
  background: var(--panel-bg);
  color: var(--fg);
  border: 1px solid var(--border);
  padding: 5px 14px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: background 140ms ease, border-color 140ms ease, transform 60ms ease;
}

.toolbar-btn:hover {
  border-color: color-mix(in srgb, var(--link) 50%, var(--border));
}

.toolbar-btn:disabled {
  opacity: 0.6;
  cursor: default;
}

.epics-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.epic-card {
  background: var(--panel-bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
}

.epic-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 140ms ease;
}

.epic-header:hover {
  background: color-mix(in srgb, var(--fg) 2%, transparent);
}

.epic-header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
  min-width: 0;
}

.expand-icon {
  font-size: 10px;
  color: var(--muted);
  transition: transform 200ms ease;
  flex-shrink: 0;
}

.expand-icon.expanded {
  transform: rotate(90deg);
}

.epic-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.epic-header-right {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.children-count {
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.epic-progress {
  width: 140px;
}

.epic-children {
  border-top: 1px solid var(--border);
  padding: 8px 16px 12px 40px;
}

.children-empty {
  color: var(--muted);
  font-size: 13px;
  font-style: italic;
  padding: 8px 0;
}

.child-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 140ms ease;
}

.child-row:hover {
  background: color-mix(in srgb, var(--fg) 3%, transparent);
}

.child-title {
  flex: 1;
  font-size: 13px;
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  color: var(--muted);
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 14px;
}
</style>

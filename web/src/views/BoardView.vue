<template>
  <div class="board-view">
    <div class="board-toolbar">
      <el-select v-model="closedFilter" size="small" class="filter-select">
        <el-option :label="t('board.closedToday')" value="today" />
        <el-option :label="t('board.closed3d')" value="3" />
        <el-option :label="t('board.closed7d')" value="7" />
      </el-select>
      <button class="toolbar-btn" @click="issueStore.fetchIssues()" :disabled="issueStore.loading">
        {{ t('issue.refresh') }}
      </button>
    </div>

    <div class="board-columns">
      <div
        v-for="col in columns"
        :key="col.key"
        class="board-column"
        :class="{ 'drag-over': dragOverCol === col.key }"
        @dragover.prevent="onDragOver($event, col.key)"
        @dragleave="onDragLeave(col.key)"
        @drop="onDrop($event, col.key)"
      >
        <div class="column-header" :class="'column-header--' + col.key">
          <span class="column-title">{{ col.label }}</span>
          <span class="column-count beads-badge" :class="col.countClass">{{ col.items.length }}</span>
        </div>
        <div class="column-body">
          <div
            v-for="issue in col.items"
            :key="issue.id"
            class="board-card"
            :class="{ 'board-card--closed': col.key === 'closed', 'dragging': dragIssueId === issue.id }"
            draggable="true"
            @dragstart="onDragStart($event, issue)"
            @dragend="onDragEnd"
            @click="onCardClick(issue)"
          >
            <div class="card-header">
              <span class="card-id">{{ issue.id }}</span>
              <span v-if="issue.priority != null" class="beads-badge" :class="'beads-badge--p' + issue.priority">{{ formatPriority(issue.priority) }}</span>
            </div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-badges">
              <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
              <span v-for="label in (issue.labels || []).slice(0, 2)" :key="label" class="card-label">{{ label }}</span>
              <span v-if="issue.labels && issue.labels.length > 2" class="card-label">+{{ issue.labels.length - 2 }}</span>
            </div>
            <div v-if="issue.assignee" class="card-assignee">
              <span class="assignee-avatar">{{ issue.assignee.charAt(0).toUpperCase() }}</span>
              <span class="assignee-name">{{ issue.assignee }}</span>
            </div>
          </div>
          <div v-if="col.items.length === 0" class="column-empty">
            {{ t('board.empty') }}
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
import IssueDetail from '../components/IssueDetail.vue'

const { t } = useI18n()
const issueStore = useIssueStore()
const closedFilter = ref('today')
const showDetail = ref(false)
const selectedIssueId = ref(null)
const dragIssueId = ref(null)
const dragOverCol = ref(null)

const filteredClosed = computed(() => {
  const now = Date.now()
  let since = 0
  if (closedFilter.value === 'today') {
    since = new Date(new Date().toDateString()).getTime()
  } else {
    since = now - (parseInt(closedFilter.value) || 7) * 86400000
  }
  return issueStore.closedIssues.filter(i => {
    if (!i.closed_at) return false
    const ct = typeof i.closed_at === 'number' ? i.closed_at : new Date(i.closed_at).getTime()
    return ct >= since
  })
})

const columns = computed(() => [
  {
    key: 'blocked',
    label: t('board.blocked'),
    items: issueStore.blockedIssues,
    status: 'open',
    countClass: 'beads-badge--id',
  },
  {
    key: 'ready',
    label: t('board.ready'),
    items: issueStore.readyIssues,
    status: 'open',
    countClass: 'beads-badge--open',
  },
  {
    key: 'in-progress',
    label: t('board.inProgress'),
    items: issueStore.inProgressIssues,
    status: 'in_progress',
    countClass: 'beads-badge--in-progress',
  },
  {
    key: 'closed',
    label: t('board.closed'),
    items: filteredClosed.value,
    status: 'closed',
    countClass: 'beads-badge--closed',
  },
])

function onCardClick(issue) {
  selectedIssueId.value = issue.id
  showDetail.value = true
}

function onDragStart(event, issue) {
  dragIssueId.value = issue.id
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/plain', issue.id)
  event.target.style.opacity = '0.5'
}

function onDragEnd(event) {
  dragIssueId.value = null
  dragOverCol.value = null
  event.target.style.opacity = '1'
}

function onDragOver(event, colKey) {
  event.dataTransfer.dropEffect = 'move'
  dragOverCol.value = colKey
}

function onDragLeave(colKey) {
  if (dragOverCol.value === colKey) {
    dragOverCol.value = null
  }
}

async function onDrop(event, colKey) {
  dragOverCol.value = null
  const issueId = event.dataTransfer.getData('text/plain')
  if (!issueId) return

  const col = columns.value.find(c => c.key === colKey)
  if (!col) return

  const targetStatus = col.status
  const issue = issueStore.issues.find(i => i.id === issueId)
  if (!issue) return

  if (issue.status === targetStatus) return

  await issueStore.updateStatus(issueId, targetStatus)
}

function typeLabel(type) {
  return { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }[type] || type
}

function formatPriority(priority) {
  return 'P' + priority
}

onMounted(() => {
  if (issueStore.issues.length === 0) issueStore.fetchIssues()
})
</script>

<style scoped>
.board-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}

.filter-select {
  width: 140px;
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

.board-columns {
  display: flex;
  gap: 12px;
  height: calc(100vh - 180px);
  overflow-x: auto;
}

.board-column {
  flex: 1;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  background: var(--panel-bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 200ms ease, box-shadow 200ms ease;
}

.board-column.drag-over {
  border-color: var(--link);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--link) 20%, transparent);
}

.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border);
}

.column-header--blocked {
  border-top: 3px solid #ef4444;
}

.column-header--ready {
  border-top: 3px solid var(--status-open-base);
}

.column-header--in-progress {
  border-top: 3px solid #e6a23c;
}

.column-header--closed {
  border-top: 3px solid var(--status-closed-base);
}

.column-title {
  font-weight: 600;
  font-size: 13px;
  color: var(--fg);
}

.column-count {
  font-size: 11px;
}

.column-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.column-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 12px;
  color: var(--muted);
  font-size: 12px;
  font-style: italic;
  min-height: 60px;
}

.board-card {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px 12px;
  margin-bottom: 6px;
  cursor: grab;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease, opacity 0.15s ease;
}

.board-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--fg) 8%, transparent);
  border-color: color-mix(in srgb, var(--link) 30%, var(--border));
}

.board-card:active {
  cursor: grabbing;
}

.board-card.dragging {
  opacity: 0.5;
}

.board-card--closed {
  opacity: 0.75;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.card-id {
  font-size: 11px;
  color: var(--muted);
  font-weight: 500;
}

.card-title {
  font-size: 13px;
  line-height: 1.4;
  margin-bottom: 6px;
  color: var(--fg);
}

.card-badges {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.card-label {
  display: inline-block;
  padding: 0 6px;
  line-height: 18px;
  height: 18px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 600;
  background: color-mix(in srgb, var(--panel-bg) 85%, #6b7280);
  color: color-mix(in srgb, #6b7280 85%, #000);
  border: 1px solid color-mix(in srgb, currentColor 20%, transparent);
}

.card-assignee {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
  padding-top: 6px;
  border-top: 1px solid color-mix(in srgb, var(--border) 60%, transparent);
}

.assignee-avatar {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--link) 20%, var(--panel-bg));
  color: var(--link);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
}

.assignee-name {
  font-size: 11px;
  color: var(--muted);
}
</style>

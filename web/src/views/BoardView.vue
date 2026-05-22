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
      <div class="board-column">
        <div class="column-header column-header--blocked">
          <span class="column-title">{{ t('board.blocked') }}</span>
          <span class="column-count beads-badge beads-badge--id">{{ issueStore.blockedIssues.length }}</span>
        </div>
        <div class="column-body">
          <div v-for="issue in issueStore.blockedIssues" :key="issue.id" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-badges">
              <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
              <span v-if="issue.priority != null" class="beads-badge" :class="'beads-badge--p' + issue.priority">{{ formatPriority(issue.priority) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header column-header--ready">
          <span class="column-title">{{ t('board.ready') }}</span>
          <span class="column-count beads-badge beads-badge--open">{{ issueStore.readyIssues.length }}</span>
        </div>
        <div class="column-body">
          <div v-for="issue in issueStore.readyIssues" :key="issue.id" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-badges">
              <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
              <span v-if="issue.priority != null" class="beads-badge" :class="'beads-badge--p' + issue.priority">{{ formatPriority(issue.priority) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header column-header--in-progress">
          <span class="column-title">{{ t('board.inProgress') }}</span>
          <span class="column-count beads-badge beads-badge--in-progress">{{ issueStore.inProgressIssues.length }}</span>
        </div>
        <div class="column-body">
          <div v-for="issue in issueStore.inProgressIssues" :key="issue.id" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-badges">
              <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
              <span v-if="issue.priority != null" class="beads-badge" :class="'beads-badge--p' + issue.priority">{{ formatPriority(issue.priority) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header column-header--closed">
          <span class="column-title">{{ t('board.closed') }}</span>
          <span class="column-count beads-badge beads-badge--closed">{{ filteredClosed.length }}</span>
        </div>
        <div class="column-body">
          <div v-for="issue in filteredClosed" :key="issue.id" class="board-card board-card--closed" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-badges">
              <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
            </div>
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

function onCardClick(issue) {
  selectedIssueId.value = issue.id
  showDetail.value = true
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

.board-card {
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px 12px;
  margin-bottom: 6px;
  cursor: pointer;
  transition: transform 0.15s ease, box-shadow 0.15s ease, border-color 0.15s ease;
}

.board-card:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px color-mix(in srgb, var(--fg) 8%, transparent);
  border-color: color-mix(in srgb, var(--link) 30%, var(--border));
}

.board-card--closed {
  opacity: 0.75;
}

.card-id {
  font-size: 11px;
  color: var(--muted);
  margin-bottom: 4px;
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
</style>

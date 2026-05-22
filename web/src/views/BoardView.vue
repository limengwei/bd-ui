<template>
  <div class="board-view">
    <div class="board-toolbar">
      <h3 style="margin: 0;">{{ t('board.title') }}</h3>
      <el-select v-model="closedFilter" size="small" style="width: 140px">
        <el-option :label="t('board.closedToday')" value="today" />
        <el-option :label="t('board.closed3d')" value="3" />
        <el-option :label="t('board.closed7d')" value="7" />
      </el-select>
      <el-button @click="issueStore.fetchIssues()" :loading="issueStore.loading" size="small">{{ t('issue.refresh') }}</el-button>
    </div>

    <div class="board-columns">
      <div class="board-column">
        <div class="column-header blocked">
          <span>{{ t('board.blocked') }}</span>
          <el-tag size="small" round>{{ issueStore.blockedIssues.length }}</el-tag>
        </div>
        <div class="column-body">
          <el-card v-for="issue in issueStore.blockedIssues" :key="issue.id" shadow="hover" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-tags">
              <el-tag v-if="issue.issue_type" :type="typeTagType(issue.issue_type)" size="small">{{ typeLabel(issue.issue_type) }}</el-tag>
              <el-tag v-if="issue.priority" :type="priorityTagType(issue.priority)" size="small">{{ issue.priority.toUpperCase() }}</el-tag>
            </div>
          </el-card>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header ready">
          <span>{{ t('board.ready') }}</span>
          <el-tag size="small" round type="success">{{ issueStore.readyIssues.length }}</el-tag>
        </div>
        <div class="column-body">
          <el-card v-for="issue in issueStore.readyIssues" :key="issue.id" shadow="hover" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-tags">
              <el-tag v-if="issue.issue_type" :type="typeTagType(issue.issue_type)" size="small">{{ typeLabel(issue.issue_type) }}</el-tag>
              <el-tag v-if="issue.priority" :type="priorityTagType(issue.priority)" size="small">{{ issue.priority.toUpperCase() }}</el-tag>
            </div>
          </el-card>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header in-progress">
          <span>{{ t('board.inProgress') }}</span>
          <el-tag size="small" round type="warning">{{ issueStore.inProgressIssues.length }}</el-tag>
        </div>
        <div class="column-body">
          <el-card v-for="issue in issueStore.inProgressIssues" :key="issue.id" shadow="hover" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-tags">
              <el-tag v-if="issue.issue_type" :type="typeTagType(issue.issue_type)" size="small">{{ typeLabel(issue.issue_type) }}</el-tag>
              <el-tag v-if="issue.priority" :type="priorityTagType(issue.priority)" size="small">{{ issue.priority.toUpperCase() }}</el-tag>
            </div>
          </el-card>
        </div>
      </div>

      <div class="board-column">
        <div class="column-header closed">
          <span>{{ t('board.closed') }}</span>
          <el-tag size="small" round type="info">{{ filteredClosed.length }}</el-tag>
        </div>
        <div class="column-body">
          <el-card v-for="issue in filteredClosed" :key="issue.id" shadow="hover" class="board-card" @click="onCardClick(issue)">
            <div class="card-id">{{ issue.id }}</div>
            <div class="card-title">{{ issue.title }}</div>
            <div class="card-tags">
              <el-tag v-if="issue.issue_type" :type="typeTagType(issue.issue_type)" size="small">{{ typeLabel(issue.issue_type) }}</el-tag>
            </div>
          </el-card>
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

function typeTagType(type) {
  return { bug: 'danger', feature: 'success', task: 'warning', epic: '', chore: 'info' }[type] || ''
}

function typeLabel(type) {
  return { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }[type] || type
}

function priorityTagType(p) {
  return { p0: 'danger', p1: 'warning', p2: '', p3: 'success', p4: 'info' }[p] || ''
}

onMounted(() => {
  if (issueStore.issues.length === 0) issueStore.fetchIssues()
})
</script>

<style scoped>
.board-toolbar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.board-columns {
  display: flex;
  gap: 16px;
  height: calc(100vh - 180px);
  overflow-x: auto;
}
.board-column {
  flex: 1;
  min-width: 260px;
  display: flex;
  flex-direction: column;
  background: var(--el-fill-color-lighter);
  border-radius: 8px;
  overflow: hidden;
}
.column-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  font-weight: 600;
  font-size: 14px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.column-header.blocked { border-top: 3px solid #f56c6c; }
.column-header.ready { border-top: 3px solid #67c23a; }
.column-header.in-progress { border-top: 3px solid #e6a23c; }
.column-header.closed { border-top: 3px solid #909399; }
.column-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}
.board-card {
  margin-bottom: 8px;
  cursor: pointer;
  transition: transform 0.15s;
}
.board-card:hover { transform: translateY(-1px); }
.card-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}
.card-title {
  font-size: 13px;
  line-height: 1.4;
  margin-bottom: 6px;
}
.card-tags {
  display: flex;
  gap: 4px;
}
</style>

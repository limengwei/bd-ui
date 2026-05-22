<template>
  <div class="epics-view">
    <div class="epics-toolbar">
      <span class="epics-title">{{ t('epics.title') }}</span>
      <button class="toolbar-btn" @click="issueStore.fetchIssues()" :disabled="issueStore.loading">
        {{ t('issue.refresh') }}
      </button>
    </div>

    <el-table :data="epics" v-loading="issueStore.loading" style="width: 100%" class="epics-table">
      <el-table-column prop="id" :label="t('issue.id')" width="100">
        <template #default="{ row }">
          <span class="beads-badge beads-badge--id">{{ row.id }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="title" :label="t('epics.name')" min-width="250" show-overflow-tooltip />
      <el-table-column :label="t('issue.status')" width="110">
        <template #default="{ row }">
          <span class="beads-badge" :class="statusBadgeClass(row.status)">{{ statusLabel(row.status) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('epics.progress')" width="250">
        <template #default="{ row }">
          <div class="progress-cell">
            <el-progress
              :percentage="epicProgress(row)"
              :status="epicProgress(row) >= 100 ? 'success' : ''"
              :stroke-width="14"
              :text-inside="true"
            />
          </div>
        </template>
      </el-table-column>
      <el-table-column :label="t('epics.children')" width="120">
        <template #default="{ row }">
          <span class="children-count">{{ row.closed_children || 0 }} / {{ row.total_children || 0 }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.updatedAt')" width="170">
        <template #default="{ row }"><span class="muted-text">{{ formatTime(row.updated_at) }}</span></template>
      </el-table-column>
    </el-table>

    <div v-if="!issueStore.loading && epics.length === 0" class="empty-state">
      <div class="empty-icon">⛰️</div>
      <div class="empty-text">{{ t('epics.empty') }}</div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'

const { t } = useI18n()
const issueStore = useIssueStore()

const epics = computed(() => issueStore.issues.filter(i => i.issue_type === 'epic'))

function epicProgress(row) {
  const total = row.total_children || 0
  if (total === 0) return 0
  return Math.round(((row.closed_children || 0) / total) * 100)
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

function formatTime(ts) {
  if (!ts) return '-'
  return new Date(typeof ts === 'number' ? ts : ts).toLocaleString()
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

.epics-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
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

.progress-cell {
  display: flex;
  align-items: center;
}

.children-count {
  font-size: 12px;
  color: var(--muted);
  font-variant-numeric: tabular-nums;
}

.muted-text {
  color: var(--muted);
  font-size: 12px;
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

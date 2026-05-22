<template>
  <div class="issues-view">
    <div class="issues-toolbar">
      <el-input
        v-model="issueStore.searchText"
        :placeholder="t('issue.search')"
        prefix-icon="Search"
        clearable
        style="width: 240px"
      />
      <el-select v-model="issueStore.filterStatus" :placeholder="t('issue.statusFilter')" clearable style="width: 120px">
        <el-option :label="t('status.all')" value="all" />
        <el-option :label="t('status.open')" value="open" />
        <el-option :label="t('status.inProgress')" value="in_progress" />
        <el-option :label="t('status.closed')" value="closed" />
        <el-option :label="t('status.ready')" value="ready" />
      </el-select>
      <el-select v-model="issueStore.filterType" :placeholder="t('issue.typeFilter')" clearable style="width: 120px">
        <el-option :label="t('type.bug')" value="bug" />
        <el-option :label="t('type.feature')" value="feature" />
        <el-option :label="t('type.task')" value="task" />
        <el-option :label="t('type.epic')" value="epic" />
        <el-option :label="t('type.chore')" value="chore" />
      </el-select>
      <el-button @click="issueStore.fetchIssues()" :loading="issueStore.loading">
        {{ t('issue.refresh') }}
      </el-button>
      <span class="issue-count">{{ t('issue.total', { count: issueStore.filteredIssues.length }) }}</span>
    </div>

    <el-table
      :data="issueStore.filteredIssues"
      stripe
      highlight-current-row
      @row-click="onRowClick"
      v-loading="issueStore.loading"
      style="width: 100%"
      max-height="calc(100vh - 160px)"
    >
      <el-table-column prop="id" :label="t('issue.id')" width="100" show-overflow-tooltip />
      <el-table-column prop="title" :label="t('issue.title')" min-width="300" show-overflow-tooltip />
      <el-table-column :label="t('issue.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.type')" width="90">
        <template #default="{ row }">
          <el-tag v-if="row.issue_type" :type="typeTagType(row.issue_type)" size="small">{{ typeLabel(row.issue_type) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.priority')" width="80">
        <template #default="{ row }">
          <el-tag v-if="row.priority != null" :type="priorityTagType(row.priority)" size="small">{{ formatPriority(row.priority) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.updatedAt')" width="170">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
      <el-table-column :label="t('issue.actions')" width="160" fixed="right">
        <template #default="{ row }">
          <el-dropdown trigger="click">
            <el-button size="small" text>{{ t('issue.statusAction') }} ▾</el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'open')">{{ t('status.open') }}</el-dropdown-item>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'in_progress')">{{ t('status.inProgress') }}</el-dropdown-item>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'closed')">{{ t('status.closed') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <el-button size="small" text type="danger" @click.stop="onDelete(row)">{{ t('issue.delete') }}</el-button>
        </template>
      </el-table-column>
    </el-table>

    <IssueDetail v-model="showDetail" :issue-id="selectedIssueId" />
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import IssueDetail from '../components/IssueDetail.vue'
import { ElMessage, ElMessageBox } from 'element-plus'

const { t } = useI18n()
const issueStore = useIssueStore()

const showDetail = ref(false)
const selectedIssueId = ref(null)

function onRowClick(row) {
  selectedIssueId.value = row.id
  showDetail.value = true
}

function onDelete(row) {
  ElMessageBox.confirm(
    t('confirm.delete', { title: row.title }),
    t('confirm.deleteTitle'),
    {
      confirmButtonText: t('confirm.confirmBtn'),
      cancelButtonText: t('confirm.cancelBtn'),
      type: 'warning',
    }
  ).then(() => {
    issueStore.deleteIssue(row.id)
    ElMessage.success('OK')
  }).catch(() => {})
}

function statusTagType(status) {
  return { open: 'success', in_progress: 'warning', closed: 'info' }[status] || ''
}

function statusLabel(status) {
  return { open: t('status.open'), in_progress: t('status.inProgress'), closed: t('status.closed') }[status] || status
}

function typeTagType(type) {
  return { bug: 'danger', feature: 'success', task: 'warning', epic: '', chore: 'info' }[type] || ''
}

function typeLabel(type) {
  return { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }[type] || type
}

function priorityTagType(priority) {
  return { 0: 'danger', 1: 'warning', 2: '', 3: 'success', 4: 'info' }[priority] || ''
}

function formatPriority(priority) {
  return 'P' + priority
}

function formatTime(ts) {
  if (!ts) return '-'
  return new Date(typeof ts === 'number' ? ts : ts).toLocaleString()
}
</script>

<style scoped>
.issues-view {
  height: 100%;
}
.issues-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.issue-count {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>

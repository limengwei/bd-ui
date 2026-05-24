<template>
  <div class="issues-view">
    <div class="issues-toolbar">
      <el-input
        v-model="issueStore.searchText"
        :placeholder="t('issue.search')"
        prefix-icon="Search"
        clearable
        class="search-input"
      />
      <el-select v-model="issueStore.filterStatus" :placeholder="t('issue.statusFilter')" clearable class="filter-select">
        <el-option :label="t('status.all')" value="all" />
        <el-option :label="t('status.open')" value="open" />
        <el-option :label="t('status.inProgress')" value="in_progress" />
        <el-option :label="t('status.closed')" value="closed" />
        <el-option :label="t('status.ready')" value="ready" />
      </el-select>
      <el-select v-model="issueStore.filterType" :placeholder="t('issue.typeFilter')" clearable class="filter-select">
        <el-option :label="t('type.bug')" value="bug" />
        <el-option :label="t('type.feature')" value="feature" />
        <el-option :label="t('type.task')" value="task" />
        <el-option :label="t('type.epic')" value="epic" />
        <el-option :label="t('type.chore')" value="chore" />
      </el-select>
      <el-select v-model="issueStore.filterAssignee" :placeholder="t('issue.assigneeFilter')" clearable class="filter-select">
        <el-option :label="t('issue.unassigned')" value="__none__" />
        <el-option v-for="a in issueStore.allAssignees" :key="a" :label="a" :value="a" />
      </el-select>
      <el-select v-model="issueStore.filterLabel" :placeholder="t('issue.labelFilter')" clearable class="filter-select">
        <el-option v-for="l in issueStore.allLabels" :key="l" :label="l" :value="l" />
      </el-select>
      <el-select v-model="issueStore.sortBy" class="filter-select sort-select">
        <el-option :label="t('issue.sortUpdated')" value="updated_at" />
        <el-option :label="t('issue.sortCreated')" value="created_at" />
        <el-option :label="t('issue.sortPriority')" value="priority" />
      </el-select>
      <button class="sort-order-btn" @click="toggleSortOrder" :title="issueStore.sortOrder === 'desc' ? t('issue.desc') : t('issue.asc')">
        {{ issueStore.sortOrder === 'desc' ? '↓' : '↑' }}
      </button>
      <button class="toolbar-btn" @click="issueStore.clearFilters()" :title="t('issue.clearFilters')">
        ✕
      </button>
      <span class="issue-count">{{ t('issue.total', { count: issueStore.filteredIssues.length }) }}</span>
    </div>

    <el-table
      :data="pagedIssues"
      highlight-current-row
      @row-click="onRowClick"
      v-loading="issueStore.loading"
      style="width: 100%"
      max-height="calc(100vh - 200px)"
      class="issues-table"
    >
      <el-table-column prop="id" :label="t('issue.id')" width="300" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="beads-badge beads-badge--id">{{ row.id }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="title" :label="t('issue.title')" min-width="300" show-overflow-tooltip />
      <el-table-column :label="t('issue.status')" width="110">
        <template #default="{ row }">
          <span class="beads-badge" :class="statusBadgeClass(row.status)">{{ statusLabel(row.status) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.type')" width="100">
        <template #default="{ row }">
          <span v-if="row.issue_type" class="beads-badge" :class="'beads-badge--' + row.issue_type">{{ typeLabel(row.issue_type) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.priority')" width="80">
        <template #default="{ row }">
          <span v-if="row.priority != null" class="beads-badge" :class="'beads-badge--p' + row.priority">{{ formatPriority(row.priority) }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.assignee')" width="100">
        <template #default="{ row }">
          <span class="muted-text">{{ row.assignee || '-' }}</span>
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.updatedAt')" width="170">
        <template #default="{ row }"><span class="muted-text">{{ formatTime(row.updated_at) }}</span></template>
      </el-table-column>
      <el-table-column :label="t('issue.actions')" width="160" fixed="right">
        <template #default="{ row }">
          <div @click.stop>
          <el-dropdown trigger="click">
            <button class="action-btn">{{ t('issue.statusAction') }} ▾</button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'open')">{{ t('status.open') }}</el-dropdown-item>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'in_progress')">{{ t('status.inProgress') }}</el-dropdown-item>
                <el-dropdown-item @click="issueStore.updateStatus(row.id, 'closed')">{{ t('status.closed') }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-row">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :page-sizes="[50, 100, 200, 500]"
        :total="issueStore.filteredIssues.length"
        layout="total, sizes, prev, pager, next"
        background
        small
      />
    </div>

    <IssueDetail v-model="showDetail" :issue-id="selectedIssueId" />
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import IssueDetail from '../components/IssueDetail.vue'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const issueStore = useIssueStore()

const showDetail = ref(false)
const selectedIssueId = ref(null)
const currentPage = ref(1)
const pageSize = ref(100)

const pagedIssues = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  return issueStore.filteredIssues.slice(start, start + pageSize.value)
})

watch([
  () => issueStore.filterStatus,
  () => issueStore.filterType,
  () => issueStore.filterAssignee,
  () => issueStore.filterLabel,
  () => issueStore.searchText,
], () => {
  currentPage.value = 1
})

function onRowClick(row) {
  selectedIssueId.value = row.id
  showDetail.value = true
}

function toggleSortOrder() {
  issueStore.sortOrder = issueStore.sortOrder === 'desc' ? 'asc' : 'desc'
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
  gap: 8px;
  margin-bottom: 14px;
  flex-wrap: wrap;
}

.search-input {
  width: 200px;
}

.filter-select {
  width: 120px;
}

.sort-select {
  width: 100px;
}

.sort-order-btn {
  background: var(--panel-bg);
  color: var(--fg);
  border: 1px solid var(--border);
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
  transition: background 140ms ease, border-color 140ms ease;
  line-height: 1;
}

.sort-order-btn:hover {
  border-color: color-mix(in srgb, var(--link) 50%, var(--border));
}

.toolbar-btn {
  background: var(--panel-bg);
  color: var(--muted);
  border: 1px solid var(--border);
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  transition: background 140ms ease, border-color 140ms ease;
}

.toolbar-btn:hover {
  border-color: color-mix(in srgb, var(--link) 50%, var(--border));
  color: var(--fg);
}

.issue-count {
  color: var(--muted);
  font-size: 12px;
  margin-left: auto;
}

.muted-text {
  color: var(--muted);
  font-size: 12px;
}

.action-btn {
  background: transparent;
  color: var(--link);
  border: none;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  transition: background 140ms ease;
}

.action-btn:hover {
  background: color-mix(in srgb, var(--link) 8%, transparent);
}

.action-btn--danger {
  color: #ef4444;
}

.action-btn--danger:hover {
  background: color-mix(in srgb, #ef4444 8%, transparent);
}

.issues-table :deep(.el-table__row) {
  cursor: pointer;
  transition: background 120ms ease;
}

.pagination-row {
  display: flex;
  justify-content: center;
  padding: 12px 0 4px;
}
</style>

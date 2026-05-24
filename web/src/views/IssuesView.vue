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

    <div class="issues-table-wrapper">
      <el-table-v2
        :columns="columns"
        :data="issueStore.filteredIssues"
        :width="tableWidth"
        :height="tableHeight"
        :row-key="rowKey"
        :row-class="rowClass"
        :row-event-handlers="rowEventHandlers"
        fixed
      >
        <template #overlay v-if="issueStore.loading">
          <div class="table-loading-overlay">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>{{ t('detail.loading') }}</span>
          </div>
        </template>
        <template #empty>
          <div class="table-empty">{{ t('board.empty') }}</div>
        </template>
      </el-table-v2>
    </div>

    <IssueDetail v-model="showDetail" :issue-id="selectedIssueId" />
  </div>
</template>

<script setup>
import { ref, computed, h, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import { Loading } from '@element-plus/icons-vue'
import { ElDropdown, ElDropdownMenu, ElDropdownItem } from 'element-plus'
import IssueDetail from '../components/IssueDetail.vue'

const { t } = useI18n()
const issueStore = useIssueStore()

const showDetail = ref(false)
const selectedIssueId = ref(null)
const tableWidth = ref(1200)
const tableHeight = ref(600)

function updateTableSize() {
  const wrapper = document.querySelector('.issues-table-wrapper')
  if (wrapper) {
    tableWidth.value = wrapper.clientWidth
    tableHeight.value = wrapper.clientHeight
  }
}

onMounted(() => {
  updateTableSize()
  window.addEventListener('resize', updateTableSize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateTableSize)
})

function rowKey({ id }) {
  return id
}

function rowClass({ rowIndex }) {
  return 'issues-v2-row'
}

const rowEventHandlers = {
  onClick: ({ rowData }) => {
    selectedIssueId.value = rowData.id
    showDetail.value = true
  },
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

function toggleSortOrder() {
  issueStore.sortOrder = issueStore.sortOrder === 'desc' ? 'asc' : 'desc'
}

function cellRendererId({ cellData }) {
  return h('span', { class: 'beads-badge beads-badge--id cell-overflow' }, cellData)
}

function cellRendererTitle({ cellData }) {
  return h('span', { class: 'cell-overflow', title: cellData }, cellData)
}

function cellRendererStatus({ rowData }) {
  return h('span', {
    class: `beads-badge ${statusBadgeClass(rowData.status)}`,
  }, statusLabel(rowData.status))
}

function cellRendererType({ rowData }) {
  if (!rowData.issue_type) return null
  return h('span', {
    class: `beads-badge beads-badge--${rowData.issue_type}`,
  }, typeLabel(rowData.issue_type))
}

function cellRendererPriority({ rowData }) {
  if (rowData.priority == null) return null
  return h('span', {
    class: `beads-badge beads-badge--p${rowData.priority}`,
  }, formatPriority(rowData.priority))
}

function cellRendererAssignee({ rowData }) {
  return h('span', { class: 'muted-text' }, rowData.assignee || '-')
}

function cellRendererTime({ cellData }) {
  return h('span', { class: 'muted-text' }, formatTime(cellData))
}

function cellRendererActions({ rowData }) {
  return h('div', { onClick: (e) => e.stopPropagation() }, [
    h(ElDropdown, { trigger: 'click' }, {
      default: () => h('button', { class: 'action-btn' }, `${t('issue.statusAction')} ▾`),
      dropdown: () => h(ElDropdownMenu, null, {
        default: () => [
          h(ElDropdownItem, { onClick: () => issueStore.updateStatus(rowData.id, 'open') }, () => t('status.open')),
          h(ElDropdownItem, { onClick: () => issueStore.updateStatus(rowData.id, 'in_progress') }, () => t('status.inProgress')),
          h(ElDropdownItem, { onClick: () => issueStore.updateStatus(rowData.id, 'closed') }, () => t('status.closed')),
        ],
      }),
    }),
  ])
}

const columns = computed(() => [
  {
    key: 'id',
    dataKey: 'id',
    title: t('issue.id'),
    width: 300,
    cellRenderer: cellRendererId,
  },
  {
    key: 'title',
    dataKey: 'title',
    title: t('issue.title'),
    width: 300,
    flexGrow: 1,
    cellRenderer: cellRendererTitle,
  },
  {
    key: 'status',
    dataKey: 'status',
    title: t('issue.status'),
    width: 110,
    cellRenderer: cellRendererStatus,
  },
  {
    key: 'issue_type',
    dataKey: 'issue_type',
    title: t('issue.type'),
    width: 100,
    cellRenderer: cellRendererType,
  },
  {
    key: 'priority',
    dataKey: 'priority',
    title: t('issue.priority'),
    width: 80,
    cellRenderer: cellRendererPriority,
  },
  {
    key: 'assignee',
    dataKey: 'assignee',
    title: t('issue.assignee'),
    width: 100,
    cellRenderer: cellRendererAssignee,
  },
  {
    key: 'updated_at',
    dataKey: 'updated_at',
    title: t('issue.updatedAt'),
    width: 170,
    cellRenderer: cellRendererTime,
  },
  {
    key: 'actions',
    dataKey: 'actions',
    title: t('issue.actions'),
    width: 160,
    fixed: 'right',
    cellRenderer: cellRendererActions,
  },
])
</script>

<style scoped>
.issues-view {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.issues-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
  flex-wrap: wrap;
  flex-shrink: 0;
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

.issues-table-wrapper {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.table-loading-overlay {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 8px;
  color: var(--muted);
  font-size: 13px;
}

.table-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: var(--muted);
  font-size: 14px;
}

.muted-text {
  color: var(--muted);
  font-size: 12px;
}

.cell-overflow {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.issues-table-wrapper :deep(.el-virtual-table) {
  --el-table-border-color: var(--border);
}

.issues-table-wrapper :deep(.el-table-v2__row-cell) {
  cursor: pointer;
}

.issues-table-wrapper :deep(.el-table-v2__header-cell) {
  font-size: 12px;
  font-weight: 600;
}
</style>

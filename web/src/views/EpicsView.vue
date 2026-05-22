<template>
  <div class="epics-view">
    <div class="epics-toolbar">
      <h3 style="margin: 0;">{{ t('epics.title') }}</h3>
      <el-button @click="issueStore.fetchIssues()" :loading="issueStore.loading">{{ t('issue.refresh') }}</el-button>
    </div>

    <el-table :data="epics" stripe v-loading="issueStore.loading" style="width: 100%">
      <el-table-column prop="id" :label="t('issue.id')" width="100" />
      <el-table-column prop="title" :label="t('epics.name')" min-width="250" show-overflow-tooltip />
      <el-table-column :label="t('issue.status')" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column :label="t('epics.progress')" width="250">
        <template #default="{ row }">
          <el-progress
            :percentage="epicProgress(row)"
            :status="epicProgress(row) >= 100 ? 'success' : ''"
            :stroke-width="16"
            :text-inside="true"
          />
        </template>
      </el-table-column>
      <el-table-column :label="t('epics.children')" width="120">
        <template #default="{ row }">
          {{ row.closed_children || 0 }} / {{ row.total_children || 0 }}
        </template>
      </el-table-column>
      <el-table-column :label="t('issue.updatedAt')" width="170">
        <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
      </el-table-column>
    </el-table>

    <el-empty v-if="!issueStore.loading && epics.length === 0" :description="t('epics.empty')" />
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

function statusTagType(status) {
  return { open: 'success', in_progress: 'warning', closed: 'info' }[status] || ''
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
  margin-bottom: 16px;
}
</style>

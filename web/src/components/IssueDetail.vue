<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('detail.title')"
    width="640px"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template v-if="issue">
      <div class="detail-header">
        <div class="detail-id">
          <el-tag size="small" effect="plain">{{ issue.id }}</el-tag>
          <el-tag :type="statusTagType(issue.status)" size="small">{{ statusLabel(issue.status) }}</el-tag>
          <el-tag v-if="issue.issue_type" :type="typeTagType(issue.issue_type)" size="small">{{ typeLabel(issue.issue_type) }}</el-tag>
          <el-tag v-if="issue.priority != null" :type="priorityTagType(issue.priority)" size="small">{{ formatPriority(issue.priority) }}</el-tag>
        </div>
      </div>

      <h3 class="detail-title">{{ issue.title }}</h3>

      <div v-if="issue.body" class="detail-body">
        <p style="white-space: pre-wrap;">{{ issue.body }}</p>
      </div>

      <el-divider />

      <el-descriptions :column="2" size="small" border>
        <el-descriptions-item :label="t('detail.statusLabel')">
          <el-select :model-value="issue.status" size="small" @change="onStatusChange">
            <el-option :label="t('status.open')" value="open" />
            <el-option :label="t('status.inProgress')" value="in_progress" />
            <el-option :label="t('status.closed')" value="closed" />
          </el-select>
        </el-descriptions-item>
        <el-descriptions-item :label="t('detail.priorityLabel')">
          <el-select :model-value="issue.priority !== undefined && issue.priority !== null ? String(issue.priority) : ''" size="small" clearable @change="onPriorityChange">
            <el-option label="P0" value="0" />
            <el-option label="P1" value="1" />
            <el-option label="P2" value="2" />
            <el-option label="P3" value="3" />
            <el-option label="P4" value="4" />
          </el-select>
        </el-descriptions-item>
        <el-descriptions-item :label="t('detail.createdAt')">{{ formatTime(issue.created_at) }}</el-descriptions-item>
        <el-descriptions-item :label="t('detail.updatedAt')">{{ formatTime(issue.updated_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="issue.closed_at" :label="t('detail.closedAt')">{{ formatTime(issue.closed_at) }}</el-descriptions-item>
        <el-descriptions-item v-if="issue.labels && issue.labels.length" :label="t('detail.labels')">
          <el-tag v-for="label in issue.labels" :key="label" size="small" closable @close="onRemoveLabel(label)">{{ label }}</el-tag>
        </el-descriptions-item>
      </el-descriptions>

      <div v-if="issue.dep_ids && issue.dep_ids.length" style="margin-top: 12px;">
        <strong>{{ t('detail.deps') }}:</strong>
        <el-tag v-for="dep in issue.dep_ids" :key="dep" size="small" style="margin-left: 4px;">{{ dep }}</el-tag>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: Boolean,
  issueId: String,
})
defineEmits(['update:modelValue'])

const { t } = useI18n()
const issueStore = useIssueStore()

const issue = computed(() => {
  if (!props.issueId) return null
  return issueStore.issues.find(i => i.id === props.issueId)
})

function statusTagType(status) {
  const map = { open: 'success', in_progress: 'warning', closed: 'info' }
  return map[status] || ''
}

function statusLabel(status) {
  const map = { open: t('status.open'), in_progress: t('status.inProgress'), closed: t('status.closed') }
  return map[status] || status
}

function typeTagType(type) {
  const map = { bug: 'danger', feature: 'success', task: 'warning', epic: '', chore: 'info' }
  return map[type] || ''
}

function typeLabel(type) {
  const map = { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }
  return map[type] || type
}

function priorityTagType(priority) {
  const map = { 0: 'danger', 1: 'warning', 2: '', 3: 'success', 4: 'info' }
  return map[priority] || ''
}

function formatPriority(priority) {
  return 'P' + priority
}

function formatTime(ts) {
  if (!ts) return '-'
  const d = new Date(typeof ts === 'number' ? ts : ts)
  return d.toLocaleString()
}

async function onStatusChange(status) {
  if (issue.value) {
    await issueStore.updateStatus(issue.value.id, status)
    ElMessage.success(t('detail.statusUpdated'))
  }
}

async function onPriorityChange(priority) {
  if (issue.value && priority) {
    await issueStore.updatePriority(issue.value.id, priority)
    ElMessage.success(t('detail.priorityUpdated'))
  }
}

async function onRemoveLabel(label) {
  if (issue.value) {
    await issueStore.removeLabel(issue.value.id, label)
    ElMessage.success(t('detail.labelRemoved'))
  }
}
</script>

<style scoped>
.detail-header {
  margin-bottom: 8px;
}
.detail-id {
  display: flex;
  gap: 6px;
  align-items: center;
}
.detail-title {
  margin: 0 0 12px;
  font-size: 18px;
}
.detail-body {
  color: var(--el-text-color-regular);
  font-size: 14px;
  line-height: 1.6;
}
</style>

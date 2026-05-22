<template>
  <el-dialog
    :model-value="modelValue"
    width="640px"
    @update:model-value="$emit('update:modelValue', $event)"
    class="detail-dialog"
  >
    <template #header>
      <div class="detail-dialog-header">
        <span class="detail-dialog-title">{{ t('detail.title') }}</span>
      </div>
    </template>
    <template v-if="issue">
      <div class="detail-badges">
        <span class="beads-badge beads-badge--id">{{ issue.id }}</span>
        <span class="beads-badge" :class="statusBadgeClass(issue.status)">{{ statusLabel(issue.status) }}</span>
        <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
        <span v-if="issue.priority != null" class="beads-badge" :class="'beads-badge--p' + issue.priority">{{ formatPriority(issue.priority) }}</span>
      </div>

      <h3 class="detail-title">{{ issue.title }}</h3>

      <div v-if="issue.description" class="detail-body">
        <p>{{ issue.description }}</p>
      </div>

      <div class="detail-divider"></div>

      <div class="detail-fields">
        <div class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.statusLabel') }}</label>
            <el-select :model-value="issue.status" size="small" @change="onStatusChange">
              <el-option :label="t('status.open')" value="open" />
              <el-option :label="t('status.inProgress')" value="in_progress" />
              <el-option :label="t('status.closed')" value="closed" />
            </el-select>
          </div>
          <div class="field-item">
            <label class="field-label">{{ t('detail.priorityLabel') }}</label>
            <el-select :model-value="issue.priority !== undefined && issue.priority !== null ? String(issue.priority) : ''" size="small" clearable @change="onPriorityChange">
              <el-option label="P0" value="0" />
              <el-option label="P1" value="1" />
              <el-option label="P2" value="2" />
              <el-option label="P3" value="3" />
              <el-option label="P4" value="4" />
            </el-select>
          </div>
        </div>
        <div class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.createdAt') }}</label>
            <span class="field-value">{{ formatTime(issue.created_at) }}</span>
          </div>
          <div class="field-item">
            <label class="field-label">{{ t('detail.updatedAt') }}</label>
            <span class="field-value">{{ formatTime(issue.updated_at) }}</span>
          </div>
        </div>
        <div v-if="issue.closed_at" class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.closedAt') }}</label>
            <span class="field-value">{{ formatTime(issue.closed_at) }}</span>
          </div>
        </div>
        <div v-if="issue.labels && issue.labels.length" class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.labels') }}</label>
            <div class="field-tags">
              <span v-for="label in issue.labels" :key="label" class="label-tag">
                {{ label }}
                <button class="label-remove" @click="onRemoveLabel(label)">×</button>
              </span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="issue.dep_ids && issue.dep_ids.length" class="detail-deps">
        <label class="field-label">{{ t('detail.deps') }}</label>
        <div class="dep-tags">
          <span v-for="dep in issue.dep_ids" :key="dep" class="beads-badge beads-badge--id">{{ dep }}</span>
        </div>
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

function statusBadgeClass(status) {
  return {
    open: 'beads-badge--open',
    in_progress: 'beads-badge--in-progress',
    closed: 'beads-badge--closed',
  }[status] || ''
}

function statusLabel(status) {
  const map = { open: t('status.open'), in_progress: t('status.inProgress'), closed: t('status.closed') }
  return map[status] || status
}

function typeLabel(type) {
  const map = { bug: t('type.bug'), feature: t('type.feature'), task: t('type.task'), epic: t('type.epic'), chore: t('type.chore') }
  return map[type] || type
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
.detail-dialog-header {
  display: flex;
  align-items: center;
}

.detail-dialog-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--fg);
}

.detail-badges {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.detail-title {
  margin: 0 0 12px;
  font-size: 18px;
  font-weight: 600;
  color: var(--fg);
  line-height: 1.4;
}

.detail-body {
  color: var(--muted);
  font-size: 14px;
  line-height: 1.6;
}

.detail-body p {
  margin: 0;
  white-space: pre-wrap;
}

.detail-divider {
  height: 1px;
  background: var(--border);
  margin: 16px 0;
}

.detail-fields {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.field-row {
  display: flex;
  gap: 24px;
}

.field-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  flex: 1;
}

.field-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.field-value {
  font-size: 13px;
  color: var(--fg);
}

.field-tags {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.label-tag {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px 8px;
  border-radius: var(--badge-radius, 999px);
  font-size: 11px;
  font-weight: 600;
  background: color-mix(in srgb, var(--panel-bg) 85%, #6b7280);
  color: color-mix(in srgb, #6b7280 85%, #000);
  border: 1px solid color-mix(in srgb, currentColor 30%, transparent);
}

.label-remove {
  background: transparent;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 0;
  font-size: 14px;
  line-height: 1;
  opacity: 0.6;
  transition: opacity 140ms ease;
}

.label-remove:hover {
  opacity: 1;
}

.detail-deps {
  margin-top: 16px;
}

.dep-tags {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}
</style>

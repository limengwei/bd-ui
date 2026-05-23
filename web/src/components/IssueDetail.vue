<template>
  <el-drawer
    :model-value="modelValue"
    direction="rtl"
    size="560px"
    :show-close="false"
    @update:model-value="$emit('update:modelValue', $event)"
    class="detail-drawer"
  >
    <template #header>
      <div class="detail-drawer-header">
        <span class="detail-drawer-title">{{ t('detail.title') }}</span>
        <button class="drawer-close-btn" @click="$emit('update:modelValue', false)">×</button>
      </div>
    </template>

    <template v-if="issue">
      <div class="detail-badges">
        <span class="beads-badge beads-badge--id">{{ issue.id }}</span>
        <el-select
          :model-value="issue.status"
          size="small"
          class="inline-status-select"
          @change="onStatusChange"
        >
          <el-option :label="t('status.open')" value="open" />
          <el-option :label="t('status.inProgress')" value="in_progress" />
          <el-option :label="t('status.closed')" value="closed" />
        </el-select>
        <span v-if="issue.issue_type" class="beads-badge" :class="'beads-badge--' + issue.issue_type">{{ typeLabel(issue.issue_type) }}</span>
        <el-select
          :model-value="issue.priority !== undefined && issue.priority !== null ? String(issue.priority) : ''"
          size="small"
          class="inline-priority-select"
          clearable
          :placeholder="t('detail.priorityLabel')"
          @change="onPriorityChange"
        >
          <el-option label="P0" value="0" />
          <el-option label="P1" value="1" />
          <el-option label="P2" value="2" />
          <el-option label="P3" value="3" />
          <el-option label="P4" value="4" />
        </el-select>
      </div>

      <div class="detail-title-row">
        <h3 v-if="!editingTitle" class="detail-title" @dblclick="startEditTitle">{{ issue.title }}</h3>
        <el-input
          v-else
          v-model="editTitleValue"
          size="large"
          class="detail-title-input"
          @blur="saveTitle"
          @keyup.enter="saveTitle"
          @keyup.escape="cancelEditTitle"
        />
        <button v-if="!editingTitle" class="edit-hint-btn" @click="startEditTitle" :title="t('detail.clickToEdit')">
          <el-icon><Edit /></el-icon>
        </button>
      </div>

      <div class="detail-section">
        <div class="section-header">
          <span class="section-label">{{ t('detail.description') }}</span>
          <button class="section-action-btn" @click="toggleDescriptionEdit">
            {{ editingDesc ? t('detail.preview') : t('detail.edit') }}
          </button>
        </div>
        <div v-if="!editingDesc" class="detail-description" @dblclick="toggleDescriptionEdit">
          <div v-if="issue.description" class="markdown-body" v-html="renderedDescription"></div>
          <div v-else class="empty-description" @click="toggleDescriptionEdit">{{ t('detail.noDescription') }}</div>
        </div>
        <div v-else class="detail-description-edit">
          <el-input
            v-model="editDescValue"
            type="textarea"
            :autosize="{ minRows: 6, maxRows: 20 }"
            :placeholder="t('detail.descPlaceholder')"
          />
          <div class="desc-edit-actions">
            <el-button size="small" @click="cancelEditDesc">{{ t('detail.cancel') }}</el-button>
            <el-button size="small" type="primary" @click="saveDesc">{{ t('detail.save') }}</el-button>
          </div>
        </div>
      </div>

      <div class="detail-divider"></div>

      <div class="detail-fields">
        <div class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.assignee') }}</label>
            <div class="field-value-row">
              <span class="field-value">{{ issue.assignee || t('detail.unassigned') }}</span>
              <button class="field-edit-btn" @click="showAssigneeInput = !showAssigneeInput">
                <el-icon><Edit /></el-icon>
              </button>
            </div>
            <div v-if="showAssigneeInput" class="inline-edit-row">
              <el-input v-model="assigneeValue" size="small" :placeholder="t('detail.assigneePlaceholder')" @keyup.enter="saveAssignee" />
              <el-button size="small" type="primary" @click="saveAssignee">{{ t('detail.save') }}</el-button>
              <el-button size="small" @click="showAssigneeInput = false">{{ t('detail.cancel') }}</el-button>
            </div>
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

        <div class="field-row">
          <div class="field-item">
            <label class="field-label">{{ t('detail.owner') }}</label>
            <span class="field-value">{{ issue.created_by || issue.owner || '-' }}</span>
          </div>
        </div>
      </div>

      <div class="detail-section">
        <div class="section-header">
          <span class="section-label">{{ t('detail.labels') }}</span>
          <button class="section-action-btn" @click="showLabelInput = !showLabelInput">+ {{ t('detail.addLabel') }}</button>
        </div>
        <div class="label-list">
          <span v-for="label in (issue.labels || [])" :key="label" class="label-tag">
            {{ label }}
            <button class="label-remove" @click="onRemoveLabel(label)">×</button>
          </span>
          <span v-if="!issue.labels || issue.labels.length === 0" class="field-value muted">{{ t('detail.noLabels') }}</span>
        </div>
        <div v-if="showLabelInput" class="inline-edit-row">
          <el-input v-model="newLabel" size="small" :placeholder="t('detail.labelPlaceholder')" @keyup.enter="onAddLabel" />
          <el-button size="small" type="primary" @click="onAddLabel">{{ t('detail.add') }}</el-button>
        </div>
      </div>

      <div class="detail-section">
        <div class="section-header">
          <span class="section-label">{{ t('detail.deps') }}</span>
          <button class="section-action-btn" @click="showDepInput = !showDepInput">+ {{ t('detail.addDep') }}</button>
        </div>
        <div class="dep-list">
          <span v-for="dep in (issue.dep_ids || [])" :key="dep" class="dep-tag">
            <span class="dep-id-link" @click="navigateToIssue(dep)">{{ dep }}</span>
            <button class="label-remove" @click="onRemoveDep(dep)">×</button>
          </span>
          <span v-if="!issue.dep_ids || issue.dep_ids.length === 0" class="field-value muted">{{ t('detail.noDeps') }}</span>
        </div>
        <div v-if="showDepInput" class="inline-edit-row">
          <el-input v-model="newDepId" size="small" :placeholder="t('detail.depPlaceholder')" @keyup.enter="onAddDep" />
          <el-button size="small" type="primary" @click="onAddDep">{{ t('detail.add') }}</el-button>
        </div>
      </div>

      <div v-if="children.length > 0" class="detail-section">
        <div class="section-header">
          <span class="section-label">{{ t('detail.children') }} ({{ children.length }})</span>
        </div>
        <div class="children-list">
          <div v-for="child in children" :key="child.id" class="child-item" @click="navigateToIssue(child.id)">
            <span class="beads-badge beads-badge--id">{{ child.id }}</span>
            <span class="child-title">{{ child.title }}</span>
            <span class="beads-badge" :class="statusBadgeClass(child.status)">{{ statusLabel(child.status) }}</span>
          </div>
        </div>
      </div>

      <div class="detail-section">
        <div class="section-header">
          <span class="section-label">{{ t('detail.comments') }} ({{ comments.length }})</span>
        </div>
        <div v-if="commentsLoading" class="comments-loading">{{ t('detail.loading') }}</div>
        <div v-else-if="comments.length === 0" class="empty-comments">{{ t('detail.noComments') }}</div>
        <div v-else class="comments-list">
          <div v-for="comment in comments" :key="comment.id || comment.created_at" class="comment-item">
            <div class="comment-header">
              <span class="comment-author">{{ comment.author || comment.created_by || '-' }}</span>
              <span class="comment-time">{{ formatTime(comment.created_at) }}</span>
            </div>
            <div class="comment-body">{{ comment.body || comment.text || comment.content }}</div>
          </div>
        </div>
        <div class="comment-input-row">
          <el-input
            v-model="newComment"
            type="textarea"
            :autosize="{ minRows: 2, maxRows: 6 }"
            :placeholder="t('detail.commentPlaceholder')"
          />
          <el-button
            size="small"
            type="primary"
            :disabled="!newComment.trim()"
            @click="onAddComment"
            class="comment-submit-btn"
          >
            {{ t('detail.submitComment') }}
          </el-button>
        </div>
      </div>
    </template>
  </el-drawer>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import { useWs } from '../composables/useWs'
import { ElMessage } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import { marked } from 'marked'

const props = defineProps({
  modelValue: Boolean,
  issueId: String,
})
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const issueStore = useIssueStore()
const { send } = useWs()

const editingTitle = ref(false)
const editTitleValue = ref('')
const editingDesc = ref(false)
const editDescValue = ref('')
const showLabelInput = ref(false)
const newLabel = ref('')
const showDepInput = ref(false)
const newDepId = ref('')
const showAssigneeInput = ref(false)
const assigneeValue = ref('')
const comments = ref([])
const commentsLoading = ref(false)
const newComment = ref('')

const issue = computed(() => {
  if (!props.issueId) return null
  return issueStore.issues.find(i => i.id === props.issueId)
})

const children = computed(() => {
  if (!issue.value) return []
  return issueStore.issues.filter(i => {
    const parentIds = i.parent_ids || (i.parent_id ? [i.parent_id] : [])
    return parentIds.includes(issue.value.id)
  })
})

const renderedDescription = computed(() => {
  if (!issue.value || !issue.value.description) return ''
  try {
    return marked(issue.value.description, { breaks: true, gfm: true })
  } catch {
    return issue.value.description
  }
})

watch(() => props.modelValue, (val) => {
  if (val && props.issueId) {
    loadComments()
  }
})

watch(() => props.issueId, (val) => {
  if (val && props.modelValue) {
    loadComments()
  }
  editingTitle.value = false
  editingDesc.value = false
  showLabelInput.value = false
  showDepInput.value = false
  showAssigneeInput.value = false
  newComment.value = ''
})

async function loadComments() {
  if (!props.issueId) return
  commentsLoading.value = true
  try {
    const result = await send('get-comments', { id: props.issueId })
    if (Array.isArray(result)) {
      comments.value = result
    } else if (result && Array.isArray(result.comments)) {
      comments.value = result.comments
    } else {
      comments.value = []
    }
  } catch {
    comments.value = []
  } finally {
    commentsLoading.value = false
  }
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

function formatTime(ts) {
  if (!ts) return '-'
  const d = new Date(typeof ts === 'number' ? ts : ts)
  return d.toLocaleString()
}

function startEditTitle() {
  if (!issue.value) return
  editingTitle.value = true
  editTitleValue.value = issue.value.title
}

async function saveTitle() {
  if (!issue.value || !editTitleValue.value.trim()) {
    editingTitle.value = false
    return
  }
  if (editTitleValue.value === issue.value.title) {
    editingTitle.value = false
    return
  }
  await issueStore.editText(issue.value.id, 'title', editTitleValue.value.trim())
  editingTitle.value = false
  ElMessage.success(t('detail.titleUpdated'))
}

function cancelEditTitle() {
  editingTitle.value = false
}

function toggleDescriptionEdit() {
  if (!issue.value) return
  editingDesc.value = !editingDesc.value
  if (editingDesc.value) {
    editDescValue.value = issue.value.description || ''
  }
}

async function saveDesc() {
  if (!issue.value) return
  await issueStore.editText(issue.value.id, 'description', editDescValue.value)
  editingDesc.value = false
  ElMessage.success(t('detail.descUpdated'))
}

function cancelEditDesc() {
  editingDesc.value = false
}

async function onStatusChange(status) {
  if (issue.value) {
    await issueStore.updateStatus(issue.value.id, status)
    ElMessage.success(t('detail.statusUpdated'))
  }
}

async function onPriorityChange(priority) {
  if (issue.value && priority !== undefined && priority !== null && priority !== '') {
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

async function onAddLabel() {
  if (!issue.value || !newLabel.value.trim()) return
  await issueStore.addLabel(issue.value.id, newLabel.value.trim())
  newLabel.value = ''
  showLabelInput.value = false
  ElMessage.success(t('detail.labelAdded'))
}

async function onRemoveDep(depId) {
  if (issue.value) {
    await issueStore.removeDep(issue.value.id, depId)
    ElMessage.success(t('detail.depRemoved'))
  }
}

async function onAddDep() {
  if (!issue.value || !newDepId.value.trim()) return
  await issueStore.addDep(issue.value.id, newDepId.value.trim())
  newDepId.value = ''
  showDepInput.value = false
  ElMessage.success(t('detail.depAdded'))
}

async function saveAssignee() {
  if (!issue.value) return
  try {
    await send('update-assignee', { id: issue.value.id, assignee: assigneeValue.value.trim() })
    showAssigneeInput.value = false
    ElMessage.success(t('detail.assigneeUpdated'))
  } catch {
    ElMessage.error(t('detail.updateFail'))
  }
}

async function onAddComment() {
  if (!issue.value || !newComment.value.trim()) return
  try {
    await send('add-comment', { id: issue.value.id, body: newComment.value.trim() })
    newComment.value = ''
    await loadComments()
    ElMessage.success(t('detail.commentAdded'))
  } catch {
    ElMessage.error(t('detail.commentFail'))
  }
}

function navigateToIssue(id) {
  emit('update:modelValue', false)
  setTimeout(() => {
    issueStore.issues
    const found = issueStore.issues.find(i => i.id === id)
    if (found) {
      emit('update:modelValue', true)
      const parent = document.querySelector('.detail-drawer')
    }
  }, 100)
}
</script>

<style scoped>
.detail-drawer :deep(.el-drawer__header) {
  margin-bottom: 0;
  padding: 14px 20px;
  border-bottom: 1px solid var(--border);
}

.detail-drawer :deep(.el-drawer__body) {
  padding: 20px;
  overflow-y: auto;
}

.detail-drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.detail-drawer-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
}

.drawer-close-btn {
  background: transparent;
  border: none;
  color: var(--muted);
  font-size: 22px;
  cursor: pointer;
  padding: 0 4px;
  line-height: 1;
  transition: color 140ms ease;
}

.drawer-close-btn:hover {
  color: var(--fg);
}

.detail-badges {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.inline-status-select {
  width: 110px;
}

.inline-priority-select {
  width: 90px;
}

.detail-title-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 16px;
}

.detail-title {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--fg);
  line-height: 1.4;
  flex: 1;
  cursor: text;
  border-radius: 4px;
  padding: 2px 4px;
  margin-left: -4px;
  transition: background 140ms ease;
}

.detail-title:hover {
  background: color-mix(in srgb, var(--fg) 4%, transparent);
}

.detail-title-input {
  flex: 1;
}

.edit-hint-btn {
  background: transparent;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  font-size: 14px;
  opacity: 0;
  transition: opacity 140ms ease, color 140ms ease;
  flex-shrink: 0;
  margin-top: 4px;
}

.detail-title-row:hover .edit-hint-btn {
  opacity: 1;
}

.edit-hint-btn:hover {
  color: var(--link);
}

.detail-section {
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.section-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.section-action-btn {
  background: transparent;
  border: none;
  color: var(--link);
  font-size: 12px;
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: background 140ms ease;
}

.section-action-btn:hover {
  background: color-mix(in srgb, var(--link) 8%, transparent);
}

.detail-description {
  color: var(--fg);
  font-size: 14px;
  line-height: 1.6;
  cursor: text;
  border-radius: 6px;
  padding: 8px;
  border: 1px solid transparent;
  transition: border-color 140ms ease;
}

.detail-description:hover {
  border-color: var(--border);
}

.empty-description {
  color: var(--muted);
  font-style: italic;
  cursor: pointer;
  padding: 8px;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin: 16px 0 8px;
  font-weight: 600;
  color: var(--fg);
}

.markdown-body :deep(h1) { font-size: 18px; }
.markdown-body :deep(h2) { font-size: 16px; }
.markdown-body :deep(h3) { font-size: 14px; }

.markdown-body :deep(p) {
  margin: 0 0 8px;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 20px;
  margin: 0 0 8px;
}

.markdown-body :deep(code) {
  background: var(--code-bg);
  color: var(--code-fg);
  padding: 1px 5px;
  border-radius: 3px;
  font-size: 13px;
  font-family: 'SF Mono', 'Fira Code', 'Consolas', monospace;
}

.markdown-body :deep(pre) {
  background: var(--pre-bg);
  color: var(--pre-fg);
  padding: 12px;
  border-radius: 6px;
  overflow-x: auto;
  margin: 0 0 8px;
}

.markdown-body :deep(pre code) {
  background: transparent;
  color: inherit;
  padding: 0;
}

.markdown-body :deep(blockquote) {
  border-left: 3px solid var(--border);
  margin: 0 0 8px;
  padding: 4px 12px;
  color: var(--muted);
}

.markdown-body :deep(a) {
  color: var(--link);
  text-decoration: none;
}

.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

.desc-edit-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.detail-divider {
  height: 1px;
  background: var(--border);
  margin: 16px 0;
}

.detail-fields {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 16px;
}

.field-row {
  display: flex;
  gap: 20px;
}

.field-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
  flex: 1;
}

.field-label {
  font-size: 11px;
  font-weight: 700;
  color: var(--muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.field-value {
  font-size: 13px;
  color: var(--fg);
}

.field-value.muted {
  color: var(--muted);
  font-style: italic;
}

.field-value-row {
  display: flex;
  align-items: center;
  gap: 6px;
}

.field-edit-btn {
  background: transparent;
  border: none;
  color: var(--muted);
  cursor: pointer;
  padding: 2px;
  font-size: 12px;
  opacity: 0;
  transition: opacity 140ms ease, color 140ms ease;
}

.field-item:hover .field-edit-btn {
  opacity: 1;
}

.field-edit-btn:hover {
  color: var(--link);
}

.inline-edit-row {
  display: flex;
  gap: 6px;
  margin-top: 4px;
  align-items: center;
}

.label-list {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  min-height: 24px;
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
  opacity: 0.5;
  transition: opacity 140ms ease;
}

.label-remove:hover {
  opacity: 1;
  color: #ef4444;
}

.dep-list {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
  min-height: 24px;
}

.dep-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: var(--badge-radius, 999px);
  font-size: 11px;
  font-weight: 600;
  background: color-mix(in srgb, var(--panel-bg) 85%, #6b7280);
  color: color-mix(in srgb, #6b7280 85%, #000);
  border: 1px solid color-mix(in srgb, currentColor 30%, transparent);
}

.dep-id-link {
  cursor: pointer;
  color: var(--link);
  transition: color 140ms ease;
}

.dep-id-link:hover {
  text-decoration: underline;
}

.children-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.child-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  background: var(--panel-bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  cursor: pointer;
  transition: border-color 140ms ease;
}

.child-item:hover {
  border-color: color-mix(in srgb, var(--link) 40%, var(--border));
}

.child-title {
  flex: 1;
  font-size: 13px;
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.comments-loading,
.empty-comments {
  color: var(--muted);
  font-size: 13px;
  padding: 8px 0;
  font-style: italic;
}

.comments-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-bottom: 12px;
}

.comment-item {
  padding: 10px;
  background: var(--panel-bg);
  border: 1px solid var(--border);
  border-radius: 8px;
}

.comment-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
}

.comment-author {
  font-size: 12px;
  font-weight: 600;
  color: var(--fg);
}

.comment-time {
  font-size: 11px;
  color: var(--muted);
}

.comment-body {
  font-size: 13px;
  color: var(--fg);
  line-height: 1.5;
  white-space: pre-wrap;
}

.comment-input-row {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.comment-submit-btn {
  align-self: flex-end;
}
</style>

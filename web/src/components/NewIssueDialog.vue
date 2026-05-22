<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('newIssue.title')"
    width="520px"
    @update:model-value="emit('update:modelValue', $event)"
    @close="resetForm"
  >
    <el-form :model="form" label-width="80px" label-position="top">
      <el-form-item :label="t('newIssue.titleLabel')" required>
        <el-input v-model="form.title" :placeholder="t('newIssue.titlePlaceholder')" />
      </el-form-item>
      <el-form-item :label="t('newIssue.bodyLabel')">
        <el-input
          v-model="form.body"
          type="textarea"
          :rows="4"
          :placeholder="t('newIssue.bodyPlaceholder')"
        />
      </el-form-item>
      <el-form-item :label="t('newIssue.typeLabel')">
        <el-select v-model="form.issue_type" :placeholder="t('newIssue.typePlaceholder')" clearable>
          <el-option :label="t('type.bug')" value="bug" />
          <el-option :label="t('type.feature')" value="feature" />
          <el-option :label="t('type.task')" value="task" />
          <el-option :label="t('type.epic')" value="epic" />
          <el-option :label="t('type.chore')" value="chore" />
        </el-select>
      </el-form-item>
      <el-form-item :label="t('newIssue.priorityLabel')">
        <el-select v-model="form.priority" :placeholder="t('newIssue.priorityPlaceholder')" clearable>
          <el-option label="P0" value="0" />
          <el-option label="P1" value="1" />
          <el-option label="P2" value="2" />
          <el-option label="P3" value="3" />
          <el-option label="P4" value="4" />
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">{{ t('newIssue.cancel') }}</el-button>
      <el-button type="primary" :disabled="!form.title" @click="submit">{{ t('newIssue.create') }}</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import { ElMessage } from 'element-plus'

const props = defineProps({ modelValue: Boolean })
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const issueStore = useIssueStore()

const form = reactive({
  title: '',
  body: '',
  issue_type: '',
  priority: '',
})

function resetForm() {
  form.title = ''
  form.body = ''
  form.issue_type = ''
  form.priority = ''
}

async function submit() {
  if (!form.title) return
  try {
    await issueStore.createIssue(form.title, form.body, form.issue_type, form.priority)
    ElMessage.success(t('newIssue.success'))
    resetForm()
    emit('update:modelValue', false)
  } catch (e) {
    ElMessage.error(t('newIssue.fail') + ': ' + (e.message || ''))
  }
}
</script>

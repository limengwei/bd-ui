<template>
  <el-dialog
    :model-value="modelValue"
    :title="t('settings.title')"
    width="480px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-position="top">
      <el-form-item :label="t('settings.bdBinLabel')">
        <el-input
          v-model="bdBinPath"
          :placeholder="t('settings.bdBinPlaceholder')"
          clearable
        >
          <template #append>
            <el-button @click="save">{{ t('settings.save') }}</el-button>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item v-if="bdVersion" :label="t('settings.versionLabel')">
        <el-tag type="success" size="small">{{ bdVersion }}</el-tag>
      </el-form-item>
    </el-form>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useWs } from '../composables/useWs'
import { ElMessage } from 'element-plus'

const props = defineProps({ modelValue: Boolean })
const emit = defineEmits(['update:modelValue'])

const { t } = useI18n()
const { send } = useWs()

const bdBinPath = ref('')
const bdVersion = ref('')

watch(() => props.modelValue, async (val) => {
  if (val) {
    try {
      const result = await send('get-bd-bin')
      if (result) {
        bdBinPath.value = result.path || ''
        bdVersion.value = result.version || ''
      }
    } catch (e) {
      console.error('Failed to get bd bin info:', e)
    }
  }
})

async function save() {
  try {
    const result = await send('set-bd-bin', { path: bdBinPath.value })
    if (result) {
      bdVersion.value = result.version || ''
      ElMessage.success(t('settings.saved'))
    }
  } catch (e) {
    ElMessage.error(t('settings.saveFail') + ': ' + (e.message || ''))
  }
}
</script>

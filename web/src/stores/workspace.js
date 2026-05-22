import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useWs } from '../composables/useWs'

export const useWorkspaceStore = defineStore('workspace', () => {
  const { send } = useWs()

  const workspaces = ref([])
  const current = ref(null)
  const loading = ref(false)

  async function loadWorkspaces() {
    loading.value = true
    try {
      const result = await send('list-workspaces')
      if (result) {
        workspaces.value = result.workspaces || []
        if (result.current) {
          current.value = {
            path: result.current.root_dir,
            database: result.current.db_path,
          }
        }
      }
    } catch (e) {
      console.error('加载工作空间失败:', e)
    } finally {
      loading.value = false
    }
  }

  async function switchWorkspace(path) {
    try {
      const result = await send('set-workspace', { path })
      if (result && result.workspace) {
        current.value = {
          path: result.workspace.root_dir,
          database: result.workspace.db_path,
        }
      }
    } catch (e) {
      console.error('切换工作空间失败:', e)
    }
  }

  async function addWorkspace(path) {
    const result = await send('add-workspace', { path })
    if (result) {
      await loadWorkspaces()
    }
    return result
  }

  async function removeWorkspace(path) {
    await send('remove-workspace', { path })
    if (current.value && current.value.path === path) {
      current.value = null
    }
    await loadWorkspaces()
  }

  function workspaceName(path) {
    if (!path) return '未知'
    const parts = path.replace(/\\/g, '/').split('/').filter(Boolean)
    return parts[parts.length - 1] || '未知'
  }

  return {
    workspaces,
    current,
    loading,
    loadWorkspaces,
    switchWorkspace,
    addWorkspace,
    removeWorkspace,
    workspaceName,
  }
})

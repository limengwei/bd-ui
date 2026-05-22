import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWs } from '../composables/useWs'

export const useIssueStore = defineStore('issues', () => {
  const { send, on, off } = useWs()

  const issues = ref([])
  const loading = ref(false)
  const filterStatus = ref('all')
  const filterType = ref('')
  const searchText = ref('')

  const filteredIssues = computed(() => {
    let list = issues.value
    if (filterStatus.value !== 'all') {
      if (filterStatus.value === 'ready') {
        list = list.filter(i => i.status === 'open' && (!i.dep_ids || i.dep_ids.length === 0))
      } else {
        list = list.filter(i => i.status === filterStatus.value)
      }
    }
    if (filterType.value) {
      list = list.filter(i => i.issue_type === filterType.value)
    }
    if (searchText.value) {
      const q = searchText.value.toLowerCase()
      list = list.filter(i =>
        (i.title || '').toLowerCase().includes(q) ||
        (i.id || '').toLowerCase().includes(q)
      )
    }
    return list
  })

  const openIssues = computed(() => issues.value.filter(i => i.status === 'open'))
  const inProgressIssues = computed(() => issues.value.filter(i => i.status === 'in_progress'))
  const closedIssues = computed(() => issues.value.filter(i => i.status === 'closed'))
  const readyIssues = computed(() => issues.value.filter(i => i.status === 'open' && (!i.dep_ids || i.dep_ids.length === 0)))
  const blockedIssues = computed(() => issues.value.filter(i => i.status === 'open' && i.dep_ids && i.dep_ids.length > 0))

  async function fetchIssues() {
    loading.value = true
    try {
      const result = await send('list-issues')
      if (Array.isArray(result)) {
        issues.value = result
      }
    } catch (e) {
      console.error('获取Issues列表失败:', e)
    } finally {
      loading.value = false
    }
  }

  async function updateStatus(id, status) {
    try {
      await send('update-status', { id, status })
      await fetchIssues()
    } catch (e) {
      console.error('更新状态失败:', e)
    }
  }

  async function editText(id, field, value) {
    try {
      await send('edit-text', { id, field, value })
      await fetchIssues()
    } catch (e) {
      console.error('编辑失败:', e)
    }
  }

  async function updatePriority(id, priority) {
    try {
      await send('update-priority', { id, priority })
      await fetchIssues()
    } catch (e) {
      console.error('更新优先级失败:', e)
    }
  }

  async function createIssue(title, body, issueType, priority) {
    try {
      await send('create-issue', { title, body, issue_type: issueType, priority })
      await fetchIssues()
    } catch (e) {
      console.error('创建Issues失败:', e)
    }
  }

  async function deleteIssue(id) {
    try {
      await send('delete-issue', { id })
      await fetchIssues()
    } catch (e) {
      console.error('删除Issues失败:', e)
    }
  }

  async function addDep(id, depId) {
    try {
      await send('dep-add', { id, dep_id: depId })
      await fetchIssues()
    } catch (e) {
      console.error('添加依赖失败:', e)
    }
  }

  async function removeDep(id, depId) {
    try {
      await send('dep-remove', { id, dep_id: depId })
      await fetchIssues()
    } catch (e) {
      console.error('移除依赖失败:', e)
    }
  }

  async function addLabel(id, label) {
    try {
      await send('label-add', { id, label })
      await fetchIssues()
    } catch (e) {
      console.error('添加标签失败:', e)
    }
  }

  async function removeLabel(id, label) {
    try {
      await send('label-remove', { id, label })
      await fetchIssues()
    } catch (e) {
      console.error('移除标签失败:', e)
    }
  }

  return {
    issues,
    loading,
    filterStatus,
    filterType,
    searchText,
    filteredIssues,
    openIssues,
    inProgressIssues,
    closedIssues,
    readyIssues,
    blockedIssues,
    fetchIssues,
    updateStatus,
    editText,
    updatePriority,
    createIssue,
    deleteIssue,
    addDep,
    removeDep,
    addLabel,
    removeLabel,
  }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWs } from '../composables/useWs'

const SUB_ID = 'main-issues'

export const useIssueStore = defineStore('issues', () => {
  const { send, on, off } = useWs()

  const issues = ref([])
  const loading = ref(false)
  const filterStatus = ref('all')
  const filterType = ref('')
  const filterAssignee = ref('')
  const filterLabel = ref('')
  const searchText = ref('')
  const sortBy = ref('updated_at')
  const sortOrder = ref('desc')

  const allAssignees = computed(() => {
    const set = new Set()
    issues.value.forEach(i => {
      if (i.assignee) set.add(i.assignee)
    })
    return Array.from(set).sort()
  })

  const allLabels = computed(() => {
    const set = new Set()
    issues.value.forEach(i => {
      if (i.labels) i.labels.forEach(l => set.add(l))
    })
    return Array.from(set).sort()
  })

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
    if (filterAssignee.value) {
      if (filterAssignee.value === '__none__') {
        list = list.filter(i => !i.assignee)
      } else {
        list = list.filter(i => i.assignee === filterAssignee.value)
      }
    }
    if (filterLabel.value) {
      list = list.filter(i => i.labels && i.labels.includes(filterLabel.value))
    }
    if (searchText.value) {
      const q = searchText.value.toLowerCase()
      list = list.filter(i =>
        (i.title || '').toLowerCase().includes(q) ||
        (i.id || '').toLowerCase().includes(q) ||
        (i.description || '').toLowerCase().includes(q)
      )
    }

    list = [...list].sort((a, b) => {
      let va, vb
      switch (sortBy.value) {
        case 'priority':
          va = a.priority != null ? a.priority : 999
          vb = b.priority != null ? b.priority : 999
          break
        case 'created_at':
          va = a.created_at || 0
          vb = b.created_at || 0
          break
        case 'updated_at':
        default:
          va = a.updated_at || 0
          vb = b.updated_at || 0
          break
      }
      if (typeof va === 'string') va = new Date(va).getTime()
      if (typeof vb === 'string') vb = new Date(vb).getTime()
      return sortOrder.value === 'desc' ? vb - va : va - vb
    })

    return list
  })

  const openIssues = computed(() => issues.value.filter(i => i.status === 'open'))
  const inProgressIssues = computed(() => issues.value.filter(i => i.status === 'in_progress'))
  const closedIssues = computed(() => issues.value.filter(i => i.status === 'closed'))
  const readyIssues = computed(() => issues.value.filter(i => i.status === 'open' && (!i.dep_ids || i.dep_ids.length === 0)))
  const blockedIssues = computed(() => issues.value.filter(i => i.status === 'open' && i.dep_ids && i.dep_ids.length > 0))

  function handleSnapshot(payload) {
    if (payload && payload.issues) {
      issues.value = payload.issues
    }
    loading.value = false
  }

  function handleUpsert(payload) {
    if (!payload || !payload.issue) return
    const updated = payload.issue
    const idx = issues.value.findIndex(i => i.id === updated.id)
    if (idx >= 0) {
      issues.value[idx] = updated
    } else {
      issues.value.push(updated)
    }
  }

  function handleDelete(payload) {
    if (!payload || !payload.issue_id) return
    issues.value = issues.value.filter(i => i.id !== payload.issue_id)
  }

  function clearFilters() {
    filterStatus.value = 'all'
    filterType.value = ''
    filterAssignee.value = ''
    filterLabel.value = ''
    searchText.value = ''
  }

  async function fetchIssues() {
    loading.value = true

    off('snapshot')
    off('upsert')
    off('delete')

    on('snapshot', handleSnapshot)
    on('upsert', handleUpsert)
    on('delete', handleDelete)

    try {
      await send('subscribe-list', {
        id: SUB_ID,
        spec: { type: 'all-issues' },
      })
    } catch (e) {
      console.error('订阅Issues失败:', e)
      loading.value = false
    }
  }

  async function updateStatus(id, status) {
    try {
      await send('update-status', { id, status })
    } catch (e) {
      console.error('更新状态失败:', e)
    }
  }

  async function editText(id, field, value) {
    try {
      await send('edit-text', { id, field, value })
    } catch (e) {
      console.error('编辑失败:', e)
    }
  }

  async function updatePriority(id, priority) {
    try {
      await send('update-priority', { id, priority })
    } catch (e) {
      console.error('更新优先级失败:', e)
    }
  }

  async function createIssue(title, body, issueType, priority, assignee, labels) {
    try {
      await send('create-issue', { title, body, issue_type: issueType, priority, assignee, labels })
    } catch (e) {
      console.error('创建Issues失败:', e)
    }
  }

  async function deleteIssue(id) {
    try {
      await send('delete-issue', { id })
    } catch (e) {
      console.error('删除Issues失败:', e)
    }
  }

  async function addDep(id, depId) {
    try {
      await send('dep-add', { id, dep_id: depId })
    } catch (e) {
      console.error('添加依赖失败:', e)
    }
  }

  async function removeDep(id, depId) {
    try {
      await send('dep-remove', { id, dep_id: depId })
    } catch (e) {
      console.error('移除依赖失败:', e)
    }
  }

  async function addLabel(id, label) {
    try {
      await send('label-add', { id, label })
    } catch (e) {
      console.error('添加标签失败:', e)
    }
  }

  async function removeLabel(id, label) {
    try {
      await send('label-remove', { id, label })
    } catch (e) {
      console.error('移除标签失败:', e)
    }
  }

  return {
    issues,
    loading,
    filterStatus,
    filterType,
    filterAssignee,
    filterLabel,
    searchText,
    sortBy,
    sortOrder,
    allAssignees,
    allLabels,
    filteredIssues,
    openIssues,
    inProgressIssues,
    closedIssues,
    readyIssues,
    blockedIssues,
    clearFilters,
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

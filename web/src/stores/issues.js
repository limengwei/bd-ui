import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useWs } from '../composables/useWs'

const SUB_ID = 'main-issues'

function parseTimestamp(v) {
  if (!v) return 0
  return typeof v === 'number' ? v : new Date(v).getTime()
}

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

  const issueIndexMap = computed(() => {
    const map = new Map()
    for (let i = 0; i < issues.value.length; i++) {
      map.set(issues.value[i].id, i)
    }
    return map
  })

  const groupedByStatus = computed(() => {
    const open = []
    const inProgress = []
    const closed = []
    const ready = []
    const blocked = []
    const assigneeSet = new Set()
    const labelSet = new Set()

    for (const issue of issues.value) {
      if (issue.assignee) assigneeSet.add(issue.assignee)
      if (issue.labels) {
        for (const l of issue.labels) labelSet.add(l)
      }

      switch (issue.status) {
        case 'open':
          open.push(issue)
          if (issue.dep_ids && issue.dep_ids.length > 0) {
            blocked.push(issue)
          } else {
            ready.push(issue)
          }
          break
        case 'in_progress':
          inProgress.push(issue)
          break
        case 'closed':
          closed.push(issue)
          break
      }
    }

    return { open, inProgress, closed, ready, blocked, assignees: Array.from(assigneeSet).sort(), labels: Array.from(labelSet).sort() }
  })

  const allAssignees = computed(() => groupedByStatus.value.assignees)
  const allLabels = computed(() => groupedByStatus.value.labels)

  const openIssues = computed(() => groupedByStatus.value.open)
  const inProgressIssues = computed(() => groupedByStatus.value.inProgress)
  const closedIssues = computed(() => groupedByStatus.value.closed)
  const readyIssues = computed(() => groupedByStatus.value.ready)
  const blockedIssues = computed(() => groupedByStatus.value.blocked)

  const parentChildMap = computed(() => {
    const map = new Map()
    for (const issue of issues.value) {
      if (issue.parent) {
        let children = map.get(issue.parent)
        if (!children) {
          children = []
          map.set(issue.parent, children)
        }
        children.push(issue)
      }
      if (issue.parent_ids) {
        for (const pid of issue.parent_ids) {
          let children = map.get(pid)
          if (!children) {
            children = []
            map.set(pid, children)
          }
          children.push(issue)
        }
      }
      if (issue.parent_id && !issue.parent_ids) {
        let children = map.get(issue.parent_id)
        if (!children) {
          children = []
          map.set(issue.parent_id, children)
        }
        children.push(issue)
      }
    }
    return map
  })

  const filteredIssues = computed(() => {
    let list = issues.value
    if (filterStatus.value !== 'all') {
      if (filterStatus.value === 'ready') {
        list = groupedByStatus.value.ready
      } else {
        const status = filterStatus.value
        list = list.filter(i => i.status === status)
      }
    }
    if (filterType.value) {
      const ft = filterType.value
      list = list.filter(i => i.issue_type === ft)
    }
    if (filterAssignee.value) {
      if (filterAssignee.value === '__none__') {
        list = list.filter(i => !i.assignee)
      } else {
        const fa = filterAssignee.value
        list = list.filter(i => i.assignee === fa)
      }
    }
    if (filterLabel.value) {
      const fl = filterLabel.value
      list = list.filter(i => i.labels && i.labels.includes(fl))
    }
    if (searchText.value) {
      const q = searchText.value.toLowerCase()
      list = list.filter(i =>
        (i.title || '').toLowerCase().includes(q) ||
        (i.id || '').toLowerCase().includes(q) ||
        (i.description || '').toLowerCase().includes(q)
      )
    }

    const sorted = [...list]
    const desc = sortOrder.value === 'desc'
    const field = sortBy.value
    sorted.sort((a, b) => {
      let va, vb
      switch (field) {
        case 'priority':
          va = a.priority != null ? a.priority : 999
          vb = b.priority != null ? b.priority : 999
          break
        case 'created_at':
          va = parseTimestamp(a.created_at)
          vb = parseTimestamp(b.created_at)
          break
        default:
          va = parseTimestamp(a.updated_at)
          vb = parseTimestamp(b.updated_at)
          break
      }
      return desc ? vb - va : va - vb
    })

    return sorted
  })

  function handleSnapshot(payload) {
    if (payload && payload.issues) {
      issues.value = payload.issues
    }
    loading.value = false
  }

  function handleUpsert(payload) {
    if (!payload || !payload.issue) return
    const updated = payload.issue
    const idx = issueIndexMap.value.get(updated.id)
    if (idx != null) {
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

  let fetchInFlight = false

  async function fetchIssues() {
    if (fetchInFlight) return
    fetchInFlight = true
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
    } finally {
      fetchInFlight = false
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
    issueIndexMap,
    parentChildMap,
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
    addDep,
    removeDep,
    addLabel,
    removeLabel,
  }
})

<template>
  <div class="graph-view">
    <div class="graph-toolbar">
      <span class="graph-title">{{ t('graph.title') }}</span>
      <span class="graph-hint">{{ t('graph.hint') }}</span>
      <button class="toolbar-btn" @click="issueStore.fetchIssues()" :disabled="issueStore.loading">
        {{ t('issue.refresh') }}
      </button>
    </div>

    <div v-if="issueStore.loading" class="graph-loading">{{ t('detail.loading') }}</div>

    <div v-else-if="issuesWithDeps.length === 0" class="empty-state">
      <div class="empty-icon">🔗</div>
      <div class="empty-text">{{ t('graph.empty') }}</div>
    </div>

    <div v-else class="graph-container">
      <div v-for="group in dependencyGroups" :key="group.root.id" class="dep-chain">
        <div class="dep-chain-header">
          <div
            class="dep-node dep-node--root"
            :class="'dep-node--' + group.root.status"
            @click="onNodeClick(group.root.id)"
          >
            <span class="node-id">{{ group.root.id }}</span>
            <span class="node-title">{{ group.root.title }}</span>
            <span class="beads-badge" :class="statusBadgeClass(group.root.status)">{{ statusLabel(group.root.status) }}</span>
          </div>
        </div>
        <div v-if="group.deps.length > 0" class="dep-chain-children">
          <div class="dep-connector">← {{ t('graph.dependsOn') }}</div>
          <div
            v-for="dep in group.deps"
            :key="dep.id"
            class="dep-node dep-node--child"
            :class="'dep-node--' + dep.status"
            @click="onNodeClick(dep.id)"
          >
            <span class="node-id">{{ dep.id }}</span>
            <span class="node-title">{{ dep.title }}</span>
            <span class="beads-badge" :class="statusBadgeClass(dep.status)">{{ statusLabel(dep.status) }}</span>
          </div>
        </div>
        <div v-if="group.dependents.length > 0" class="dep-chain-children">
          <div class="dep-connector dep-connector--reverse">→ {{ t('graph.blockedBy') }}</div>
          <div
            v-for="dep in group.dependents"
            :key="dep.id"
            class="dep-node dep-node--dependent"
            :class="'dep-node--' + dep.status"
            @click="onNodeClick(dep.id)"
          >
            <span class="node-id">{{ dep.id }}</span>
            <span class="node-title">{{ dep.title }}</span>
            <span class="beads-badge" :class="statusBadgeClass(dep.status)">{{ statusLabel(dep.status) }}</span>
          </div>
        </div>
      </div>
    </div>

    <IssueDetail v-model="showDetail" :issue-id="selectedIssueId" />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useIssueStore } from '../stores/issues'
import IssueDetail from '../components/IssueDetail.vue'

const { t } = useI18n()
const issueStore = useIssueStore()
const showDetail = ref(false)
const selectedIssueId = ref(null)

const issuesWithDeps = computed(() => {
  return issueStore.issues.filter(i => {
    return (i.dep_ids && i.dep_ids.length > 0) || i.status === 'open'
  })
})

const depIdToDependentsMap = computed(() => {
  const map = new Map()
  for (const issue of issueStore.issues) {
    if (issue.dep_ids) {
      for (const depId of issue.dep_ids) {
        let list = map.get(depId)
        if (!list) {
          list = []
          map.set(depId, list)
        }
        list.push(issue)
      }
    }
  }
  return map
})

const dependencyGroups = computed(() => {
  const blocked = issueStore.blockedIssues
  const indexMap = issueStore.issueIndexMap
  return blocked.map(issue => {
    const deps = (issue.dep_ids || []).map(depId => {
      const idx = indexMap.get(depId)
      return idx != null ? issueStore.issues[idx] : { id: depId, title: depId, status: 'unknown' }
    })
    const dependents = depIdToDependentsMap.value.get(issue.id) || []
    return { root: issue, deps, dependents }
  })
})

function onNodeClick(id) {
  selectedIssueId.value = id
  showDetail.value = true
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

onMounted(() => {
  if (issueStore.issues.length === 0) issueStore.fetchIssues()
})
</script>

<style scoped>
.graph-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.graph-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--fg);
}

.graph-hint {
  font-size: 12px;
  color: var(--muted);
  flex: 1;
}

.toolbar-btn {
  background: var(--panel-bg);
  color: var(--fg);
  border: 1px solid var(--border);
  padding: 5px 14px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}

.toolbar-btn:disabled {
  opacity: 0.6;
}

.graph-loading,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 48px 0;
  color: var(--muted);
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.empty-text {
  font-size: 14px;
}

.graph-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dep-chain {
  background: var(--panel-bg);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 12px 16px;
}

.dep-chain-header {
  margin-bottom: 8px;
}

.dep-chain-children {
  margin-left: 24px;
  margin-top: 4px;
}

.dep-connector {
  font-size: 11px;
  color: var(--muted);
  font-weight: 600;
  margin-bottom: 4px;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.dep-connector--reverse {
  color: #ef4444;
}

.dep-node {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
  transition: background 140ms ease, border-color 140ms ease;
  border: 1px solid transparent;
}

.dep-node:hover {
  background: color-mix(in srgb, var(--fg) 3%, transparent);
  border-color: color-mix(in srgb, var(--link) 30%, var(--border));
}

.dep-node--root {
  background: var(--bg);
  border: 1px solid var(--border);
}

.dep-node--child {
  background: color-mix(in srgb, var(--panel-bg) 50%, transparent);
}

.dep-node--dependent {
  background: color-mix(in srgb, #ef4444 5%, transparent);
}

.node-id {
  font-size: 11px;
  color: var(--muted);
  font-weight: 500;
  flex-shrink: 0;
}

.node-title {
  flex: 1;
  font-size: 13px;
  color: var(--fg);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

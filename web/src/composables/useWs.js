import { ref, readonly } from 'vue'
import { ElMessage } from 'element-plus'
import i18n from '../locales'

const { t } = i18n.global

const ws = ref(null)
const connected = ref(false)
const handlers = new Map()
let requestId = 0
const pendingRequests = new Map()
let reconnectTimer = null
let heartbeatTimer = null
const HEARTBEAT_INTERVAL = 3000
const HEARTBEAT_TIMEOUT = 10000

function nextId() {
  return `req-${Date.now().toString(36)}-${(++requestId).toString(36)}`
}

function stopHeartbeat() {
  if (heartbeatTimer !== null) {
    clearInterval(heartbeatTimer)
    heartbeatTimer = null
  }
}

function startHeartbeat() {
  stopHeartbeat()
  heartbeatTimer = setInterval(() => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) return

    const timeout = setTimeout(() => {
      console.warn('[ws] heartbeat timeout, closing connection')
      ws.value.close()
    }, HEARTBEAT_TIMEOUT)

    send('ping')
      .then(() => { clearTimeout(timeout) })
      .catch(() => { clearTimeout(timeout) })
  }, HEARTBEAT_INTERVAL)
}

function connect() {
  const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const url = `${protocol}//${location.host}/ws`
  ws.value = new WebSocket(url)

  ws.value.onopen = () => {
    connected.value = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    startHeartbeat()
  }

  ws.value.onclose = () => {
    connected.value = false
    stopHeartbeat()
    ElMessage.error(t('ws.disconnected'))
    reconnectTimer = setTimeout(connect, 3000)
  }

  ws.value.onerror = () => {
    connected.value = false
  }

  ws.value.onmessage = (event) => {
    try {
      const msg = JSON.parse(event.data)
      if (msg.id && pendingRequests.has(msg.id)) {
        const { resolve, reject } = pendingRequests.get(msg.id)
        pendingRequests.delete(msg.id)
        if (msg.ok) {
          resolve(msg.payload)
        } else {
          reject(msg.error || { code: 'unknown', message: t('ws.notConnected') })
        }
        return
      }
      if (handlers.has(msg.type)) {
        handlers.get(msg.type)(msg.payload)
      }
    } catch (e) {
      console.error(t('ws.parseError'), e)
    }
  }
}

function send(type, payload = {}) {
  return new Promise((resolve, reject) => {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      reject({ code: 'not_connected', message: t('ws.notConnected') })
      return
    }
    const id = nextId()
    pendingRequests.set(id, { resolve, reject })
    ws.value.send(JSON.stringify({ id, type, payload }))
    setTimeout(() => {
      if (pendingRequests.has(id)) {
        pendingRequests.delete(id)
        reject({ code: 'timeout', message: t('ws.timeout') })
      }
    }, 30000)
  })
}

function on(eventType, handler) {
  handlers.set(eventType, handler)
}

function off(eventType) {
  handlers.delete(eventType)
}

connect()

export function useWs() {
  return {
    connected: readonly(connected),
    send,
    on,
    off,
  }
}

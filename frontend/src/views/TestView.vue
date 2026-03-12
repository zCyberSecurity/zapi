<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { keyApi, type APIKey } from '../api/index'

interface Message {
  role: 'user' | 'assistant'
  content: string
  error?: boolean
}

const keys = ref<APIKey[]>([])
const selectedKey = ref('')
const model = ref('')
const models = ref<string[]>([])
const modelsLoading = ref(false)
const stream = ref(true)
const userInput = ref('')
const messages = ref<Message[]>([])
const loading = ref(false)
const messagesEl = ref<HTMLElement>()

onMounted(async () => {
  const res = await keyApi.list()
  keys.value = res.data.filter(k => k.enabled)
})

watch(selectedKey, async (key) => {
  model.value = ''
  models.value = []
  if (!key) return
  modelsLoading.value = true
  try {
    const res = await fetch('/v1/models', {
      headers: { Authorization: `Bearer ${key}` },
    })
    const data = await res.json()
    models.value = (data.data ?? []).map((m: any) => m.id)
  } catch {}
  modelsLoading.value = false
})

async function scrollToBottom() {
  await nextTick()
  if (messagesEl.value) {
    messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  }
}

async function send() {
  const text = userInput.value.trim()
  if (!text || !selectedKey.value || !model.value || loading.value) return

  messages.value.push({ role: 'user', content: text })
  userInput.value = ''
  loading.value = true
  scrollToBottom()

  const assistantMsg: Message = { role: 'assistant', content: '' }
  messages.value.push(assistantMsg)
  const idx = messages.value.length - 1

  const body = JSON.stringify({
    model: model.value,
    stream: stream.value,
    messages: messages.value.slice(0, idx).map(m => ({ role: m.role, content: m.content })),
  })

  try {
    if (stream.value) {
      const resp = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${selectedKey.value}`,
        },
        body,
      })

      const streamMsg = messages.value[idx]!
      if (!resp.ok) {
        const err = await resp.text()
        streamMsg.content = err
        streamMsg.error = true
        return
      }

      const reader = resp.body!.getReader()
      const decoder = new TextDecoder()
      let buf = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break
        buf += decoder.decode(value, { stream: true })
        const lines = buf.split('\n')
        buf = lines.pop() ?? ''
        for (const line of lines) {
          if (!line.startsWith('data: ')) continue
          const data = line.slice(6).trim()
          if (data === '[DONE]') break
          try {
            const chunk = JSON.parse(data)
            const delta = chunk.choices?.[0]?.delta?.content
            if (delta) {
              streamMsg.content += delta
              scrollToBottom()
            }
          } catch {}
        }
      }
    } else {
      const resp = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${selectedKey.value}`,
        },
        body,
      })
      const data = await resp.json()
      const msg = messages.value[idx]!
      if (!resp.ok) {
        msg.content = JSON.stringify(data, null, 2)
        msg.error = true
      } else {
        msg.content = data.choices?.[0]?.message?.content ?? JSON.stringify(data, null, 2)
      }
      scrollToBottom()
    }
  } catch (e: any) {
    const msg = messages.value[idx]!
    msg.content = e.message
    msg.error = true
  } finally {
    loading.value = false
    scrollToBottom()
  }
}

function clearMessages() {
  messages.value = []
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send()
  }
}
</script>

<template>
  <div class="chat-container">
    <!-- Toolbar -->
    <div class="chat-toolbar">
      <el-select v-model="selectedKey" placeholder="选择 API Key" size="small" style="width:220px">
        <el-option
          v-for="k in keys"
          :key="k.id"
          :label="`${k.name}(${k.key.slice(0, 20)}...)`"
          :value="k.key"
        />
      </el-select>
      <el-select v-model="model" placeholder="选择 Model" size="small" style="width:200px" :loading="modelsLoading" :disabled="!selectedKey">
        <el-option v-for="m in models" :key="m" :label="m" :value="m" />
      </el-select>
      <el-switch v-model="stream" active-text="流式" inactive-text="普通" size="small" />
      <el-button size="small" @click="clearMessages" :disabled="loading">清空</el-button>
    </div>

    <!-- Messages -->
    <div class="chat-messages" ref="messagesEl">
      <div v-if="messages.length === 0" class="chat-empty">
        选择 API Key 和 Model，开始对话
      </div>
      <div
        v-for="(msg, i) in messages"
        :key="i"
        class="chat-row"
        :class="msg.role"
      >
        <div class="chat-avatar">{{ msg.role === 'user' ? 'U' : 'A' }}</div>
        <div class="chat-bubble" :class="{ error: msg.error }">
          <span v-if="msg.role === 'assistant' && loading && i === messages.length - 1 && msg.content === ''" class="typing-dot" />
          <pre v-else>{{ msg.content }}</pre>
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="chat-input-area">
      <el-input
        v-model="userInput"
        type="textarea"
        :rows="3"
        placeholder="输入消息… Enter 发送，Shift+Enter 换行"
        resize="none"
        @keydown="onKeydown"
        :disabled="loading"
      />
      <el-button
        type="primary"
        :loading="loading"
        :disabled="!selectedKey || !model"
        @click="send"
        style="align-self:flex-end"
      >
        发送
      </el-button>
    </div>
  </div>
</template>

<style scoped>
.chat-container {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  max-width: 860px;
  margin: 0 auto;
}

.chat-toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 0;
  flex-shrink: 0;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px 0;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.chat-empty {
  text-align: center;
  color: #999;
  margin-top: 60px;
  font-size: 14px;
}

.chat-row {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.chat-row.user {
  flex-direction: row-reverse;
}

.chat-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
  background: #e8f4ff;
  color: #409eff;
}

.chat-row.user .chat-avatar {
  background: #f0f9eb;
  color: #67c23a;
}

.chat-bubble {
  max-width: 70%;
  padding: 10px 14px;
  border-radius: 12px;
  background: #f4f4f5;
  font-size: 14px;
  line-height: 1.65;
  word-break: break-word;
}

.chat-row.user .chat-bubble {
  background: #409eff;
  color: #fff;
  border-radius: 12px 2px 12px 12px;
}

.chat-row.assistant .chat-bubble {
  border-radius: 2px 12px 12px 12px;
}

.chat-bubble.error {
  background: #fef0f0;
  color: #f56c6c;
}

.chat-bubble pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: inherit;
}

.typing-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #909399;
  animation: blink 1s infinite;
}

@keyframes blink {
  0%, 100% { opacity: 0.2; }
  50% { opacity: 1; }
}

.chat-input-area {
  display: flex;
  gap: 10px;
  padding: 10px 0;
  flex-shrink: 0;
  border-top: 1px solid #eee;
}
</style>

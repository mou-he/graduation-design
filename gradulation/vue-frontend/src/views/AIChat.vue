<template>
  <div class="ai-chat-container">
    <!-- 左侧会话列表 -->
    <div class="session-list">
      <div class="session-list-header">
        <span>会话列表</span>
        <button class="new-chat-btn" @click="createNewSession">＋ 新聊天</button>
      </div>
      <ul class="session-list-ul">
        <li
          v-for="session in sessions"
          :key="session.id"
          :class="['session-item', { active: currentSessionId === session.id }]"
          @click="switchSession(session.id)"
        >          {{ session.name || `会话 ${session.id}` }}
        <button 
            class="delete-session-btn"
            @click.stop="deleteSession(session.id)"
            :disabled="loading"
          >删除</button>

        </li>
      </ul>
    </div>

    <!-- 右侧聊天区域 -->
    <div class="chat-section">
      <div class="top-bar">
        <button class="back-btn" @click="$router.push('/menu')">← 返回</button>
        <button class="sync-btn" @click="syncHistory" :disabled="!currentSessionId || tempSession">同步历史数据</button>
        <label for="modelType">选择模型：</label>
        <select id="modelType" v-model="selectedModel" class="model-select">
          <option value="1">HaiAI</option>
          <option value="2">HaiAI RAG</option>
          <option value="3">HaiAI MCP</option>
        </select>
        <label for="streamingMode" style="margin-left: 20px;">
          <input type="checkbox" id="streamingMode" v-model="isStreaming" />
          流式响应
        </label>
        <button class="upload-btn" @click="triggerFileUpload" :disabled="uploading">📎 上传文档(.md/.txt)</button>
        <input
          ref="fileInput"
          type="file"
          accept=".md,.txt,text/markdown,text/plain"
          style="display: none"
          @change="handleFileUpload"
        />
      </div>

      <div class="chat-messages" ref="messagesRef">
        <div
          v-for="(message, index) in currentMessages"
          :key="index"
          :class="['message', message.role === 'user' ? 'user-message' : 'ai-message']"
        >
          <div class="message-header">
            <b>{{ message.role === 'user' ? '你' : 'AI' }}:</b>
            <button v-if="message.role === 'assistant'" class="tts-btn" @click="playTTS(message.content)">🔊</button>
            <span v-if="message.meta && message.meta.status === 'streaming'" class="streaming-indicator"> ··</span>
          </div>
          <div class="message-content" v-html="renderMarkdown(message.content)"></div>
        </div>
      </div>

      <div class="chat-input">
        <textarea
          v-model="inputMessage"
          placeholder="请输入你的问题..."
          @keydown.enter.exact.prevent="sendMessage"
          :disabled="loading"
          ref="messageInput"
          rows="1"
        ></textarea>
        <button
          type="button"
          :disabled="!inputMessage.trim() || loading"
          @click="sendMessage"
          class="send-btn"
        >
          {{ loading ? '发送中...' : '发送' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, nextTick, computed, onMounted } from 'vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import api from '../utils/api'

export default {
  name: 'AIChat',
  setup() {
    const sessions = ref({})
    const currentSessionId = ref(null)
    const tempSession = ref(false)
    const currentMessages = ref([])
    const inputMessage = ref('')
    const loading = ref(false)
    const messagesRef = ref(null)
    const messageInput = ref(null)
    const selectedModel = ref('1')
    const isStreaming = ref(false)
    const uploading = ref(false)
    const fileInput = ref(null)

    const renderMarkdown = (text) => {
      if (!text && text !== '') return ''
      return String(text)
        .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
        .replace(/\*(.*?)\*/g, '<em>$1</em>')
        .replace(/`(.*?)`/g, '<code>$1</code>')
        .replace(/\n/g, '<br>')
    }

    const playTTS = async (text) => {
      try {
        const createResponse = await api.post('/AI/chat/tts', { text })
        if (createResponse.data && createResponse.data.status_code === 1000 && createResponse.data.task_id) {
          const taskId = createResponse.data.task_id
          await new Promise(resolve => setTimeout(resolve, 5000))
          const maxAttempts = 30
          const pollInterval = 2000
          let attempts = 0
          const pollResult = async () => {
            const queryResponse = await api.get('/AI/chat/tts/query', { params: { task_id: taskId } })
            if (queryResponse.data && queryResponse.data.status_code === 1000) {
              const taskStatus = queryResponse.data.task_status
              if (taskStatus === 'Success' && queryResponse.data.task_result) {
                const audio = new Audio(queryResponse.data.task_result)
                audio.play()
                return true
              } else if (taskStatus === 'Running' || taskStatus === 'Created') {
                attempts++
                if (attempts < maxAttempts) {
                  await new Promise(resolve => setTimeout(resolve, pollInterval))
                  return await pollResult()
                } else {
                  ElMessage.error('语音合成超时')
                  return true
                }
              } else {
                ElMessage.error('语音合成失败')
                return true
              }
            }
            attempts++
            if (attempts < maxAttempts) {
              await new Promise(resolve => setTimeout(resolve, pollInterval))
              return await pollResult()
            } else {
              ElMessage.error('语音合成超时')
              return true
            }
          }
          await pollResult()
        } else {
          ElMessage.error('无法创建语音合成任务')
        }
      } catch (error) {
        console.error('TTS error:', error)
        ElMessage.error('请求语音接口失败')
      }
    }

   const loadSessions = async () => {
  try {
    const response = await api.get('/AI/chat/sessions')
    if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.sessions)) {
      const sessionMap = {}
      response.data.sessions.forEach(s => {
        const sid = String(s.id)                // 使用 id 字段
        sessionMap[sid] = {
          id: sid,
          name: s.title || `会话 ${sid}`,       // 使用 title 作为显示名称
          messages: []
        }
      })
      sessions.value = sessionMap
      const sessionIds = Object.keys(sessionMap)
      if (sessionIds.length > 0) {
        await switchSession(sessionIds[0])
      } else {
        createNewSession()
      }
    } else {
      createNewSession()
    }
  } catch (error) {
    console.error('Load sessions error:', error)
    createNewSession()
  }
}

    const createNewSession = () => {
      currentSessionId.value = 'temp'
      tempSession.value = true
      currentMessages.value = []
      nextTick(() => {
        if (messageInput.value) messageInput.value.focus()
      })
    }

    const switchSession = async (sessionId) => {
      if (!sessionId) return
      const sid = String(sessionId)
      if (!sessions.value[sid]) {
        ElMessage.warning('会话不存在')
        return
      }
      currentSessionId.value = sid
      tempSession.value = false
      if (!sessions.value[sid].messages || sessions.value[sid].messages.length === 0) {
        try {
          const response = await api.post('/AI/chat/history', { sessionId: sid })
          if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
            const messages = response.data.history.map(item => ({
              role: item.isUser ? 'user' : 'assistant',
              content: item.content
            }))
            sessions.value[sid].messages = messages
          }
        } catch (err) {
          console.error('Load history error:', err)
        }
      }
      currentMessages.value = [...(sessions.value[sid].messages || [])]
      await nextTick()
      scrollToBottom()
    }
    
const deleteSession = async (sessionId) => {

  // 2. 获取会话名称（修复核心：从对象直接取值）
  const session = sessions.value[sessionId]
  if (!session) {
    ElMessage.warning('会话不存在！')
    return
  }
  const sessionName = session.name || `会话 ${sessionId}`

  // 3. 确认删除
  try {
    await ElMessageBox.confirm(
      `确定要删除会话「${sessionName}」吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
  } catch (err) {
    return
  }

  // 4. 调用接口删除
  try {
    loading.value = true
    const response = await api.delete('/AI/chat/delete-session', {
      headers: {
    'Content-Type': 'application/json' // 显式指定JSON格式（部分库会自动加，但建议显式写）
    },
      data: {
      session_id: sessionId, // 必传，对应后端的SessionID
      user_name: ""          // 需传（空值即可），满足参数校验
    }
    })

    if (response.data && response.data.status_code === 1000) {
      // 删除对象中的会话
      delete sessions.value[sessionId]
      sessions.value = { ...sessions.value }

      // 切换会话
      if (currentSessionId.value === sessionId) {
        const remainingSessionIds = Object.keys(sessions.value)
        if (remainingSessionIds.length > 0) {
          await switchSession(remainingSessionIds[0])
        } else {
          createNewSession()
        }
      }

      ElMessage.success('会话删除成功')
    } else {
      ElMessage.error(response.data?.status_msg || '删除会话失败')
    }
  } catch (error) {
    console.error('Delete session error:', error)
    ElMessage.error('删除会话失败，请重试')
  } finally {
    loading.value = false
  }
}
    const syncHistory = async () => {
      if (!currentSessionId.value || tempSession.value) {
        ElMessage.warning('请选择已有会话进行同步')
        return
      }
      try {
        const response = await api.post('/AI/chat/history', { sessionId: currentSessionId.value })
        if (response.data && response.data.status_code === 1000 && Array.isArray(response.data.history)) {
          const messages = response.data.history.map(item => ({
            role: item.isUser ? 'user' : 'assistant',
            content: item.content
          }))
          sessions.value[currentSessionId.value].messages = messages
          currentMessages.value = [...messages]
          await nextTick()
          scrollToBottom()
        } else {
          ElMessage.error('无法获取历史数据')
        }
      } catch (err) {
        console.error('Sync history error:', err)
        ElMessage.error('请求历史数据失败')
      }
    }

    const sendMessage = async () => {
      if (!inputMessage.value || !inputMessage.value.trim()) {
        ElMessage.warning('请输入消息内容')
        return
      }
      const userMessage = {
        role: 'user',
        content: inputMessage.value
      }
      const currentInput = inputMessage.value
      inputMessage.value = ''
      currentMessages.value.push(userMessage)
      await nextTick()
      scrollToBottom()
      try {
        loading.value = true
        if (isStreaming.value) {
          await handleStreaming(currentInput)
        } else {
          await handleNormal(currentInput)
        }
      } catch (err) {
        console.error('Send message error:', err)
        ElMessage.error('发送失败，请重试')
        if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value] && sessions.value[currentSessionId.value].messages) {
          const sessionArr = sessions.value[currentSessionId.value].messages
          if (sessionArr && sessionArr.length) sessionArr.pop()
        }
        currentMessages.value.pop()
      } finally {
        if (!isStreaming.value) {
          loading.value = false
        }
        await nextTick()
        scrollToBottom()
      }
    }

    async function handleStreaming(question) {
      const aiMessage = {
        role: 'assistant',
        content: '',
        meta: { status: 'streaming' }
      }
      const aiMessageIndex = currentMessages.value.length
      currentMessages.value.push(aiMessage)
      if (!tempSession.value && currentSessionId.value && sessions.value[currentSessionId.value]) {
        if (!sessions.value[currentSessionId.value].messages) {
          sessions.value[currentSessionId.value].messages = []
        }
        sessions.value[currentSessionId.value].messages.push({ role: 'user', content: question })
        sessions.value[currentSessionId.value].messages.push({ role: 'assistant', content: '' })
      }
      const url = tempSession.value
        ? '/api/AI/chat/send-stream-new-session'
        : '/api/AI/chat/send-stream'
      const headers = {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
      }
      const body = tempSession.value
        ? { question: question, modelType: selectedModel.value }
        : { question: question, modelType: selectedModel.value, sessionId: currentSessionId.value }
      console.log('[Streaming] Request body:', body)
      try {
        const response = await fetch(url, {
          method: 'POST',
          headers,
          body: JSON.stringify(body)
        })
        if (!response.ok) {
          loading.value = false
          throw new Error(`Network error: ${response.status} ${response.statusText}`)
        }
        const reader = response.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ''
        let newSessionId = null
        let isReading = true
        while (isReading) {
          const { done, value } = await reader.read()
          if (done) break
          const chunk = decoder.decode(value, { stream: true })
          buffer += chunk
          const lines = buffer.split('\n')
          buffer = lines.pop() || ''
          for (const line of lines) {
            const trimmedLine = line.trim()
            if (!trimmedLine) continue
            if (trimmedLine.startsWith('data:')) {
              const data = trimmedLine.slice(5).trim()
              console.log('[SSE] Received:', data)
              if (data === '[DONE]') {
                console.log('[SSE] Stream done')
                loading.value = false
                currentMessages.value[aiMessageIndex].meta = { status: 'done' }
                currentMessages.value = [...currentMessages.value]
                if (newSessionId && tempSession.value) {
                  sessions.value[newSessionId].messages = [...currentMessages.value]
                }
                break
              }
              // FIX: 增强 sessionId 解析逻辑
              try {
                const parsed = JSON.parse(data)
                // 兼容 sessionId 和 session_id 字段
                if (parsed.sessionId || parsed.session_id) {
                  newSessionId = String(parsed.sessionId || parsed.session_id)
                  console.log('[SSE] Got sessionId:', newSessionId)
                  if (tempSession.value) {
                    sessions.value[newSessionId] = {
                      id: newSessionId,
                      name: '新会话',
                      messages: [...currentMessages.value]
                    }
                    currentSessionId.value = newSessionId
                    tempSession.value = false
                    sessions.value = { ...sessions.value } // 强制响应式更新
                  }
                } else {
                  // 如果是 JSON 但没有 sessionId，则视为普通内容
                  currentMessages.value[aiMessageIndex].content += data
                }
              } catch (e) {
                // 不是 JSON，可能是纯文本回复或纯 sessionId
                // 如果还未设置 sessionId 且 data 看起来像 ID（不含空格，长度合适），则视为 sessionId
                if (!newSessionId && !data.includes(' ') && data.length > 5 && tempSession.value) {
                  newSessionId = data
                  console.log('[SSE] Fallback sessionId:', newSessionId)
                  sessions.value[newSessionId] = {
                    id: newSessionId,
                    name: '新会话',
                    messages: [...currentMessages.value]
                  }
                  currentSessionId.value = newSessionId
                  tempSession.value = false
                  sessions.value = { ...sessions.value }
                } else {
                  // 否则作为普通内容追加
                  currentMessages.value[aiMessageIndex].content += data
                }
              }
              currentMessages.value = [...currentMessages.value]
              await new Promise(resolve => {
                requestAnimationFrame(() => {
                  scrollToBottom()
                  resolve()
                })
              })
            }
          }
        }
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'done' }
        currentMessages.value = [...currentMessages.value]
        if (currentSessionId.value && !tempSession.value) {
          const sessMsgs = sessions.value[currentSessionId.value].messages
          if (Array.isArray(sessMsgs) && sessMsgs.length) {
            const lastIndex = sessMsgs.length - 1
            if (sessMsgs[lastIndex] && sessMsgs[lastIndex].role === 'assistant') {
              sessMsgs[lastIndex].content = currentMessages.value[aiMessageIndex].content
              sessions.value[currentSessionId.value].messages = [...sessMsgs] // 强制更新
            }
          }
        }
      } catch (err) {
        console.error('Stream error:', err)
        loading.value = false
        currentMessages.value[aiMessageIndex].meta = { status: 'error' }
        currentMessages.value = [...currentMessages.value]
        ElMessage.error(`流式传输出错：${err.message}`)
      }
    }

    async function handleNormal(question) {
      if (tempSession.value) {
        const response = await api.post('/AI/chat/send-new-session', {
          question: question,
          modelType: selectedModel.value
        })
        if (response.data && response.data.status_code === 1000) {
          const sessionId = String(response.data.sessionId)
          if (!sessionId) {
            throw new Error('后端未返回sessionId')
          }
          const aiMessage = {
            role: 'assistant',
            content: response.data.Information || ''
          }
          sessions.value[sessionId] = {
            id: sessionId,
            name: '新会话',
            messages: [ { role: 'user', content: question }, aiMessage ]
          }
          currentSessionId.value = sessionId
          tempSession.value = false
          currentMessages.value = [...sessions.value[sessionId].messages]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          currentMessages.value.pop()
        }
      } else {
        if (!currentSessionId.value) {
          ElMessage.error('会话ID不存在，请新建会话')
          currentMessages.value.pop()
          loading.value = false
          return
        }
        const sessionMsgs = sessions.value[currentSessionId.value].messages
        sessionMsgs.push({ role: 'user', content: question })
        const response = await api.post('/AI/chat/send', {
          question: question,
          modelType: selectedModel.value,
          sessionId: currentSessionId.value
        })
        if (response.data && response.data.status_code === 1000) {
          const aiMessage = { role: 'assistant', content: response.data.Information || '' }
          sessionMsgs.push(aiMessage)
          currentMessages.value = [...sessionMsgs]
        } else {
          ElMessage.error(response.data?.status_msg || '发送失败')
          sessionMsgs.pop()
          currentMessages.value.pop()
        }
      }
    }

    const scrollToBottom = () => {
      if (messagesRef.value) {
        try {
          messagesRef.value.scrollTop = messagesRef.value.scrollHeight
        } catch (e) {
          // ignore
        }
      }
    }

    const triggerFileUpload = () => {
      if (fileInput.value) {
        fileInput.value.click()
      }
    }

    const handleFileUpload = async (event) => {
      const file = event.target.files[0]
      if (!file) return
      const fileName = file.name.toLowerCase()
      if (!fileName.endsWith('.md') && !fileName.endsWith('.txt')) {
        ElMessage.error('只允许上传 .md 或 .txt 文件')
        if (fileInput.value) {
          fileInput.value.value = ''
        }
        return
      }
      try {
        uploading.value = true
        const formData = new FormData()
        formData.append('file', file)
        const response = await api.post('/file/upload', formData, {
          headers: {
            'Content-Type': 'multipart/form-data'
          }
        })
        if (response.data && response.data.status_code === 1000) {
          ElMessage.success(`文件上传成功`)
        } else {
          ElMessage.error(response.data?.status_msg || '上传失败')
        }
      } catch (error) {
        console.error('File upload error:', error)
        ElMessage.error('文件上传失败')
      } finally {
        uploading.value = false
        if (fileInput.value) {
          fileInput.value.value = ''
        }
      }
    }

    onMounted(() => {
      loadSessions()
    })

    return {
      sessions: computed(() => Object.values(sessions.value)),
      currentSessionId,
      tempSession,
      deleteSession,
      currentMessages,
      inputMessage,
      loading,
      messagesRef,
      messageInput,
      selectedModel,
      isStreaming,
      uploading,
      fileInput,
      renderMarkdown,
      playTTS,
      createNewSession,
      switchSession,
      syncHistory,
      sendMessage,
      triggerFileUpload,
      handleFileUpload
    }
  }
}
</script>

<style scoped>
/* 1. 基础容器与动态背景 */
.ai-chat-container {
  height: 100vh;
  display: flex;
  /* 统一的蓝色渐变背景，更舒适护眼 */
  background: linear-gradient(135deg, #e0c3fc 0%, #8ec5fc 100%); 
  /* 或者更深邃的科技蓝： background: linear-gradient(135deg, #1e3c72 0%, #2a5298 100%); */
  position: relative;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial;
  color: #333;
}

/* 动态粒子背景 */
.ai-chat-container::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.4)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.4)"/><circle cx="40" cy="60" r="1" fill="rgba(255,255,255,0.3)"/></svg>');
  animation: float 20s ease-in-out infinite;
  pointer-events: none;
}

@keyframes float {
  0%, 100% { transform: translateY(0px); }
  50% { transform: translateY(-20px); }
}

/* 2. 左侧侧边栏 (玻璃拟态) */
.session-list {
  width: 280px;
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.65); /* 半透明白 */
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.5);
  box-shadow: 5px 0 15px rgba(0, 0, 0, 0.05);
  z-index: 10;
}

.session-list-header {
  padding: 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);
}

.session-list-header span {
  font-size: 18px;
  font-weight: 700;
  color: #2c3e50;
  letter-spacing: 1px;
}

.new-chat-btn {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: 10px;
  /* 舒适的主蓝色 */
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 0 4px 10px rgba(79, 172, 254, 0.3);
}

.new-chat-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 15px rgba(79, 172, 254, 0.5);
}

.session-list-ul {
  list-style: none;
  padding: 10px;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 12px 16px;
  margin-bottom: 8px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  color: #555;
  font-size: 14px;
}

.session-item:hover {
  background: rgba(255, 255, 255, 0.5);
}

.session-item.active {
  background: rgba(255, 255, 255, 0.9);
  color: #409eff;
  font-weight: 600;
  border-left: 4px solid #409eff;
  box-shadow: 0 2px 8px rgba(0,0,0,0.05);
}

/* 3. 右侧主聊天区 */
.chat-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  position: relative;
  background: rgba(255, 255, 255, 0.2); /* 极淡的遮罩 */
  backdrop-filter: blur(5px);
}

/* 顶部工具栏 */
.top-bar {
  padding: 15px 30px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(15px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  gap: 15px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.02);
}

/* 顶部按钮统一样式 */
.top-bar button, .model-select {
  padding: 8px 16px;
  border-radius: 8px;
  border: 1px solid rgba(255,255,255,0.6);
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  outline: none;
}

.back-btn {
  background: rgba(255,255,255,0.8);
  color: #606266;
}
.back-btn:hover { background: #fff; color: #409eff; }

.sync-btn {
  background: #e1f3d8;
  color: #67c23a;
  border-color: #e1f3d8;
}
.sync-btn:hover { background: #67c23a; color: white; }

.upload-btn {
  background: #fdf6ec;
  color: #e6a23c;
  border-color: #fdf6ec;
}
.upload-btn:hover { background: #e6a23c; color: white; }

.model-select {
  background: rgba(255,255,255,0.9);
  color: #333;
}

/* 4. 消息列表区 */
.chat-messages {
  flex: 1;
  padding: 30px;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* 消息气泡基础 */
.message {
  max-width: 75%;
  padding: 16px 20px;
  border-radius: 16px;
  line-height: 1.6;
  font-size: 15px;
  position: relative;
  box-shadow: 0 4px 15px rgba(0,0,0,0.05);
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 用户消息：蓝色渐变 */
.user-message {
  align-self: flex-end;
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: white;
  border-bottom-right-radius: 4px; /* 稍微方一点的角表示来源 */
}

/* AI消息：纯白玻璃 */
.ai-message {
  align-self: flex-start;
  background: rgba(255, 255, 255, 0.85);
  color: #333;
  border: 1px solid rgba(255,255,255,0.6);
  border-bottom-left-radius: 4px;
}

.message-header {
  font-size: 12px;
  margin-bottom: 6px;
  opacity: 0.8;
  display: flex;
  align-items: center;
  gap: 8px;
}

.tts-btn {
  background: transparent;
  border: 1px solid rgba(0,0,0,0.1);
  border-radius: 50%;
  width: 24px; height: 24px;
  padding: 0;
  display: flex; 
  align-items: center; 
  justify-content: center;
  cursor: pointer;
}

/* 5. 底部输入区 */
.chat-input {
  padding: 20px 30px;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(20px);
  border-top: 1px solid rgba(255, 255, 255, 0.5);
  position: relative;
}

.chat-input textarea {
  width: 100%;
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0,0,0,0.05);
  border-radius: 12px;
  padding: 15px;
  padding-right: 100px; /* 给发送按钮留位置 */
  font-size: 15px;
  color: #333;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
  resize: none;
  transition: all 0.3s;
}

.chat-input textarea:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  background: white;
}

.send-btn {
  position: absolute;
  right: 45px;
  bottom: 35px;
  width: 80px;
  height: 36px;
  border: none;
  border-radius: 18px;
  background: #409eff;
  color: white;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  box-shadow: 0 4px 10px rgba(64, 158, 255, 0.3);
}

.send-btn:hover:not(:disabled) {
  background: #66b1ff;
  transform: translateY(-1px);
}

.send-btn:disabled {
  background: #a0cfff;
  cursor: not-allowed;
  box-shadow: none;
}

/* 自定义滚动条 */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-thumb {
  background: rgba(0, 0, 0, 0.1);
  border-radius: 3px;
}
::-webkit-scrollbar-track { background: transparent; }
</style>

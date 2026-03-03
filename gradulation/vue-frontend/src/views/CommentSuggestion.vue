<template>
  <div class="page-container comment-theme">
    <!-- 未登录提示 -->
    <div v-if="!isLogin" class="login-tip glass-card">
      <el-empty description="请先登录后发表评论" />
      <el-button type="primary" @click="toLogin">去登录</el-button>
    </div>

    <!-- 已登录状态 -->
    <template v-else>
      <el-header class="header">
        <div class="left">
          <el-button link @click="router.back()" class="back-btn">
            <el-icon><ArrowLeft /></el-icon> 返回
          </el-button>
          <h2>评论互动区</h2>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="comment-studio">
          <!-- 发布评论输入区 -->
          <div class="input-section glass-card">
            <el-input
              v-model="publishContent"
              placeholder="分享你的想法（1-500字）"
              class="comment-input"
              :maxlength="500"
              show-word-limit
              type="textarea"
              rows="3"
              clearable
            />
            <el-button
              type="primary"
              size="large"
              @click="handlePublish"
              :loading="isPublishing"
              class="publish-btn"
            >
              发布评论
            </el-button>
          </div>

          <!-- 评论列表区域 -->
          <div class="comment-list glass-card">
            <!-- 根评论列表 -->
            <div v-if="rootComments.length > 0" class="root-comments">
              <div
                v-for="comment in rootComments"
                :key="comment.id"
                class="root-comment-item"
              >
                <!-- 根评论内容 -->
                <div class="comment-header">
                  <span class="username">{{ comment.username }}</span>
                  <span class="create-time">{{ formatTime(comment.create_time) }}</span>
                </div>
                <div class="comment-content">{{ comment.content }}</div>

                <!-- 评论操作区 -->
                <div class="comment-actions">
                  <el-button
                    type="text"
                    @click="toggleReplyPanel(comment.id)"
                    class="reply-btn"
                  >
                    回复 ({{ comment.reply_count }})
                  </el-button>
                  <!-- 仅自己的评论显示编辑/删除 -->
                  <el-button
                    v-if="comment.username === currentUser.username"
                    type="text"
                    @click="showUpdateDialog(comment.id, comment.content)"
                    class="edit-btn"
                  >
                    编辑
                  </el-button>
                  <el-button
                    v-if="comment.username === currentUser.username"
                    type="text"
                    @click="handleDelete(comment.id)"
                    class="delete-btn"
                  >
                    删除
                  </el-button>
                </div>

                <!-- 回复输入面板 -->
                <div v-if="showReplyId === comment.id" class="reply-input-panel">
                  <el-input
                    v-model="replyContent"
                    placeholder="输入你的回复..."
                    :maxlength="500"
                    show-word-limit
                    type="textarea"
                    rows="2"
                    class="reply-input"
                  />
                  <div class="reply-actions">
                    <el-button
                      type="primary"
                      size="small"
                      @click="handleReply(comment.id)"
                      :loading="isReplying"
                    >
                      提交回复
                    </el-button>
                    <el-button size="small" @click="toggleReplyPanel(0)">
                      取消
                    </el-button>
                  </div>
                </div>

                <!-- 回复列表（新增编辑/删除按钮） -->
                <div v-if="showReplyId === comment.id" class="reply-list">
                  <div
                    v-for="reply in replyList[comment.id] || []"
                    :key="reply.id"
                    class="reply-item"
                  >
                    <div class="reply-header">
                      <span class="reply-username">{{ reply.username }}</span>
                      <span v-if="reply.reply_to_username" class="reply-to">
                        回复 @{{ reply.reply_to_username }}
                      </span>
                    </div>
                    <div class="reply-content">{{ reply.content }}</div>
                    <div class="reply-actions">
                      <el-button
                        type="text"
                        size="small"
                        @click="replyToUser(reply.id, reply.user_id, reply.username)"
                      >
                        回复
                      </el-button>
                      <!-- 新增：回复的编辑/删除按钮 -->
                      <el-button
                        v-if="reply.username === currentUser.username"
                        type="text"
                        size="small"
                        @click="showReplyUpdateDialog(reply.id, reply.content, comment.id)"
                        class="reply-edit-btn"
                      >
                        编辑
                      </el-button>
                      <el-button
                        v-if="reply.username === currentUser.username"
                        type="text"
                        size="small"
                        @click="handleReplyDelete(reply.id, comment.id)"
                        class="reply-delete-btn"
                      >
                        删除
                      </el-button>
                    </div>
                  </div>

                  <!-- 回复分页 -->
                  <el-pagination
                    v-if="replyPageInfo[comment.id] && replyPageInfo[comment.id].total > 0"
                    @size-change="handleReplySizeChange(comment.id, $event)"
                    @current-change="handleReplyPageChange(comment.id, $event)"
                    :current-page="(replyPageInfo[comment.id] && replyPageInfo[comment.id].page) || 1"
                    :page-sizes="[5, 10, 20]"
                    :page-size="(replyPageInfo[comment.id] && replyPageInfo[comment.id].page_size) || 10"
                    layout="total, sizes, prev, pager, next, jumper"
                    :total="(replyPageInfo[comment.id] && replyPageInfo[comment.id].total) || 0"
                    small
                    class="reply-pagination"
                  />
                </div>
              </div>
            </div>

            <!-- 空状态 -->
            <div v-else class="empty-comment">
              <el-empty description="暂无评论，快来发表第一条评论吧～" />
            </div>

            <!-- 根评论分页 -->
            <el-pagination
              v-if="rootPageInfo.total > 0"
              @size-change="handleRootSizeChange"
              @current-change="handleRootPageChange"
              :current-page="rootPageInfo.page"
              :page-sizes="[10, 20, 50]"
              :page-size="rootPageInfo.page_size"
              layout="total, sizes, prev, pager, next, jumper"
              :total="rootPageInfo.total"
              class="root-pagination"
            />
          </div>
        </div>
      </el-main>
    </template>

    <!-- 根评论编辑弹窗 -->
    <el-dialog
      v-model="updateDialogVisible"
      title="编辑评论"
      width="500px"
      class="update-dialog"
    >
      <el-input
        v-model="updateContent"
        placeholder="修改你的评论内容..."
        :maxlength="500"
        show-word-limit
        type="textarea"
        rows="4"
      />
      <template #footer>
        <el-button @click="updateDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleUpdateConfirm"
          :loading="isUpdating"
        >
          确认修改
        </el-button>
      </template>
    </el-dialog>

    <!-- 新增：回复编辑弹窗 -->
    <el-dialog
      v-model="replyUpdateDialogVisible"
      title="编辑回复"
      width="400px"
      class="update-dialog"
    >
      <el-input
        v-model="replyUpdateContent"
        placeholder="修改你的回复内容..."
        :maxlength="500"
        show-word-limit
        type="textarea"
        rows="3"
      />
      <template #footer>
        <el-button @click="replyUpdateDialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          @click="handleReplyUpdateConfirm"
          :loading="isReplyUpdating"
        >
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, computed, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from 'axios'

// 初始化路由
const router = useRouter()
const route = useRoute()

// 取消请求令牌
const cancelSources = ref([])

// 创建axios实例
const createRequest = () => {
  const request = axios.create({
    baseURL: '/api',
    timeout: 10000,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })

  // 请求拦截器
  request.interceptors.request.use(
    config => {
      const source = axios.CancelToken.source()
      config.cancelToken = source.token
      cancelSources.value.push(source)

      const userInfo = JSON.parse(localStorage.getItem('userInfo') || '{"userId":"","username":""}')
      if (userInfo.userId && userInfo.username) {
        config.headers['X-User-ID'] = userInfo.userId
        config.headers['X-Username'] = userInfo.username
        config.headers['Authorization'] = 'Bearer ' + localStorage.getItem('token')
      }
      return config
    },
    error => Promise.reject(error)
  )

  // 响应拦截器
  request.interceptors.response.use(
    response => {
      const res = response.data
      if (res.status_code !== 1000) {
        ElMessage.error(res.status_msg || '操作失败')
        return Promise.reject(res)
      }
      return res
    },
    error => {
      if (axios.isCancel(error)) return Promise.reject(error)
      ElMessage.error(error.message || '网络异常，请稍后重试')
      return Promise.reject(error)
    }
  )

  return request
}

const request = createRequest()

// 当前登录用户
const currentUser = ref(JSON.parse(localStorage.getItem('userInfo') || '{"userId":"","username":""}'))
const isLogin = computed(() => !!currentUser.value.username)

// 监听localStorage变化
const handleStorageChange = () => {
  currentUser.value = JSON.parse(localStorage.getItem('userInfo') || '{"userId":"","username":""}')
}

// 发布评论相关
const publishContent = ref('')
const isPublishing = ref(false)

// 根评论相关
const rootComments = ref([])
const rootPageInfo = reactive({
  page: 1,
  page_size: 10,
  total: 0
})

// 回复相关
const showReplyId = ref(0)
const replyContent = ref('')
const isReplying = ref(false)
const replyToUserId = ref('')
const replyToUsername = ref('')
const replyList = ref({})
const replyPageInfo = reactive({})

// 根评论编辑相关
const updateDialogVisible = ref(false)
const updateContent = ref('')
const updateCommentId = ref(0)
const isUpdating = ref(false)

// 新增：回复编辑相关
const replyUpdateDialogVisible = ref(false)
const replyUpdateContent = ref('')
const replyUpdateId = ref(0)
const replyUpdateRootId = ref(0)
const isReplyUpdating = ref(false)

// 格式化时间
const formatTime = (timeStr) => {
  if (!timeStr) return ''
  try {
    const date = new Date(timeStr)
    const year = date.getFullYear()
    const month = ('0' + (date.getMonth() + 1)).slice(-2)
    const day = ('0' + date.getDate()).slice(-2)
    const hour = ('0' + date.getHours()).slice(-2)
    const minute = ('0' + date.getMinutes()).slice(-2)
    return `${year}-${month}-${day} ${hour}:${minute}`
  } catch (e) {
    return '未知时间'
  }
}

// 获取根评论列表
const getRootComments = async () => {
  try {
    const res = await request.post('/suggestion/root/page', {
      page: rootPageInfo.page,
      page_size: rootPageInfo.page_size
    })
    rootComments.value = res.list || []
    rootPageInfo.total = res.total || 0
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error('获取根评论失败：', err)
    }
  }
}

// 获取指定根评论的回复列表
const getReplyList = async (rootId) => {
  if (!replyPageInfo[rootId]) {
    replyPageInfo[rootId] = {
      page: 1,
      page_size: 10,
      total: 0
    }
  }
  try {
    const res = await request.post('/suggestion/reply/page', {
      root_id: rootId,
      page: replyPageInfo[rootId].page,
      page_size: replyPageInfo[rootId].page_size
    })
    replyList.value[rootId] = res.list || []
    replyPageInfo[rootId].total = res.total || 0
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error(`获取根评论${rootId}的回复失败：`, err)
    }
  }
}

// 发布根评论
const handlePublish = async () => {
  if (!publishContent.value.trim()) {
    return ElMessage.warning('评论内容不能为空')
  }
  if (!isLogin.value) {
    return ElMessage.warning('请先登录')
  }

  isPublishing.value = true
  try {
    await request.post('/suggestion/publish', {
      content: publishContent.value.trim()
    })
    ElMessage.success('评论发布成功')
    publishContent.value = ''
    rootPageInfo.page = 1
    await getRootComments()
    window.scrollTo({ top: 0, behavior: 'smooth' })
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error('发布评论失败：', err)
    }
  } finally {
    isPublishing.value = false
  }
}

// 切换回复面板显示
const toggleReplyPanel = (rootId) => {
  showReplyId.value = rootId
  if (rootId > 0) {
    if (replyPageInfo[rootId]) {
      replyPageInfo[rootId].page = 1
    }
    getReplyList(rootId)
    replyToUserId.value = ''
    replyToUsername.value = ''
  }
  replyContent.value = ''
}

// 回复指定用户
const replyToUser = (replyId, userId, username) => {
  replyToUserId.value = userId
  replyToUsername.value = username
  replyContent.value = `@${username} `
  nextTick(() => {
    const replyInput = document.querySelector('.reply-input .el-textarea__inner')
    if (replyInput) replyInput.focus()
  })
}

// 提交回复
const handleReply = async (rootId) => {
  if (!replyContent.value.trim()) {
    return ElMessage.warning('回复内容不能为空')
  }
  if (!isLogin.value) {
    return ElMessage.warning('请先登录')
  }

  isReplying.value = true
  try {
    await request.post('/suggestion/reply', {
      parent_id: rootId,
      reply_to_user_id: replyToUserId.value,
      reply_to_username: replyToUsername.value,
      content: replyContent.value.trim()
    })
    ElMessage.success('回复成功')
    replyContent.value = ''
    replyPageInfo[rootId].page = 1
    await getReplyList(rootId)
    await getRootComments()
    nextTick(() => {
      const replyList = document.querySelector(`.reply-list`)
      if (replyList) {
        replyList.scrollTop = replyList.scrollHeight
      }
    })
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error('回复失败：', err)
    }
  } finally {
    isReplying.value = false
  }
}

// 显示根评论编辑弹窗
const showUpdateDialog = (id, content) => {
  updateCommentId.value = id
  updateContent.value = content
  updateDialogVisible.value = true
}

// 确认更新根评论
const handleUpdateConfirm = async () => {
  if (!updateContent.value.trim()) {
    return ElMessage.warning('评论内容不能为空')
  }
  if (!isLogin.value) {
    return ElMessage.warning('请先登录')
  }

  isUpdating.value = true
  try {
    await request.post('/suggestion/update', {
      id: updateCommentId.value,
      content: updateContent.value.trim()
    })
    ElMessage.success('评论修改成功')
    updateDialogVisible.value = false
    await getRootComments()
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error('更新评论失败：', err)
    }
  } finally {
    isUpdating.value = false
  }
}

// 新增：显示回复编辑弹窗
const showReplyUpdateDialog = (id, content, rootId) => {
  replyUpdateId.value = id
  replyUpdateContent.value = content
  replyUpdateRootId.value = rootId
  replyUpdateDialogVisible.value = true
}

// 新增：确认更新回复
const handleReplyUpdateConfirm = async () => {
  if (!replyUpdateContent.value.trim()) {
    return ElMessage.warning('回复内容不能为空')
  }
  if (!isLogin.value) {
    return ElMessage.warning('请先登录')
  }

  isReplyUpdating.value = true
  try {
    await request.post('/suggestion/update', {
      id: replyUpdateId.value,
      content: replyUpdateContent.value.trim()
    })
    ElMessage.success('回复修改成功')
    replyUpdateDialogVisible.value = false
    await getReplyList(replyUpdateRootId.value)
    await getRootComments()
  } catch (err) {
    if (!axios.isCancel(err)) {
      console.error('更新回复失败：', err)
    }
  } finally {
    isReplyUpdating.value = false
  }
}

// 删除根评论
const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这条评论吗？删除后无法恢复',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await request.post('/suggestion/delete', { id })
    ElMessage.success('评论删除成功')
    await getRootComments()
  } catch (err) {
    if (err !== 'cancel' && !axios.isCancel(err)) {
      console.error('删除评论失败：', err)
    }
  }
}

// 新增：删除回复
const handleReplyDelete = async (replyId, rootId) => {
  try {
    await ElMessageBox.confirm(
      '确定要删除这条回复吗？删除后无法恢复',
      '提示',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    await request.post('/suggestion/delete', { id: replyId })
    ElMessage.success('回复删除成功')
    await getReplyList(rootId)
    await getRootComments()
  } catch (err) {
    if (err !== 'cancel' && !axios.isCancel(err)) {
      console.error('删除回复失败：', err)
    }
  }
}

// 根评论分页事件
const handleRootSizeChange = (val) => {
  rootPageInfo.page_size = val
  getRootComments()
}

const handleRootPageChange = (val) => {
  rootPageInfo.page = val
  getRootComments()
}

// 回复分页事件
const handleReplySizeChange = (rootId, val) => {
  replyPageInfo[rootId].page_size = val
  getReplyList(rootId)
}

const handleReplyPageChange = (rootId, val) => {
  replyPageInfo[rootId].page = val
  getReplyList(rootId)
}

// 跳转到登录页
const toLogin = () => {
  router.push({ path: '/login', query: { redirect: route.fullPath } })
}

// 组件生命周期
onMounted(() => {
  window.addEventListener('storage', handleStorageChange)
  if (isLogin.value) {
    getRootComments()
  }
})

onUnmounted(() => {
  window.removeEventListener('storage', handleStorageChange)
  cancelSources.value.forEach(source => source.cancel('组件已卸载'))
  cancelSources.value = []
})
</script>

<style scoped>
.comment-theme {
  background: linear-gradient(135deg, #2c3e50 0%, #000000 100%);
}

.page-container {
  min-height: 100vh;
  color: white;
}

/* 登录提示 */
.login-tip {
  margin: 20px;
  padding: 40px;
  text-align: center;
}

.header {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  display: flex;
  align-items: center;
  padding: 0 20px;
}

.left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.back-btn {
  color: white !important;
  font-size: 16px;
}

.main-content {
  padding: 30px;
  display: flex;
  justify-content: center;
}

.comment-studio {
  width: 100%;
  max-width: 800px;
  display: flex;
  flex-direction: column;
  gap: 30px;
}

/* 玻璃拟态卡片 */
.glass-card {
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* 发布评论区域 */
.input-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.comment-input :deep(.el-textarea__wrapper) {
  background: rgba(0,0,0,0.3);
  border-radius: 15px;
  box-shadow: none;
  border: 1px solid #555;
}

.comment-input :deep(.el-textarea__inner) {
  color: white;
  font-size: 16px;
  line-height: 1.5;
}

.publish-btn {
  align-self: flex-end;
  border-radius: 50px;
  padding: 10px 30px;
}

/* 评论列表区域 */
.comment-list {
  min-height: 400px;
}

.root-comments {
  display: flex;
  flex-direction: column;
  gap: 25px;
}

.root-comment-item {
  padding-bottom: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.comment-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
}

.username {
  font-weight: 600;
  color: #4facfe;
}

.create-time {
  font-size: 12px;
  color: #aeb6bf;
}

.comment-content {
  line-height: 1.6;
  margin-bottom: 15px;
  font-size: 15px;
}

.comment-actions {
  display: flex;
  gap: 15px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.comment-actions :deep(.el-button) {
  color: #bdc3c7;
}

/* 回复区域 */
.reply-input-panel {
  margin-top: 20px;
  padding: 20px;
  background: rgba(0,0,0,0.2);
  border-radius: 15px;
  margin-bottom: 20px;
}

.reply-input {
  margin-bottom: 15px;
}

.reply-input :deep(.el-textarea__wrapper) {
  background: rgba(0,0,0,0.3);
  border-radius: 10px;
}

.reply-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  align-items: center;
  flex-wrap: wrap;
}

.reply-list {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-left: 20px;
  border-left: 2px solid rgba(255, 255, 255, 0.1);
  max-height: 400px;
  overflow-y: auto;
}

.reply-item {
  padding: 15px;
  background: rgba(0,0,0,0.15);
  border-radius: 10px;
}

.reply-header {
  margin-bottom: 8px;
}

.reply-username {
  font-weight: 500;
  color: #00f2fe;
}

.reply-to {
  font-size: 12px;
  color: #aeb6bf;
  margin-left: 8px;
}

.reply-content {
  font-size: 14px;
  line-height: 1.5;
  margin-bottom: 10px;
}

/* 回复编辑/删除按钮样式 */
.reply-edit-btn {
  color: #3498db !important;
}

.reply-delete-btn {
  color: #e74c3c !important;
}

/* 分页样式 */
.root-pagination {
  margin-top: 30px;
  text-align: center;
}

.reply-pagination {
  margin-top: 15px;
  text-align: right;
}

:deep(.el-pagination) {
  --el-pagination-text-color: white;
}

:deep(.el-pagination button) {
  color: white !important;
}

/* 空状态 */
.empty-comment {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 编辑弹窗 */
.update-dialog :deep(.el-dialog) {
  background: rgba(0,0,0,0.8);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.update-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.update-dialog :deep(.el-dialog__title) {
  color: white;
}

.update-dialog :deep(.el-textarea__wrapper) {
  background: rgba(0,0,0,0.3);
  color: white;
}

/* 响应式适配 */
@media (max-width: 768px) {
  .main-content {
    padding: 15px;
  }
  .glass-card {
    padding: 20px;
  }
  .comment-actions {
    gap: 10px;
  }
  .reply-list {
    padding-left: 10px;
  }
  .reply-actions {
    justify-content: flex-start;
  }
}
</style>
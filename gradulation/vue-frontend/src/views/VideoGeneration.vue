<template>
  <div class="page-container">
    <!-- 顶部导航 -->
    <el-header class="header">
      <div class="header-content">
        <el-button link @click="$router.back()" class="back-btn">
          <el-icon><ArrowLeft /></el-icon> 返回
        </el-button>
        <h2>AI 视频创作工坊</h2>
      </div>
    </el-header>

    <el-main class="main-content">
      <div class="content-wrapper">
        <!-- 左侧：控制面板 -->
        <div class="control-panel glass-card">
          <div class="panel-header">
            <h3><el-icon><EditPen /></el-icon> 灵感描述</h3>
            <span class="sub-title">用文字描绘你的想象</span>
          </div>
          
          <el-input
            v-model="form.prompt"
            type="textarea"
            :rows="6"
            placeholder="例如：一只蓝色的鲸鱼在云端遨游，阳光穿透云层，梦幻唯美，吉卜力画风..."
            class="custom-input"
            resize="none"
          />
          
          <div class="settings-grid">
            <div class="setting-item">
              <span class="label">生成模式</span>
              <el-select v-model="form.quality" class="custom-select" popper-class="light-popper">
                <el-option label="画质优先 (Quality)" value="quality" />
                <el-option label="极速生成 (Speed)" value="speed" />
              </el-select>
            </div>
            <div class="setting-item">
              <span class="label">视频时长</span>
              <el-radio-group v-model="form.duration" size="default" class="custom-radio">
                <el-radio-button :label="5">5秒</el-radio-button>
                <el-radio-button :label="10">10秒</el-radio-button>
              </el-radio-group>
            </div>
          </div>

           <div class="setting-item audio-setting">
              <el-checkbox v-model="form.with_audio" class="custom-checkbox">
                包含音效 (AI Sound)
              </el-checkbox>
           </div>

          <el-button 
            type="primary" 
            class="generate-btn" 
            :loading="isGenerating"
            @click="handleGenerate"
          >
            {{ btnText }}
          </el-button>
        </div>

        <!-- 右侧：预览/结果区域 -->
        <div class="preview-panel glass-card">
          <!-- 初始状态 -->
          <div v-if="!resultVideo && !isGenerating" class="empty-state">
            <div class="icon-bg">
              <el-icon size="50"><VideoPlay /></el-icon>
            </div>
            <p class="main-tip">视频预览区域</p>
            <p class="sub-tip">在左侧输入描述后点击生成</p>
          </div>

          <!-- 生成中状态 -->
          <div v-if="isGenerating" class="loading-state">
            <div class="loader"></div>
            <p>AI 正在绘制画面...</p>
            <el-tag effect="light" round>{{ statusText }}</el-tag>
          </div>

          <!-- 结果展示 -->
          <div v-if="resultVideo" class="video-result">
    <div class="video-wrapper">
      <!-- video标签会自动请求src指向的播放代理接口 -->
      <video 
        controls 
        autoplay 
        loop 
        :src="resultVideo" 
        class="result-player"
        playsinline
        webkit-playsinline
        x5-playsinline
      >
        您的浏览器不支持HTML5视频播放，请升级浏览器
      </video>
    </div>
            
            <div class="actions">
              <!-- 导出按钮 -->
              <el-button type="primary" size="large" :icon="Download" round @click="downloadVideo" class="action-btn">
                导出视频 (MP4)
              </el-button>
            </div>
            
            <div v-if="failReason" class="error-msg">
               <el-alert :title="'生成失败: ' + failReason" type="error" show-icon />
            </div>
          </div>
        </div>
      </div>
    </el-main>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onUnmounted } from 'vue'
import { ArrowLeft, EditPen, VideoPlay, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import axios from 'axios'


axios.interceptors.request.use(
  (config) => {
    // 从localStorage读取token（也可以用sessionStorage，根据需求选择）
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

// 表单数据
const form = reactive({
  prompt: '',
  quality: 'quality', 
  duration: 5,        
  with_audio: true    
})

const isGenerating = ref(false)
const resultVideo = ref('')
const statusText = ref('') 
const failReason = ref('')
let pollTimer = null

const btnText = computed(() => {
  if (isGenerating.value) return 'AI 正在渲染中...'
  return '立即生成'
})

// 提交生成任务
const handleGenerate = async () => {
  if (!form.prompt) return ElMessage.warning('请输入视频描述')
  
  isGenerating.value = true
  resultVideo.value = ''
  failReason.value = ''
  statusText.value = '提交任务中...'

  try {
    const res = await axios.post('/api/video/generate', form)
    
    if (res.data.status_code === 1000) {
      // ========== 可选：如果接口返回新的token，这里保存 ==========
      if (res.data.token) {
        localStorage.setItem('video_ai_token', res.data.token)
      }

      const taskId = res.data.task_id
      ElMessage.success('任务已提交，AI开始创作...')
      startPolling(taskId)
    } else {
      ElMessage.error(res.data.msg || '提交失败')
      isGenerating.value = false
    }
  } catch (error) {
    console.error(error)
    // ========== 新增：token过期处理 ==========
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      // 清除过期token并跳转到登录页
      localStorage.removeItem('video_ai_token')
      // 假设登录页路由是 /login
      // $router.push('/login')
    } else {
      ElMessage.error('网络请求失败')
    }
    isGenerating.value = false
  }
}

// 轮询查询状态
const startPolling = (taskId) => {
  pollTimer = setInterval(async () => {
    try {
      const res = await axios.get(`/api/video/status/${taskId}`)
      
      if (res.data.status_code === 1000) {
        const data = res.data.data
        statusText.value = mapStatusText(data.status)

        if (data.status === 'SUCCESS') {
          clearInterval(pollTimer)
          resultVideo.value = data.video_url
          isGenerating.value = false
          ElMessage.success('视频生成完成！')
        } else if (data.status === 'FAIL') {
          clearInterval(pollTimer)
          isGenerating.value = false
          failReason.value = data.fail_error
          ElMessage.error('视频生成失败')
        }
      } else {
        clearInterval(pollTimer)
        isGenerating.value = false
        ElMessage.error(res.data.msg)
      }
    } catch (error) {
      console.error('轮询出错', error)
    }
  }, 3000)
}

const mapStatusText = (status) => {
  const map = {
    'PENDING': '排队中...',
    'PROCESSING': '正在生成画面...',
    'SUCCESS': '生成成功',
    'FAIL': '生成失败'
  }
  return map[status] || status
}

// 导出/下载视频
const downloadVideo = () => {
  if (!resultVideo.value) return
  
  // 创建一个隐藏的a标签来触发下载，体验更好
  const link = document.createElement('a')
  link.href = resultVideo.value
  // 如果是后端代理流，这里设download属性可以提示浏览器下载
  // 假设resultVideo是后端代理地址 /api/v1/video/download/xxx
  link.setAttribute('download', `ai_video_${Date.now()}.mp4`) 
  link.style.display = 'none'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
})
</script>

<style scoped>
/* ================= 全局容器 ================= */
.page-container {
  min-height: 100vh;
  /* 淡蓝色渐变背景 */
  background: linear-gradient(120deg, #e0c3fc 0%, #8ec5fc 100%); 
  /* 或者更清新的纯蓝系: background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%); */
  background: linear-gradient(135deg, #f0f7ff 0%, #d4e4f7 100%);
  color: #2c3e50; /* 深色字体 */
  display: flex;
  flex-direction: column;
}

/* ================= 头部 ================= */
.header {
  background: rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.header-content {
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  display: flex;
  align-items: center;
  gap: 15px;
}

.header h2 {
  font-size: 20px;
  color: #334e68;
  font-weight: 600;
  margin: 0;
}

.back-btn {
  font-size: 16px;
  color: #555;
  font-weight: normal;
}
.back-btn:hover {
  color: #409eff;
}

/* ================= 内容布局 ================= */
.main-content {
  padding: 20px;
}

.content-wrapper {
  display: flex;
  gap: 25px;
  max-width: 1200px;
  margin: 0 auto;
  height: calc(100vh - 120px); /* 适应屏幕高度 */
  align-items: stretch;
}

/* ================= 通用卡片 ================= */
.glass-card {
  background: rgba(255, 255, 255, 0.85); /* 高透明度白色 */
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 30px;
  border: 1px solid rgba(255, 255, 255, 0.9);
  box-shadow: 0 10px 40px rgba(142, 197, 252, 0.2); /* 蓝色系阴影 */
  display: flex;
  flex-direction: column;
}

/* ================= 左侧控制面板 ================= */
.control-panel {
  flex: 1;
  min-width: 350px;
  gap: 24px;
}

.panel-header h3 {
  margin: 0 0 5px 0;
  font-size: 18px;
  color: #102a43;
  display: flex;
  align-items: center;
  gap: 8px;
}

.sub-title {
  font-size: 13px;
  color: #829ab1;
}

.settings-grid {
  display: flex;
  gap: 15px;
}

.setting-item {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.label {
  font-size: 14px;
  color: #486581;
  font-weight: 500;
}

.audio-setting {
  margin-top: -10px;
}

/* 按钮样式 */
.generate-btn {
  margin-top: auto; /* 推到底部 */
  height: 52px;
  font-size: 16px;
  letter-spacing: 1px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); /* 蓝紫渐变 */
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); /* 清新蓝渐变 */
  border: none;
  border-radius: 12px;
  box-shadow: 0 4px 15px rgba(79, 172, 254, 0.4);
  transition: all 0.3s ease;
}
.generate-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(79, 172, 254, 0.6);
}

/* ================= 右侧预览面板 ================= */
.preview-panel {
  flex: 1.6;
  position: relative;
  justify-content: center;
  align-items: center;
  background: rgba(255, 255, 255, 0.9);
}

/* 空状态 */
.empty-state {
  text-align: center;
  color: #829ab1;
}
.icon-bg {
  width: 100px;
  height: 100px;
  background: #f0f4f8;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  color: #bcccdc;
}
.main-tip {
  font-size: 18px;
  font-weight: 600;
  color: #334e68;
  margin-bottom: 5px;
}

/* 加载状态 */
.loading-state {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
  color: #486581;
}
.loader {
  width: 50px;
  height: 50px;
  border: 4px solid #e1e4e8;
  border-bottom-color: #409eff;
  border-radius: 50%;
  animation: rotation 1s linear infinite;
}

/* 结果展示 */
.video-result {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.video-wrapper {
  flex: 1;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 10px 30px rgba(0,0,0,0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}
.result-player {
  width: 100%;
  height: 100%;
  max-height: 500px;
  object-fit: contain;
}

.actions {
  display: flex;
  justify-content: center;
}
.action-btn {
  width: 200px;
  background: #409eff;
  box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
}

/* ================= Element Plus 覆盖 (适配浅色主题) ================= */
:deep(.custom-input .el-textarea__inner) {
  background-color: #f7f9fc !important;
  border: 1px solid #dceefb !important;
  color: #334e68 !important;
  border-radius: 12px;
  padding: 15px;
  font-size: 15px;
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.02);
}
:deep(.custom-input .el-textarea__inner:focus) {
  border-color: #409eff !important;
  background-color: #fff !important;
}

:deep(.custom-select .el-input__wrapper) {
  background-color: #f7f9fc !important;
  border: 1px solid #dceefb !important;
  border-radius: 8px;
  box-shadow: none !important;
}

:deep(.custom-radio .el-radio-button__inner) {
  background: #f7f9fc;
  border-color: #dceefb;
  color: #486581;
}
:deep(.custom-radio .el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background-color: #409eff;
  border-color: #409eff;
  color: white;
  box-shadow: none;
}

:deep(.custom-checkbox.is-bordered) {
  background-color: #f7f9fc;
  border-color: #dceefb;
  width: 100%;
  border-radius: 8px;
}
:deep(.custom-checkbox.is-bordered.is-checked) {
  border-color: #409eff;
  background-color: #ecf5ff;
}

@keyframes rotation {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@media (max-width: 768px) {
  .content-wrapper {
    flex-direction: column;
    height: auto;
  }
  .preview-panel {
    min-height: 300px;
  }
}
</style>
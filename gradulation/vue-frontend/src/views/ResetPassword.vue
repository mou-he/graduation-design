<template>
  <div class="forget-container">
    <el-card class="forget-card">
      <template #header>
        <div class="card-header">
          <h2>找回密码</h2>
        </div>
      </template>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <!-- 邮箱 -->
        <el-form-item label="邮箱" prop="email">
          <el-input
            v-model="form.email"
            placeholder="请输入你的邮箱"
            clearable
          />
        </el-form-item>

        <!-- 验证码 -->
        <el-form-item label="验证码" prop="captcha"> <!-- prop 改为 captcha -->
          <div style="display: flex; gap: 10px;">
            <el-input v-model="form.captcha" placeholder="请输入验证码" /> <!-- v-model 改为 captcha -->
            <el-button
              type="primary"
              :disabled="codeSending || countdown > 0"
              @click="sendCode"
              style="width: 120px;"
            >
              {{ countdown > 0 ? `${countdown}秒后重试` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <!-- 新密码 -->
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="form.newPassword"
            type="password"
            placeholder="请输入新密码"
            show-password
          />
        </el-form-item>

        <!-- 确认新密码 -->
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="submitting"
            @click="handleSubmit"
            style="width: 100%; height: 48px; font-size: 16px;"
          >
            确认重置
          </el-button>
        </el-form-item>

        <el-form-item>
          <div style="display: flex; justify-content: space-between; width: 100%;">
            <el-button type="text" @click="$router.push('/login')">返回登录</el-button>
            <el-button type="text" @click="$router.push('/register')">去注册</el-button>
          </div>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../utils/api' // 根据你的实际路径调整

export default {
  name: 'ResetPassword',
  setup() {
    const router = useRouter()
    const formRef = ref(null)
    const submitting = ref(false)
    const codeSending = ref(false)
    const countdown = ref(0)
    let timer = null

    // 表单数据，将 code 更名为 captcha 以匹配后端
    const form = reactive({
      email: '',
      captcha: '', // 将 code 更名为 captcha
      newPassword: '',
      confirmPassword: ''
    })

    // 邮箱正则
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

    // 表单校验规则
    const rules = {
      email: [
        { required: true, message: '请输入邮箱', trigger: 'blur' },
        { pattern: emailRegex, message: '请输入正确的邮箱格式', trigger: 'blur' }
      ],
      captcha: [ // 规则名称也改为 captcha
        { required: true, message: '请输入验证码', trigger: 'blur' },
        { len: 6, message: '验证码应为6位数字', trigger: 'blur' }
      ],
      newPassword: [
        { required: true, message: '请输入新密码', trigger: 'blur' },
        { min: 6, message: '密码长度不能少于6位', trigger: 'blur' },
        { max: 20, message: '密码长度不能超过20位', trigger: 'blur' } // 额外添加最大长度校验
      ],
      confirmPassword: [
        { required: true, message: '请再次输入新密码', trigger: 'blur' },
        { validator: (rule, value, callback) => {
          if (value !== form.newPassword) {
            callback(new Error('两次输入密码不一致'))
          } else {
            callback()
          }
        }, trigger: 'blur' }
      ]
    }

    // 发送验证码
    const sendCode = async () => {
      // 先校验邮箱是否填写且格式正确
      try {
        await formRef.value.validateField('email')
      } catch {
        return
      }

      codeSending.value = true
      try {
        // 调用发送验证码接口，修改为后端定义的路径
        await api.post('/user/captcha', { email: form.email })
        ElMessage.success('验证码已发送到邮箱')
        // 开始倒计时
        countdown.value = 60
        if (timer) clearInterval(timer)
        timer = setInterval(() => {
          countdown.value--
          if (countdown.value <= 0) {
            clearInterval(timer)
            timer = null
          }
        }, 1000)
      } catch (error) {
        ElMessage.error(error.response?.data?.status_msg || '发送失败')
      } finally {
        codeSending.value = false
      }
    }

    // 提交重置密码
    const handleSubmit = async () => {
      if (!formRef.value) return
      await formRef.value.validate()
      submitting.value = true
      try {
        // 调用重置密码接口，修改请求体字段名以匹配后端，并添加 confirm_password
        await api.post('/user/reset-password', {
          email: form.email,
          captcha: form.captcha,         // 对应后端 'captcha'
          new_password: form.newPassword, // 对应后端 'new_password'
          confirm_password: form.confirmPassword // 对应后端 'confirm_password'
        })
        ElMessage.success('密码重置成功，请登录')
        router.push('/login')
      } catch (error) {
        ElMessage.error(error.response?.data?.status_msg || '重置失败')
      } finally {
        submitting.value = false
      }
    }

    return {
      formRef,
      form,
      rules,
      submitting,
      codeSending,
      countdown,
      sendCode,
      handleSubmit
    }
  }
}
</script>

<style scoped>
/* 样式保持不变 */
.forget-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  position: relative;
  overflow: hidden;
}

.forget-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="80" cy="80" r="2" fill="rgba(255,255,255,0.1)"/><circle cx="40" cy="60" r="1" fill="rgba(255,255,255,0.1)"/><circle cx="60" cy="30" r="1.5" fill="rgba(255,255,255,0.1)"/></svg>');
  animation: float 20s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0px) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(180deg); }
}

.forget-card {
  width: 460px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 20px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.2);
  animation: slideIn 0.8s ease-out;
  position: relative;
  z-index: 1;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(30px) scale(0.9);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.card-header {
  text-align: center;
  padding: 30px 0 20px 0;
}

.card-header h2 {
  margin: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 28px;
  font-weight: 600;
  animation: glow 2s ease-in-out infinite alternate;
}

@keyframes glow {
  from { filter: brightness(1); }
  to { filter: brightness(1.2); }
}

.el-form-item {
  margin-bottom: 24px;
}

.el-input {
  transition: all 0.3s ease;
}

.el-input:focus-within {
  transform: scale(1.02);
}

.el-button {
  border-radius: 12px;
  font-weight: 600;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.el-button::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  transition: left 0.5s;
}

.el-button:hover::before {
  left: 100%;
}

.el-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(64, 158, 255, 0.3);
}
</style>
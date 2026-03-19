<template>
  <div class="settings-page">
    <el-card shadow="never">
      <template #header>
        <span>系统设置</span>
      </template>

      <el-tabs v-model="activeTab">
        <!-- 邮件设置 -->
        <el-tab-pane label="邮件设置" name="email">
          <el-form label-width="120px" style="max-width: 600px;">
            <el-form-item label="启用邮件">
              <el-switch v-model="settings.mailEnabled" />
            </el-form-item>
            <el-form-item label="SMTP 服务器">
              <el-input v-model="settings.smtpHost" placeholder="smtp.example.com" :disabled="!settings.mailEnabled" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="settings.smtpPort" :min="1" :max="65535" :disabled="!settings.mailEnabled" />
            </el-form-item>
            <el-form-item label="发件人邮箱">
              <el-input v-model="settings.smtpUser" placeholder="noreply@example.com" :disabled="!settings.mailEnabled" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="settings.smtpPassword" type="password" show-password :disabled="!settings.mailEnabled" />
            </el-form-item>
            <el-form-item label="发件人名称">
              <el-input v-model="settings.smtpName" placeholder="系统通知" :disabled="!settings.mailEnabled" />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" :disabled="!settings.mailEnabled">保存设置</el-button>
              <el-button :disabled="!settings.mailEnabled">发送测试邮件</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 邮件模板 -->
        <el-tab-pane label="邮件模板" name="emailTemplates">
          <div class="section-header">
            <span>邮件模板列表</span>
            <el-button type="primary" size="small" @click="loadEmailTemplates">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>

          <el-table :data="emailTemplates" v-loading="emailTemplatesLoading" stripe>
            <el-table-column prop="type" label="类型" width="150">
              <template #default="{ row }">
                <el-tag>{{ getTemplateTypeName(row.type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="subject" label="主题" />
            <el-table-column prop="enabled" label="状态" width="80">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="180">
              <template #default="{ row }">
                <el-button link type="primary" @click="openEditTemplateDialog(row)">编辑</el-button>
                <el-button link type="success" @click="openTestTemplateDialog(row)">测试</el-button>
              </template>
            </el-table-column>
          </el-table>

          <!-- 编辑模板对话框 -->
          <el-dialog v-model="templateDialogVisible" title="编辑邮件模板" width="600px">
            <el-form :model="templateForm" label-width="80px">
              <el-form-item label="类型">
                <el-input v-model="templateForm.type" disabled />
              </el-form-item>
              <el-form-item label="主题">
                <el-input v-model="templateForm.subject" />
              </el-form-item>
              <el-form-item label="启用">
                <el-switch v-model="templateForm.enabled" />
              </el-form-item>
              <el-form-item label="内容">
                <el-input v-model="templateForm.body" type="textarea" :rows="10" placeholder="支持 HTML 格式，可使用模板变量" />
              </el-form-item>
            </el-form>
            <template #footer>
              <el-button @click="templateDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="saveTemplate" :loading="templateSaving">保存</el-button>
            </template>
          </el-dialog>

          <!-- 测试模板对话框 -->
          <el-dialog v-model="testDialogVisible" title="测试邮件模板" width="400px">
            <el-form label-width="80px">
              <el-form-item label="模板">
                <el-input :value="getTemplateTypeName(testTemplateType)" disabled />
              </el-form-item>
              <el-form-item label="收件人">
                <el-input v-model="testEmail" placeholder="请输入测试邮箱" />
              </el-form-item>
            </el-form>
            <template #footer>
              <el-button @click="testDialogVisible = false">取消</el-button>
              <el-button type="primary" @click="sendTestEmail" :loading="testSending">发送</el-button>
            </template>
          </el-dialog>
        </el-tab-pane>

        <!-- 存储设置 -->
        <el-tab-pane label="存储设置" name="storage">
          <el-form label-width="120px" style="max-width: 600px;">
            <el-form-item label="存储方式">
              <el-radio-group v-model="settings.storageType">
                <el-radio value="local">本地存储</el-radio>
                <el-radio value="s3">S3 存储</el-radio>
              </el-radio-group>
            </el-form-item>

            <template v-if="settings.storageType === 'local'">
              <el-form-item label="存储路径">
                <el-input v-model="settings.uploadPath" />
              </el-form-item>
              <el-form-item label="最大文件大小">
                <el-input-number v-model="settings.maxFileSize" :min="1" :max="100" />
                <span style="margin-left: 8px;">MB</span>
              </el-form-item>
            </template>

            <template v-else>
              <el-form-item label="S3 Endpoint">
                <el-input v-model="settings.s3Endpoint" placeholder="s3.amazonaws.com" />
              </el-form-item>
              <el-form-item label="Bucket">
                <el-input v-model="settings.s3Bucket" />
              </el-form-item>
              <el-form-item label="Access Key">
                <el-input v-model="settings.s3AccessKey" />
              </el-form-item>
              <el-form-item label="Secret Key">
                <el-input v-model="settings.s3SecretKey" type="password" show-password />
              </el-form-item>
              <el-form-item label="Region">
                <el-input v-model="settings.s3Region" placeholder="us-east-1" />
              </el-form-item>
            </template>

            <el-form-item>
              <el-button type="primary">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 安全设置 -->
        <el-tab-pane label="安全" name="security">
          <el-form label-width="150px" style="max-width: 600px;">
            <el-divider>速率限制</el-divider>
            <el-form-item label="启用速率限制">
              <el-switch v-model="settings.rateLimitEnabled" />
            </el-form-item>
            <el-form-item label="每分钟请求数" v-if="settings.rateLimitEnabled">
              <el-input-number v-model="settings.rateLimitPerMinute" :min="1" :max="1000" />
            </el-form-item>
            <el-form-item label="突发容量" v-if="settings.rateLimitEnabled">
              <el-input-number v-model="settings.rateLimitBurst" :min="1" :max="100" />
            </el-form-item>

            <el-divider>CORS</el-divider>
            <el-form-item label="允许的来源">
              <el-input v-model="settings.corsOrigins" placeholder="* 或具体域名，用逗号分隔" />
            </el-form-item>
            <el-form-item label="允许的方法">
              <el-checkbox-group v-model="settings.corsMethods">
                <el-checkbox label="GET" />
                <el-checkbox label="POST" />
                <el-checkbox label="PUT" />
                <el-checkbox label="PATCH" />
                <el-checkbox label="DELETE" />
              </el-checkbox-group>
            </el-form-item>

            <el-form-item>
              <el-button type="primary">保存设置</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, Folder } from '@element-plus/icons-vue'
import { getAdmins, createAdmin, updateAdmin, deleteAdmin } from '@/api/admin'
import { getEmailTemplates, updateEmailTemplate, testEmailTemplate } from '@/api/logs'

const router = useRouter()
const activeTab = ref('email')

function goToDictionaries() {
  router.push('/dictionaries')
}

// 管理员管理
const admins = ref([])
const adminsLoading = ref(false)
const adminDialogVisible = ref(false)
const adminDialogTitle = ref('添加管理员')
const adminSaving = ref(false)
const adminForm = reactive({
  id: '',
  email: '',
  password: ''
})

async function loadAdmins() {
  adminsLoading.value = true
  try {
    const res = await getAdmins({ page: 1, perPage: 100 })
    admins.value = res.data.items || []
  } catch (e) {
    console.error('加载管理员失败:', e)
  } finally {
    adminsLoading.value = false
  }
}

function openAddAdminDialog() {
  adminDialogTitle.value = '添加管理员'
  adminForm.id = ''
  adminForm.email = ''
  adminForm.password = ''
  adminDialogVisible.value = true
}

function openEditAdminDialog(row) {
  adminDialogTitle.value = '编辑管理员'
  adminForm.id = row.id
  adminForm.email = row.email
  adminForm.password = ''
  adminDialogVisible.value = true
}

async function saveAdmin() {
  if (!adminForm.email) {
    ElMessage.warning('请输入邮箱')
    return
  }
  if (!adminForm.id && !adminForm.password) {
    ElMessage.warning('请输入密码')
    return
  }

  adminSaving.value = true
  try {
    if (adminForm.id) {
      await updateAdmin(adminForm.id, { email: adminForm.email })
      ElMessage.success('更新成功')
    } else {
      await createAdmin({ email: adminForm.email, password: adminForm.password })
      ElMessage.success('创建成功')
    }
    adminDialogVisible.value = false
    loadAdmins()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    adminSaving.value = false
  }
}

async function deleteAdminConfirm(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除管理员 "${row.email}" 吗？`,
      '警告',
      { type: 'warning' }
    )
    await deleteAdmin(row.id)
    ElMessage.success('删除成功')
    loadAdmins()
  } catch (e) {
    if (e !== 'cancel') {
      ElMessage.error(e.message || '删除失败')
    }
  }
}

// 邮件模板
const emailTemplates = ref([])
const emailTemplatesLoading = ref(false)
const templateDialogVisible = ref(false)
const templateSaving = ref(false)
const templateForm = reactive({
  type: '',
  subject: '',
  body: '',
  enabled: true
})
const testDialogVisible = ref(false)
const testTemplateType = ref('')
const testEmail = ref('')
const testSending = ref(false)

const templateTypeMap = {
  'verification': '邮箱验证',
  'password-reset': '密码重置',
  'email-change': '邮箱变更',
  'welcome': '欢迎邮件'
}

function getTemplateTypeName(type) {
  return templateTypeMap[type] || type
}

async function loadEmailTemplates() {
  emailTemplatesLoading.value = true
  try {
    const res = await getEmailTemplates()
    emailTemplates.value = res.data || []
  } catch (e) {
    console.error('加载邮件模板失败:', e)
  } finally {
    emailTemplatesLoading.value = false
  }
}

function openEditTemplateDialog(row) {
  templateForm.type = row.type
  templateForm.subject = row.subject
  templateForm.body = row.body
  templateForm.enabled = row.enabled
  templateDialogVisible.value = true
}

async function saveTemplate() {
  templateSaving.value = true
  try {
    await updateEmailTemplate(templateForm.type, {
      subject: templateForm.subject,
      body: templateForm.body,
      enabled: templateForm.enabled
    })
    ElMessage.success('保存成功')
    templateDialogVisible.value = false
    loadEmailTemplates()
  } catch (e) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    templateSaving.value = false
  }
}

function openTestTemplateDialog(row) {
  testTemplateType.value = row.type
  testEmail.value = ''
  testDialogVisible.value = true
}

async function sendTestEmail() {
  if (!testEmail.value) {
    ElMessage.warning('请输入测试邮箱')
    return
  }
  testSending.value = true
  try {
    await testEmailTemplate(testTemplateType.value, testEmail.value)
    ElMessage.success('测试邮件已发送')
    testDialogVisible.value = false
  } catch (e) {
    ElMessage.error(e.message || '发送失败')
  } finally {
    testSending.value = false
  }
}

const settings = reactive({
  appName: 'Go Gin API Admin',
  appMode: 'debug',
  appPort: 8090,
  mailEnabled: false,
  smtpHost: '',
  smtpPort: 587,
  smtpUser: '',
  smtpPassword: '',
  smtpName: '',
  storageType: 'local',
  uploadPath: './uploads',
  maxFileSize: 10,
  s3Endpoint: '',
  s3Bucket: '',
  s3AccessKey: '',
  s3SecretKey: '',
  s3Region: '',
  rateLimitEnabled: true,
  rateLimitPerMinute: 60,
  rateLimitBurst: 10,
  corsOrigins: '*',
  corsMethods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
})

onMounted(() => {
  loadAdmins()
  loadEmailTemplates()
})
</script>

<style lang="scss" scoped>
.settings-page {
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }

  .log-filter {
    margin-bottom: 16px;
  }
}
</style>

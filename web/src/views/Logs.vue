<template>
  <div class="logs-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>操作日志</span>
          <el-button size="small" @click="loadLogs">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-form :inline="true" class="log-filter">
        <el-form-item label="操作类型">
          <el-select v-model="logFilters.action" placeholder="全部" clearable style="width: 120px;">
            <el-option label="全部" value="" />
            <el-option label="登录" value="login" />
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="读取" value="read" />
          </el-select>
        </el-form-item>
        <el-form-item label="用户类型">
          <el-select v-model="logFilters.userType" placeholder="全部" clearable style="width: 100px;">
            <el-option label="全部" value="" />
            <el-option label="管理员" value="admin" />
            <el-option label="用户" value="user" />
          </el-select>
        </el-form-item>
        <el-form-item label="集合">
          <el-input v-model="logFilters.collection" placeholder="集合名称" clearable style="width: 120px;" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadLogs">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="logs" v-loading="logsLoading" stripe>
        <el-table-column prop="action" label="操作" width="80">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)" size="small">
              {{ row.action }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="collection" label="集合" width="120" />
        <el-table-column prop="userEmail" label="用户" width="180">
          <template #default="{ row }">
            {{ row.userEmail || row.userId || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="userType" label="类型" width="70">
          <template #default="{ row }">
            <el-tag :type="row.userType === 'admin' ? 'warning' : 'info'" size="small">
              {{ row.userType === 'admin' ? '管理员' : '用户' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="ip" label="IP" width="130" />
        <el-table-column prop="status" label="状态" width="70">
          <template #default="{ row }">
            <el-tag :type="row.status < 400 ? 'success' : 'danger'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created" label="时间" width="170" />
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-button link type="primary" @click="showLogDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-if="logTotal > 0"
        v-model:current-page="logPage"
        v-model:page-size="logPerPage"
        :page-sizes="[10, 20, 50, 100]"
        :total="logTotal"
        layout="total, sizes, prev, pager, next"
        @size-change="loadLogs"
        @current-change="loadLogs"
        style="margin-top: 16px; justify-content: flex-end;"
      />
    </el-card>

    <!-- 日志详情对话框 -->
    <el-dialog v-model="detailDialogVisible" title="日志详情" width="600px">
      <el-descriptions :column="2" border v-if="currentLog">
        <el-descriptions-item label="操作类型">
          <el-tag :type="getActionType(currentLog.action)" size="small">{{ currentLog.action }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="currentLog.status < 400 ? 'success' : 'danger'" size="small">
            {{ currentLog.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="集合">{{ currentLog.collection || '-' }}</el-descriptions-item>
        <el-descriptions-item label="记录ID">{{ currentLog.recordId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="用户">{{ currentLog.userEmail || currentLog.userId || '-' }}</el-descriptions-item>
        <el-descriptions-item label="用户类型">
          <el-tag :type="currentLog.userType === 'admin' ? 'warning' : 'info'" size="small">
            {{ currentLog.userType === 'admin' ? '管理员' : '用户' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="IP地址">{{ currentLog.ip }}</el-descriptions-item>
        <el-descriptions-item label="时间">{{ currentLog.created }}</el-descriptions-item>
        <el-descriptions-item label="请求路径" :span="2">{{ currentLog.path }}</el-descriptions-item>
        <el-descriptions-item label="请求方法">{{ currentLog.method }}</el-descriptions-item>
      </el-descriptions>
      
      <el-divider v-if="currentLog?.requestBody || currentLog?.responseBody">详细数据</el-divider>
      
      <div v-if="currentLog?.requestBody" class="log-detail-section">
        <div class="section-label">请求体:</div>
        <pre class="code-block">{{ formatJSON(currentLog.requestBody) }}</pre>
      </div>
      
      <div v-if="currentLog?.responseBody" class="log-detail-section">
        <div class="section-label">响应体:</div>
        <pre class="code-block">{{ formatJSON(currentLog.responseBody) }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getLogs } from '@/api/logs'
import { Refresh } from '@element-plus/icons-vue'

const logs = ref([])
const logsLoading = ref(false)
const logPage = ref(1)
const logPerPage = ref(30)
const logTotal = ref(0)
const logFilters = reactive({
  action: '',
  userType: '',
  collection: ''
})

const detailDialogVisible = ref(false)
const currentLog = ref(null)

async function loadLogs() {
  logsLoading.value = true
  try {
    const params = {
      page: logPage.value,
      perPage: logPerPage.value
    }
    if (logFilters.action) params.action = logFilters.action
    if (logFilters.userType) params.userType = logFilters.userType
    if (logFilters.collection) params.collection = logFilters.collection

    const res = await getLogs(params)
    logs.value = res.data.items || []
    logTotal.value = res.data.totalItems || 0
  } catch (e) {
    console.error('加载日志失败:', e)
  } finally {
    logsLoading.value = false
  }
}

function getActionType(action) {
  const map = {
    login: 'success',
    create: 'primary',
    update: 'warning',
    delete: 'danger',
    read: 'info'
  }
  return map[action] || 'info'
}

function showLogDetail(row) {
  currentLog.value = row
  detailDialogVisible.value = true
}

function formatJSON(str) {
  if (!str) return '-'
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}

onMounted(() => {
  loadLogs()
})
</script>

<style lang="scss" scoped>
.logs-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .log-filter {
    margin-bottom: 16px;
  }

  .log-detail-section {
    margin-bottom: 12px;
    
    .section-label {
      font-weight: 500;
      margin-bottom: 8px;
      color: #606266;
    }

    .code-block {
      background: #f5f7fa;
      padding: 12px;
      border-radius: 4px;
      font-size: 12px;
      overflow-x: auto;
      max-height: 200px;
      margin: 0;
    }
  }
}
</style>

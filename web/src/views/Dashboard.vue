<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon collections">
              <el-icon><Collection /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.collections }}</div>
              <div class="stat-label">集合数量</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon records">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.records }}</div>
              <div class="stat-label">记录数量</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon users">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.users }}</div>
              <div class="stat-label">管理员数量</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-content">
            <div class="stat-icon storage">
              <el-icon><Folder /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.storage }}</div>
              <div class="stat-label">存储空间</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="16">
        <el-card shadow="hover">
          <template #header>
            <span>快速操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="goTo('/collections')">
              <el-icon><Plus /></el-icon>
              创建集合
            </el-button>
            <el-button @click="goTo('/settings')">
              <el-icon><Setting /></el-icon>
              系统设置
            </el-button>
            <el-button @click="openDoc">
              <el-icon><Document /></el-icon>
              API 文档
            </el-button>
          </div>
        </el-card>
      </el-col>

      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>
            <span>系统信息</span>
          </template>
          <el-descriptions :column="1" size="small">
            <el-descriptions-item label="版本">1.0.0</el-descriptions-item>
            <el-descriptions-item label="Go版本">1.21+</el-descriptions-item>
            <el-descriptions-item label="数据库">MySQL 8.0</el-descriptions-item>
            <el-descriptions-item label="运行时间">{{ uptime }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getStats } from '@/api/admin'

const router = useRouter()

const stats = reactive({
  collections: 0,
  records: 0,
  users: 0,
  storage: '0 MB'
})

const uptime = ref('--')

function goTo(path) {
  router.push(path)
}

function openDoc() {
  window.open('/documentation', '_blank')
}

async function loadStats() {
  try {
    const res = await getStats()
    stats.collections = res.data.collections || 0
    stats.records = res.data.records || 0
    stats.users = res.data.users || 0
    stats.storage = res.data.storage || '0 MB'
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

onMounted(() => {
  loadStats()
})
</script>

<style lang="scss" scoped>
.dashboard {
  .stat-card {
    .stat-content {
      display: flex;
      align-items: center;
    }
    
    .stat-icon {
      width: 56px;
      height: 56px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 24px;
      color: #fff;
      margin-right: 16px;
      
      &.collections {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }
      
      &.records {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }
      
      &.users {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
      }
      
      &.storage {
        background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
      }
    }
    
    .stat-info {
      .stat-value {
        font-size: 28px;
        font-weight: bold;
        color: #333;
      }
      
      .stat-label {
        font-size: 14px;
        color: #999;
        margin-top: 4px;
      }
    }
  }
  
  .quick-actions {
    display: flex;
    gap: 12px;
  }
}
</style>

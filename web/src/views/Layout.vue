<template>
  <el-container class="layout-container">
    <!-- 侧边栏 -->
    <el-aside :width="isCollapse ? '64px' : '240px'" class="layout-aside">
      <div class="logo">
        <div class="logo-icon">
          <el-icon :size="28"><Grid /></el-icon>
        </div>
        <transition name="fade">
          <span v-show="!isCollapse" class="logo-text">API Admin</span>
        </transition>
      </div>

      <el-scrollbar wrap-class="menu-scroll">
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          class="layout-menu"
          background-color="#1a1a2e"
          text-color="#a6a8b1"
          active-text-color="#409eff"
        >
          <!-- 仪表盘 -->
          <el-menu-item index="/dashboard">
            <el-icon><DataBoard /></el-icon>
            <template #title>仪表盘</template>
          </el-menu-item>


          <!-- 快速入口 - 集合列表 -->
          <template v-if="collections.length > 0">
            <template v-for="collection in collections" :key="collection.id">
              <!-- 快速入口直接跳转到记录列表 -->
              <el-menu-item
                v-if="!collection.menuHidden && collection.type !== 'view' && collection.type !== 'transaction'"
                :index="'/collections/' + collection.name"
              >
                <!-- 折叠状态：显示tooltip -->
                <template v-if="isCollapse">
                  <el-tooltip :content="collection.label || collection.name" placement="right">
                    <el-icon><component :is="getCollectionIcon(collection)" /></el-icon>
                  </el-tooltip>
                </template>
                <!-- 展开状态：显示完整内容 -->
                <template v-else>
                  <el-icon><component :is="getCollectionIcon(collection)" /></el-icon>
                  <span>{{ collection.label || collection.name }}</span>
                  <el-tag v-if="collection.type !== 'base'" size="small" :type="collection.type === 'auth' ? 'success' : 'warning'" class="type-tag">
                    {{ collection.type }}
                  </el-tag>
                </template>
              </el-menu-item>
            </template>
          </template>

          <el-divider class="menu-divider"></el-divider>


          <!-- 集合管理 -->
          <el-menu-item index="/collections">
            <el-icon><Folder /></el-icon>
            <template #title>集合管理</template>
          </el-menu-item>

          <!-- 字典管理 -->
          <el-menu-item index="/dictionaries">
            <el-icon><Collection /></el-icon>
            <template #title>字典管理</template>
          </el-menu-item>

          <!-- 管理员 -->
          <el-menu-item index="/admins">
            <el-icon><User /></el-icon>
            <template #title>管理员</template>
          </el-menu-item>

          <!-- 操作日志 -->
          <el-menu-item index="/logs">
            <el-icon><Tickets /></el-icon>
            <template #title>操作日志</template>
          </el-menu-item>

          <!-- 系统设置 -->
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <template #title>系统设置</template>
          </el-menu-item>

        </el-menu>
      </el-scrollbar>

      <!-- 底部折叠按钮 -->
      <div class="aside-footer" @click="isCollapse = !isCollapse">
        <el-icon v-if="isCollapse"><Expand /></el-icon>
        <el-icon v-else><Fold /></el-icon>
        <span v-if="!isCollapse">收起</span>
      </div>
    </el-aside>

    <!-- 主内容区 -->
    <el-container>
      <!-- 顶部导航 -->
      <el-header class="layout-header">
        <div class="header-left">
          <el-breadcrumb separator="/">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="$route.meta.title">
              {{ $route.meta.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
        </div>

        <div class="header-right">
          <el-tooltip content="刷新" placement="bottom">
            <el-icon class="header-icon" @click="refreshPage"><Refresh /></el-icon>
          </el-tooltip>
          <el-dropdown trigger="click" @command="handleCommand">
            <span class="user-info">
              <el-avatar :size="32" :icon="User" />
              <span class="user-name">{{ authStore.admin?.email || '管理员' }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="password">
                  <el-icon><Key /></el-icon>修改密码
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 内容 -->
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>

    <!-- 修改密码对话框 -->
    <el-dialog v-model="passwordDialogVisible" title="修改密码" width="400px">
      <el-form :model="passwordForm" label-width="100px">
        <el-form-item label="当前密码">
          <el-input v-model="passwordForm.oldPassword" type="password" show-password placeholder="请输入当前密码" />
        </el-form-item>
        <el-form-item label="新密码">
          <el-input v-model="passwordForm.newPassword" type="password" show-password placeholder="请输入新密码" />
        </el-form-item>
        <el-form-item label="确认密码">
          <el-input v-model="passwordForm.confirmPassword" type="password" show-password placeholder="请再次输入新密码" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="savePassword" :loading="passwordSaving">确定</el-button>
      </template>
    </el-dialog>
  </el-container>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { getCollections } from '@/api/collection'
import { updateAdminPassword } from '@/api/admin'
import {
  Grid, DataBoard, Folder, Document, Lock, View, Setting, Collection, Tickets, Key,
  Fold, Expand, ArrowDown, Refresh, User, SwitchButton, Calendar, ChatDotRound, Star,
  Location, Phone, Message, ShoppingCart, Bell, Timer,
  Edit, Delete, Search, Upload, Download,
  Picture, VideoCamera, Headset, Wallet, TrendCharts
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isCollapse = ref(false)
const collections = ref([])

const activeMenu = computed(() => {
  return route.path
})

// 图标映射
const iconMap = {
  Document, Folder, User, Setting, Lock, View, Grid, DataBoard,
  Collection, Tickets, Key, Calendar, ChatDotRound, Star,
  Location, Phone, Message, ShoppingCart, Bell, Timer,
  Edit, Delete, Search, Upload, Download,
  Picture, VideoCamera, Headset, Wallet, TrendCharts
}

// 获取集合图标
function getCollectionIcon(collection) {
  if (collection.icon && iconMap[collection.icon]) {
    return iconMap[collection.icon]
  }
  // 默认图标
  if (collection.type === 'auth') return Lock
  if (collection.type === 'view') return View
  if (collection.type === 'transaction') return TrendCharts
  return Document
}

async function loadCollections() {
  try {
    const res = await getCollections({ page: 1, perPage: 100 })
    collections.value = res.data.items || []
  } catch (error) {
    console.error('加载集合失败:', error)
  }
}

async function handleCommand(command) {
  switch (command) {
    case 'password':
      showPasswordDialog()
      break
    case 'logout':
      await authStore.logout()
      router.push('/login')
      break
  }
}

// 修改密码对话框
const passwordDialogVisible = ref(false)
const passwordSaving = ref(false)
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

function showPasswordDialog() {
  passwordForm.oldPassword = ''
  passwordForm.newPassword = ''
  passwordForm.confirmPassword = ''
  passwordDialogVisible.value = true
}

async function savePassword() {
  if (!passwordForm.oldPassword) {
    ElMessage.warning('请输入当前密码')
    return
  }
  if (!passwordForm.newPassword) {
    ElMessage.warning('请输入新密码')
    return
  }
  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }
  if (passwordForm.newPassword.length < 6) {
    ElMessage.warning('密码长度至少6位')
    return
  }

  passwordSaving.value = true
  try {
    const res = await updateAdminPassword({
      oldPassword: passwordForm.oldPassword,
      newPassword: passwordForm.newPassword
    })
    ElMessage.success('密码修改成功')
    passwordDialogVisible.value = false
  } catch (error) {
    ElMessage.error(error.message || '密码修改失败')
  } finally {
    passwordSaving.value = false
  }
}

function refreshPage() {
  window.location.reload()
}

onMounted(() => {
  loadCollections()
})
</script>

<style lang="scss" scoped>
.layout-container {
  height: 100vh;
}

.layout-aside {
  background: linear-gradient(180deg, #1a1a2e 0%, #16213e 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
  overflow-x: hidden;

  .logo {
    height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);

    .logo-icon {
      width: 36px;
      height: 36px;
      background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      border-radius: 10px;
      display: flex;
      align-items: center;
      justify-content: center;
      color: #fff;
    }

    .logo-text {
      font-size: 18px;
      font-weight: 700;
      color: #fff;
      margin-left: 12px;
      white-space: nowrap;
    }
  }

  .menu-scroll {
    flex: 1;
    overflow: hidden;
    
    :deep(.el-scrollbar__wrap) {
      overflow-x: hidden !important;
      max-width: 100%;
    }
    
    :deep(.el-scrollbar__view) {
      max-width: 100%;
      overflow-x: hidden;
    }
  }

  .aside-footer {
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    color: #a6a8b1;
    cursor: pointer;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
    transition: all 0.3s;

    &:hover {
      background: rgba(255, 255, 255, 0.05);
      color: #fff;
    }
  }
}

.layout-menu {
  border-right: none;
  background: transparent;
  width: 100%;
  max-width: 100%;

  &:not(.el-menu--collapse) {
    width: 100%;
  }

  :deep(.el-menu-item),
  :deep(.el-sub-menu__title) {
    height: 48px;
    line-height: 48px;
    margin: 4px 8px;
    border-radius: 8px;
    transition: all 0.3s;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

    &:hover {
      background: rgba(64, 158, 255, 0.15) !important;
    }

    &.is-active {
      background: rgba(64, 158, 255, 0.2) !important;
      color: #409eff !important;
    }
  }

  // 折叠状态下图标居中
  &.el-menu--collapse {
    :deep(.el-menu-item),
    :deep(.el-sub-menu__title) {
      padding: 0 !important;
      justify-content: center;
      
      .el-icon {
        margin-right: 0 !important;
      }
    }
  }

  :deep(.el-sub-menu .el-menu-item) {
    height: 40px;
    line-height: 40px;
    margin: 2px 0;
  }

  :deep(.el-menu-item span) {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 140px;
    display: inline-block;
  }

  .type-tag {
    margin-left: 8px;
    font-size: 10px;
    padding: 0 6px;
    flex-shrink: 0;
  }

  .menu-divider {
    margin: 16px 16px 8px;
    border-color: rgba(255, 255, 255, 0.08);
  }
}

.layout-header {
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  z-index: 10;

  .header-left {
    display: flex;
    align-items: center;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .header-icon {
      font-size: 18px;
      color: #666;
      cursor: pointer;
      padding: 8px;
      border-radius: 6px;
      transition: all 0.3s;

      &:hover {
        background: #f5f7fa;
        color: #409eff;
      }
    }
  }
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 6px 12px;
  border-radius: 8px;
  transition: all 0.3s;

  &:hover {
    background: #f5f7fa;
  }

  .user-name {
    color: #333;
    font-weight: 500;
  }
}

.layout-main {
  background: #f0f2f5;
  padding: 20px 24px;
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

<template>
  <div class="admins-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>管理员列表</span>
          <el-button type="primary" @click="openAddDialog">
            <el-icon><Plus /></el-icon>
            添加管理员
          </el-button>
        </div>
      </template>

      <el-table :data="admins" v-loading="loading" stripe>
        <el-table-column prop="email" label="邮箱" />
        <el-table-column prop="created" label="创建时间" width="180" />
        <el-table-column label="操作" width="150">
          <template #default="{ row }">
            <el-button link type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-button link type="danger" @click="deleteAdminConfirm(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.perPage"
          :total="pagination.total"
          :page-sizes="[10, 20, 30, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadAdmins"
          @current-change="loadAdmins"
        />
      </div>
    </el-card>

    <!-- 添加/编辑管理员对话框 -->
    <el-dialog v-model="dialogVisible" :title="dialogTitle" width="400px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" v-if="!form.id">
          <el-input v-model="form.password" type="password" placeholder="请输入密码（至少6位）" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveAdmin" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdmins, createAdmin, updateAdmin, deleteAdmin } from '@/api/admin'

const loading = ref(false)
const admins = ref([])
const dialogVisible = ref(false)
const dialogTitle = ref('添加管理员')
const saving = ref(false)

const pagination = reactive({
  page: 1,
  perPage: 30,
  total: 0
})

const form = reactive({
  id: '',
  email: '',
  password: ''
})

async function loadAdmins() {
  loading.value = true
  try {
    const res = await getAdmins({
      page: pagination.page,
      perPage: pagination.perPage
    })
    admins.value = res.data.items || []
    pagination.total = res.data.totalItems || 0
  } catch (e) {
    console.error('加载管理员失败:', e)
  } finally {
    loading.value = false
  }
}

function openAddDialog() {
  dialogTitle.value = '添加管理员'
  form.id = ''
  form.email = ''
  form.password = ''
  dialogVisible.value = true
}

function openEditDialog(row) {
  dialogTitle.value = '编辑管理员'
  form.id = row.id
  form.email = row.email
  form.password = ''
  dialogVisible.value = true
}

async function saveAdmin() {
  if (!form.email) {
    ElMessage.warning('请输入邮箱')
    return
  }
  if (!form.id && !form.password) {
    ElMessage.warning('请输入密码')
    return
  }

  saving.value = true
  try {
    if (form.id) {
      await updateAdmin(form.id, { email: form.email })
      ElMessage.success('更新成功')
    } else {
      await createAdmin({ email: form.email, password: form.password })
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadAdmins()
  } catch (e) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    saving.value = false
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

onMounted(() => {
  loadAdmins()
})
</script>

<style lang="scss" scoped>
.admins-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>

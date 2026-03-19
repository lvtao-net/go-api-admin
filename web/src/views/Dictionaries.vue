<template>
  <div class="dictionaries-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>字典管理</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新建字典
          </el-button>
        </div>
      </template>

      <el-table :data="dictionaries" v-loading="loading" stripe>
        <el-table-column prop="name" label="字典名称" min-width="150" />
        <el-table-column prop="label" label="显示名称" min-width="150" />
        <el-table-column prop="description" label="描述" min-width="200" />
        <el-table-column prop="system" label="系统字典" width="100">
          <template #default="{ row }">
            <el-tag :type="row.system ? 'danger' : 'info'" size="small">
              {{ row.system ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="items" label="选项数" width="100">
          <template #default="{ row }">
            {{ row.items?.length || 0 }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="editDictionary(row)">
              编辑
            </el-button>
            <el-button link type="primary" @click="manageItems(row)">
              管理选项
            </el-button>
            <el-button link type="danger" @click="deleteDictionary(row)" :disabled="row.system">
              删除
            </el-button>
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
          @size-change="loadDictionaries"
          @current-change="loadDictionaries"
        />
      </div>
    </el-card>

    <!-- 创建/编辑字典对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑字典' : '新建字典'"
      width="600px"
    >
      <el-form :model="form" label-width="100px">
        <el-form-item label="字典名称" required>
          <el-input v-model="form.name" :disabled="isEdit" placeholder="如: gender, status" />
        </el-form-item>
        <el-form-item label="显示名称">
          <el-input v-model="form.label" placeholder="如: 性别、状态" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="saveDictionary" :loading="saving">
          保存
        </el-button>
      </template>
    </el-dialog>

    <!-- 字典项管理对话框 -->
    <el-dialog
      v-model="itemsDialogVisible"
      title="管理字典选项"
      width="800px"
    >
      <div class="items-header">
        <el-button type="primary" size="small" @click="showAddItemDialog">
          <el-icon><Plus /></el-icon>
          添加选项
        </el-button>
      </div>

      <el-table :data="currentItems" stripe>
        <el-table-column prop="label" label="显示名称" min-width="120">
          <template #default="{ row, $index }">
            <el-input v-if="row.editing" v-model="row.editLabel" size="small" />
            <span v-else>{{ row.label }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="值" min-width="120">
          <template #default="{ row, $index }">
            <el-input v-if="row.editing" v-model="row.editValue" size="small" />
            <span v-else>{{ row.value }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="sort" label="排序" width="80">
          <template #default="{ row, $index }">
            <el-input-number
              v-if="row.editing"
              v-model="row.editSort"
              size="small"
              :min="0"
              controls-position="right"
            />
            <span v-else>{{ row.sort }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="disabled" label="禁用" width="80">
          <template #default="{ row }">
            <el-switch v-model="row.disabled" :disabled="row.editing" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row, $index }">
            <template v-if="row.editing">
              <el-button link type="primary" size="small" @click="saveItem(row, $index)">
                保存
              </el-button>
              <el-button link size="small" @click="cancelEdit(row, $index)">
                取消
              </el-button>
            </template>
            <template v-else>
              <el-button link type="primary" size="small" @click="editItem(row, $index)">
                编辑
              </el-button>
              <el-button link type="danger" size="small" @click="deleteItem(row, $index)">
                删除
              </el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>

      <template #footer>
        <el-button @click="itemsDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 添加字典项对话框 -->
    <el-dialog
      v-model="addItemDialogVisible"
      title="添加字典选项"
      width="400px"
    >
      <el-form :model="itemForm" label-width="80px">
        <el-form-item label="显示名称" required>
          <el-input v-model="itemForm.label" placeholder="如: 男" />
        </el-form-item>
        <el-form-item label="值" required>
          <el-input v-model="itemForm.value" placeholder="如: male" />
        </el-form-item>
        <el-form-item label="排序">
          <el-input-number v-model="itemForm.sort" :min="0" />
        </el-form-item>
        <el-form-item label="禁用">
          <el-switch v-model="itemForm.disabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="addItemDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="addItem" :loading="addingItem">添加</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getDictionaries,
  createDictionary,
  updateDictionary,
  deleteDictionary as deleteDict,
  getDictionaryItems,
  createDictionaryItem,
  updateDictionaryItem,
  deleteDictionaryItem
} from '@/api/dictionary'

const loading = ref(false)
const saving = ref(false)
const addingItem = ref(false)
const dictionaries = ref([])

const pagination = reactive({
  page: 1,
  perPage: 30,
  total: 0
})

const dialogVisible = ref(false)
const isEdit = ref(false)
const form = reactive({
  name: '',
  label: '',
  description: ''
})

const itemsDialogVisible = ref(false)
const addItemDialogVisible = ref(false)
const currentDictId = ref('')
const currentItems = ref([])
const itemForm = reactive({
  label: '',
  value: '',
  sort: 0,
  disabled: false
})

async function loadDictionaries() {
  loading.value = true
  try {
    const res = await getDictionaries({
      page: pagination.page,
      perPage: pagination.perPage
    })
    dictionaries.value = res.data.items || []
    pagination.total = res.data.totalItems || 0
  } catch (error) {
    ElMessage.error('加载字典失败')
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEdit.value = false
  form.name = ''
  form.label = ''
  form.description = ''
  dialogVisible.value = true
}

function editDictionary(row) {
  isEdit.value = true
  form.name = row.name
  form.label = row.label
  form.description = row.description
  form.id = row.id
  dialogVisible.value = true
}

async function saveDictionary() {
  if (!form.name) {
    ElMessage.warning('请输入字典名称')
    return
  }

  saving.value = true
  try {
    if (isEdit.value) {
      await updateDictionary(form.id, {
        label: form.label,
        description: form.description
      })
      ElMessage.success('更新成功')
    } else {
      await createDictionary(form)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadDictionaries()
  } catch (error) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function deleteDictionary(row) {
  try {
    await ElMessageBox.confirm(
      `确定要删除字典"${row.label || row.name}"吗？`,
      '警告',
      { type: 'warning' }
    )
    await deleteDict(row.id)
    ElMessage.success('删除成功')
    loadDictionaries()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function manageItems(row) {
  currentDictId.value = row.id
  currentItems.value = row.items || []
  itemsDialogVisible.value = true
}

function showAddItemDialog() {
  itemForm.label = ''
  itemForm.value = ''
  itemForm.sort = currentItems.value.length
  itemForm.disabled = false
  addItemDialogVisible.value = true
}

async function addItem() {
  if (!itemForm.label || !itemForm.value) {
    ElMessage.warning('请填写显示名称和值')
    return
  }

  addingItem.value = true
  try {
    const item = await createDictionaryItem(currentDictId.value, itemForm)
    currentItems.value.push(item.data)
    addItemDialogVisible.value = false
    ElMessage.success('添加成功')
  } catch (error) {
    ElMessage.error(error.message || '添加失败')
  } finally {
    addingItem.value = false
  }
}

function editItem(row, index) {
  row.editing = true
  row.editLabel = row.label
  row.editValue = row.value
  row.editSort = row.sort
}

async function saveItem(row, index) {
  try {
    await updateDictionaryItem(currentDictId.value, row.id, {
      label: row.editLabel,
      value: row.editValue,
      sort: row.editSort,
      disabled: row.disabled
    })
    row.label = row.editLabel
    row.value = row.editValue
    row.sort = row.editSort
    row.editing = false
    ElMessage.success('保存成功')
  } catch (error) {
    ElMessage.error(error.message || '保存失败')
  }
}

function cancelEdit(row, index) {
  row.editing = false
}

async function deleteItem(row, index) {
  try {
    await ElMessageBox.confirm(
      `确定要删除选项"${row.label}"吗？`,
      '警告',
      { type: 'warning' }
    )
    await deleteDictionaryItem(currentDictId.value, row.id)
    currentItems.value.splice(index, 1)
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

onMounted(() => {
  loadDictionaries()
})
</script>

<style lang="scss" scoped>
.dictionaries-page {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .pagination-container {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }

  .items-header {
    margin-bottom: 16px;
  }
}
</style>

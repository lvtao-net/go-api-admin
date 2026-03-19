<template>
  <div class="record-list-page">
    <div class="page-header">
      <h2>
        <el-icon v-if="collectionType === 'auth'"><Lock /></el-icon>
        <el-icon v-else-if="collectionType === 'view'"><View /></el-icon>
        <el-icon v-else-if="collectionType === 'transaction'"><TrendCharts /></el-icon>
        <el-icon v-else><Document /></el-icon>
        {{ collectionLabel || collectionName }}
        <el-tag size="small" style="margin-left: 8px;">{{ collectionType }}</el-tag>
      </h2>
      <div class="header-actions" v-if="!isNoDataCollection">
        <el-button type="primary" @click="showCreateDialog">
          <el-icon><Plus /></el-icon>
          新建
        </el-button>
        <el-button @click="loadRecords">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 视图/事务集合提示 -->
    <el-alert
      v-if="isNoDataCollection"
      :title="collectionType === 'view' ? '视图集合' : '事务集合'"
      :type="collectionType === 'view' ? 'warning' : 'error'"
      :closable="false"
      show-icon
      style="margin-bottom: 20px;"
    >
      <template #default>
        <div v-if="collectionType === 'view'">
          视图集合是基于 SQL 查询的只读虚拟集合，没有实际数据表，无需管理记录。
          请在集合管理中进行配置。
        </div>
        <div v-else>
          事务集合用于执行多步骤的原子操作，没有实际数据表，无需管理记录。
          请通过 API 调用事务或前往集合管理进行配置。
        </div>
        <el-button type="primary" size="small" style="margin-top: 12px;" @click="$router.push('/collections')">
          前往集合管理
        </el-button>
      </template>
    </el-alert>

    <!-- 搜索过滤 -->
    <div class="filter-bar" v-if="!isNoDataCollection">
      <el-input
        v-model="searchText"
        placeholder="搜索..."
        clearable
        style="width: 300px;"
        @keyup.enter="loadRecords"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>

    <!-- 数据表格 -->
    <el-table v-if="!isNoDataCollection" :data="records" v-loading="loading" stripe border>
      <el-table-column prop="id" label="ID" width="150" />
      <el-table-column
        v-for="col in displayColumns"
        :key="col.name"
        :prop="col.name"
        :label="col.label || col.name"
        :min-width="col.width || 120"
      >
        <template #default="{ row }">
          <template v-if="col.type === 'bool'">
            <el-tag :type="row[col.name] ? 'success' : 'info'" size="small">
              {{ row[col.name] ? '是' : '否' }}
            </el-tag>
          </template>
          <template v-else-if="col.type === 'date'">
            {{ formatDate(row[col.name]) }}
          </template>
          <template v-else-if="col.type === 'relation'">
            <el-tag v-if="row[col.name]" size="small">{{ row[col.name] }}</el-tag>
            <span v-else>-</span>
          </template>
          <template v-else-if="col.type === 'file'">
            <el-link v-if="row[col.name]" type="primary" :href="row[col.name]" target="_blank">
              查看文件
            </el-link>
            <span v-else>-</span>
          </template>
          <template v-else>
            {{ truncateText(row[col.name], 50) }}
          </template>
        </template>
      </el-table-column>
      <el-table-column label="创建时间" width="160">
        <template #default="{ row }">
          {{ formatDate(row.created) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="viewRecord(row)">详情</el-button>
          <el-button link type="primary" @click="editRecord(row)">编辑</el-button>
          <el-button link type="danger" @click="deleteRecordConfirm(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container" v-if="!isNoDataCollection">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.perPage"
        :total="pagination.total"
        :page-sizes="[10, 20, 30, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadRecords"
        @current-change="loadRecords"
      />
    </div>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑记录' : '新建记录'"
      width="700px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="formData" label-width="120px">
        <el-form-item
          v-for="field in editableFields"
          :key="field.name"
          :label="field.label || field.name"
          :required="field.required"
        >
          <!-- 富文本编辑器 (displayType: rich_text) -->
          <div
            v-if="field.displayType === 'rich_text' || field.type === 'editor'"
            class="rich-text-editor"
          >
            <el-input
              v-model="formData[field.name]"
              type="textarea"
              :rows="10"
              :placeholder="field.placeholder || `请输入${field.label || field.name}`"
            />
            <div class="editor-preview" v-if="formData[field.name]" v-html="formData[field.name]"></div>
          </div>
          <!-- 图片上传 (displayType: image) -->
          <el-upload
            v-else-if="field.displayType === 'image' || (field.type === 'file' && field.accept?.includes('image'))"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :on-success="(res) => handleUploadSuccess(res, field.name)"
            :show-file-list="false"
            :accept="field.accept || 'image/*'"
            :limit="field.maxCount || 1"
            list-type="picture-card"
          >
            <img v-if="formData[field.name]" :src="formData[field.name]" class="uploaded-image" />
            <el-icon v-else><Plus /></el-icon>
          </el-upload>
          <!-- 文件上传 (displayType: upload) -->
          <el-upload
            v-else-if="field.displayType === 'upload' || field.type === 'file'"
            :action="uploadUrl"
            :headers="uploadHeaders"
            :on-success="(res) => handleUploadSuccess(res, field.name)"
            :show-file-list="true"
            :accept="field.accept"
            :limit="field.maxCount || 1"
            drag
          >
            <el-icon class="el-icon--upload"><Upload /></el-icon>
            <div class="el-upload__text">拖拽文件到此处或 <em>点击上传</em></div>
          </el-upload>
          <!-- 日期选择器 (displayType: date) -->
          <el-date-picker
            v-else-if="field.displayType === 'date' || field.type === 'date'"
            v-model="formData[field.name]"
            type="date"
            :placeholder="field.placeholder || '选择日期'"
            style="width: 100%;"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
          />
          <!-- 日期时间选择器 (displayType: datetime) -->
          <el-date-picker
            v-else-if="field.displayType === 'datetime'"
            v-model="formData[field.name]"
            type="datetime"
            :placeholder="field.placeholder || '选择日期时间'"
            style="width: 100%;"
          />
          <!-- 开关 (displayType: switch) -->
          <el-switch
            v-else-if="field.displayType === 'switch' || field.type === 'bool'"
            v-model="formData[field.name]"
            :active-text="field.options?.activeText || '是'"
            :inactive-text="field.options?.inactiveText || '否'"
          />
          <!-- 滑块 (displayType: slider) -->
          <el-slider
            v-else-if="field.displayType === 'slider'"
            v-model="formData[field.name]"
            :min="field.minValue || 0"
            :max="field.maxValue || 100"
            :step="parseFloat(field.step) || 1"
            show-input
          />
          <!-- 颜色选择器 (displayType: color) -->
          <el-color-picker
            v-else-if="field.displayType === 'color'"
            v-model="formData[field.name]"
            show-alpha
          />
          <!-- 评分 (displayType: rate) -->
          <el-rate
            v-else-if="field.displayType === 'rate'"
            v-model="formData[field.name]"
            :max="field.maxValue || 5"
            allow-half
            show-score
          />
          <!-- 密码输入 (displayType: password) -->
          <el-input
            v-else-if="field.displayType === 'password' || field.type === 'password'"
            v-model="formData[field.name]"
            type="password"
            show-password
            :placeholder="field.placeholder || `请输入${field.label || field.name}`"
          />
          <!-- 多行文本 (displayType: textarea) -->
          <el-input
            v-else-if="field.displayType === 'textarea' || field.type === 'text'"
            v-model="formData[field.name]"
            type="textarea"
            :rows="field.rows || 3"
            :placeholder="field.placeholder || `请输入${field.label || field.name}`"
            :maxlength="field.max"
            :show-word-limit="!!field.max"
          />
          <!-- JSON编辑器 (displayType: json) -->
          <el-input
            v-else-if="field.displayType === 'json' || field.type === 'json'"
            v-model="formData[field.name]"
            type="textarea"
            :rows="field.rows || 6"
            placeholder="JSON 格式"
            @blur="formatJsonField(field.name)"
          />
          <!-- 数字输入 -->
          <el-input-number
            v-else-if="field.type === 'number'"
            v-model="formData[field.name]"
            :min="field.minValue"
            :max="field.maxValue"
            :step="parseFloat(field.step) || 1"
            :precision="field.options?.precision"
            style="width: 100%;"
            :placeholder="field.placeholder"
          />
          <!-- 邮箱 -->
          <el-input
            v-else-if="field.type === 'email'"
            v-model="formData[field.name]"
            type="email"
            :placeholder="field.placeholder || `请输入${field.label || field.name}`"
          />
          <!-- URL -->
          <el-input
            v-else-if="field.type === 'url'"
            v-model="formData[field.name]"
            type="url"
            :placeholder="field.placeholder || `请输入${field.label || field.name}`"
          >
            <template #prepend>https://</template>
          </el-input>
          <!-- 单选 (radio) -->
          <el-radio-group
            v-else-if="field.type === 'radio'"
            v-model="formData[field.name]"
          >
            <el-radio
              v-for="opt in getFieldOptions(field)"
              :key="opt.value"
              :label="opt.value"
            >
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
          <!-- 多选 (checkbox) -->
          <el-checkbox-group
            v-else-if="field.type === 'checkbox'"
            v-model="formData[field.name]"
          >
            <el-checkbox
              v-for="opt in getFieldOptions(field)"
              :key="opt.value"
              :label="opt.value"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
          <!-- 下拉选择 (select) -->
          <el-select
            v-else-if="field.type === 'select'"
            v-model="formData[field.name]"
            :multiple="field.multiple || field.options?.maxSelect > 1"
            :placeholder="field.placeholder || '请选择'"
            style="width: 100%;"
            clearable
            filterable
          >
            <el-option
              v-for="opt in getFieldOptions(field)"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <!-- 关联 -->
          <el-select
            v-else-if="field.type === 'relation'"
            v-model="formData[field.name]"
            :multiple="field.relationMax > 1"
            :placeholder="field.placeholder || '请选择关联记录'"
            style="width: 100%;"
            filterable
            remote
            :remote-method="(query) => loadRelationOptions(field, query)"
            clearable
          >
            <el-option
              v-for="opt in relationOptions[field.name] || []"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
          <!-- 默认单行文本 -->
          <el-input
            v-else
            v-model="formData[field.name]"
            :placeholder="field.placeholder || `请输入${field.label || field.name}`"
            :maxlength="field.max"
            :show-word-limit="!!field.max"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="viewDialogVisible"
      title="记录详情"
      width="600px"
    >
      <el-descriptions :column="2" border v-if="currentRecord">
        <el-descriptions-item label="ID">{{ currentRecord.id }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDate(currentRecord.created) }}</el-descriptions-item>
        <el-descriptions-item label="更新时间">{{ formatDate(currentRecord.updated) }}</el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <div class="detail-fields">
        <div v-for="field in editableFields" :key="field.name" class="detail-item">
          <span class="detail-label">{{ field.label || field.name }}:</span>
          <span class="detail-value">{{ formatFieldValue(field, currentRecord?.[field.name]) }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="viewDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="editCurrentRecord">编辑</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, View, Document, Plus, Refresh, Search, Upload, TrendCharts } from '@element-plus/icons-vue'
import { getCollection, getRecords, createRecord, updateRecord, deleteRecord } from '@/api/collection'
import { getDictionaryByName } from '@/api/dictionary'

const route = useRoute()

const collectionName = computed(() => route.params.name)
const collectionType = ref('base')
const collectionLabel = ref('')
const fields = ref([])
const dictionaryCache = reactive({})

// 是否为视图或事务集合（无记录）
const isNoDataCollection = computed(() => {
  return collectionType.value === 'view' || collectionType.value === 'transaction'
})

const loading = ref(false)
const records = ref([])
const pagination = reactive({
  page: 1,
  perPage: 30,
  total: 0
})
const searchText = ref('')

// 字段选项缓存（用于 radio/checkbox/select）
const fieldOptionsCache = reactive({})

// 显示列（排除系统字段和特殊字段）
const displayColumns = computed(() => {
  return fields.value
    .filter(f => !['password', 'tokenKey', 'emailVisibility', 'verified'].includes(f.name))
    .map(f => ({
      name: f.name,
      label: f.label || f.name,
      type: f.type,
      width: f.type === 'text' || f.type === 'editor' ? 200 : 120
    }))
})

// 可编辑字段
const editableFields = computed(() => {
  return fields.value.filter(f => !['id', 'created', 'updated'].includes(f.name))
})

// 对话框
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref(null)
const formData = reactive({})
const relationOptions = reactive({})

const viewDialogVisible = ref(false)
const currentRecord = ref(null)

// 上传配置
const uploadUrl = computed(() => `/api/files`)
const uploadHeaders = computed(() => ({
  Authorization: `Bearer ${localStorage.getItem('admin_token')}`
}))

// 获取字段选项（从缓存中获取）
function getFieldOptions(field) {
  // 如果有缓存的选项
  if (fieldOptionsCache[field.name]) {
    return fieldOptionsCache[field.name]
  }
  // 手动选项
  if (field.options?.values) {
    return field.options.values.map(v => ({ label: v, value: v }))
  }
  if (field.fieldOptions) {
    return field.fieldOptions.map(o => ({ label: o.label || o.value, value: o.value }))
  }
  return []
}

// 加载字段选项（字典关联）
async function loadFieldOptions() {
  for (const field of fields.value) {
    if (['radio', 'checkbox', 'select'].includes(field.type) && field.dictionary) {
      if (dictionaryCache[field.dictionary]) {
        fieldOptionsCache[field.name] = dictionaryCache[field.dictionary]
      } else {
        try {
          const res = await getDictionaryByName(field.dictionary)
          const items = res.data?.items || []
          const options = items.map(item => ({
            label: item.label || item.value,
            value: item.value
          }))
          dictionaryCache[field.dictionary] = options
          fieldOptionsCache[field.name] = options
        } catch (e) {
          console.error('加载字典失败:', e)
        }
      }
    }
  }
}

// 加载关联字段选项
async function loadRelationOptions(field, query) {
  if (!field.relationCollection) return
  try {
    const res = await getRecords(field.relationCollection, { page: 1, perPage: 20 })
    const items = res.data?.items || []
    relationOptions[field.name] = items.map(item => ({
      label: item.name || item.title || item.id,
      value: item.id
    }))
  } catch (e) {
    console.error('加载关联选项失败:', e)
  }
}

// 上传成功
function handleUploadSuccess(res, fieldName) {
  if (res.code === 0) {
    formData[fieldName] = res.data?.url || res.data?.filename
    ElMessage.success('上传成功')
  }
}

// 显示创建对话框
function showCreateDialog() {
  isEdit.value = false
  Object.keys(formData).forEach(key => delete formData[key])
  // 设置默认值
  fields.value.forEach(f => {
    if (f.defaultValue !== undefined) {
      formData[f.name] = f.defaultValue
    } else if (f.type === 'checkbox') {
      formData[f.name] = []
    }
  })
  dialogVisible.value = true
}

// 编辑记录
function editRecord(row) {
  isEdit.value = true
  Object.keys(formData).forEach(key => delete formData[key])
  Object.assign(formData, row)
  // 处理 checkbox 类型的数据（可能是 JSON 字符串）
  fields.value.forEach(f => {
    if (f.type === 'checkbox' && typeof formData[f.name] === 'string') {
      try {
        formData[f.name] = JSON.parse(formData[f.name])
      } catch (e) {
        formData[f.name] = formData[f.name] ? [formData[f.name]] : []
      }
    }
  })
  dialogVisible.value = true
}

// 查看记录
function viewRecord(row) {
  currentRecord.value = row
  viewDialogVisible.value = true
}

// 编辑当前记录
function editCurrentRecord() {
  if (currentRecord.value) {
    viewDialogVisible.value = false
    editRecord(currentRecord.value)
  }
}

// 提交表单
async function submitForm() {
  submitting.value = true
  try {
    const data = { ...formData }
    if (isEdit.value) {
      const id = data.id
      delete data.id
      delete data.created
      delete data.updated
      await updateRecord(collectionName.value, id, data)
      ElMessage.success('更新成功')
    } else {
      await createRecord(collectionName.value, data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadRecords()
  } catch (e) {
    // handled
  } finally {
    submitting.value = false
  }
}

// 删除确认
async function deleteRecordConfirm(row) {
  await ElMessageBox.confirm('确定要删除此记录吗？', '警告', { type: 'warning' })
  try {
    await deleteRecord(collectionName.value, row.id)
    ElMessage.success('删除成功')
    loadRecords()
  } catch (e) {
    // handled
  }
}

// 加载集合信息
async function loadCollection() {
  try {
    const res = await getCollection(collectionName.value)
    collectionType.value = res.data?.type || 'base'
    collectionLabel.value = res.data?.label || ''
    fields.value = res.data?.fields || []
    // 加载字段选项（字典关联）
    await loadFieldOptions()
  } catch (e) {
    console.error('加载集合失败:', e)
  }
}

// 加载记录
async function loadRecords() {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      perPage: pagination.perPage
    }
    if (searchText.value) {
      params.filter = searchText.value
    }
    const res = await getRecords(collectionName.value, params)
    records.value = res.data?.items || []
    pagination.total = res.data?.totalItems || 0
  } catch (e) {
    console.error('加载记录失败:', e)
  } finally {
    loading.value = false
  }
}

// 格式化日期
function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

// 截断文本
function truncateText(text, maxLen) {
  if (!text) return '-'
  if (typeof text !== 'string') text = String(text)
  return text.length > maxLen ? text.slice(0, maxLen) + '...' : text
}

// 格式化字段值
function formatFieldValue(field, value) {
  if (value === null || value === undefined) return '-'
  if (field.type === 'bool') return value ? '是' : '否'
  if (field.type === 'date') return formatDate(value)
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

// 格式化JSON字段
function formatJsonField(fieldName) {
  try {
    if (formData[fieldName] && typeof formData[fieldName] === 'string') {
      const parsed = JSON.parse(formData[fieldName])
      formData[fieldName] = JSON.stringify(parsed, null, 2)
    }
  } catch (e) {
    // JSON格式错误，不处理
  }
}

// 监听路由变化
watch(collectionName, () => {
  loadCollection()
  loadRecords()
})

onMounted(() => {
  loadCollection()
  loadRecords()
})
</script>

<style lang="scss" scoped>
.record-list-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h2 {
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 0;
      font-size: 20px;
    }

    .header-actions {
      display: flex;
      gap: 10px;
    }
  }

  .filter-bar {
    margin-bottom: 16px;
  }

  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .detail-fields {
    .detail-item {
      display: flex;
      padding: 8px 0;
      border-bottom: 1px solid #eee;

      &:last-child {
        border-bottom: none;
      }

      .detail-label {
        width: 120px;
        color: #666;
        flex-shrink: 0;
      }

      .detail-value {
        flex: 1;
        word-break: break-all;
      }
    }
  }

  .rich-text-editor {
    width: 100%;

    .editor-preview {
      margin-top: 10px;
      padding: 10px;
      border: 1px solid #ddd;
      border-radius: 4px;
      background: #fafafa;
      max-height: 200px;
      overflow-y: auto;

      img {
        max-width: 100%;
      }
    }
  }

  .uploaded-image {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
}
</style>

<template>
  <div class="collections-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>集合列表</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新建集合
          </el-button>
        </div>
      </template>
      
      <el-table :data="collections" v-loading="loading" stripe>
        <el-table-column prop="name" label="集合名称" min-width="150">
          <template #default="{ row }">
            <el-link type="primary" @click="goToDetail(row.name)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="label" label="中文别名" min-width="120">
          <template #default="{ row }">
            <span>{{ row.label || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.type)">
              {{ getTypeName(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created" label="创建时间" width="180" />
        <el-table-column prop="updated" label="更新时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="goToDetail(row.name)">
              管理
            </el-button>
            <el-button link type="danger" @click="deleteCollection(row)">
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
          @size-change="loadCollections"
          @current-change="loadCollections"
        />
      </div>
    </el-card>
    
    <!-- 创建/编辑集合对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑集合' : '新建集合'"
      width="700px"
    >
      <!-- 新建集合引导 -->
      <div v-if="!isEdit" class="create-guide">
        <el-alert
          title="快速了解四种集合类型"
          type="info"
          :closable="false"
          show-icon
          style="margin-bottom: 16px;"
        />
        <div class="type-cards">
          <div
            class="type-card"
            :class="{ active: form.type === 'base' }"
            @click="form.type = 'base'"
          >
            <div class="type-icon">📦</div>
            <div class="type-title">基础集合 (Base)</div>
            <div class="type-desc">最常用的集合类型，用于存储普通业务数据</div>
            <ul class="type-features">
              <li>✓ 支持完整的 CRUD 操作</li>
              <li>✓ 可自定义字段和索引</li>
              <li>✓ 支持 API 规则控制</li>
              <li>✓ 可关联其他集合</li>
            </ul>
          </div>
          <div
            class="type-card"
            :class="{ active: form.type === 'auth' }"
            @click="form.type = 'auth'"
          >
            <div class="type-icon">🔐</div>
            <div class="type-title">认证集合 (Auth)</div>
            <div class="type-desc">用于用户认证的专用集合类型</div>
            <ul class="type-features">
              <li>✓ 内置邮箱/密码登录</li>
              <li>✓ 邮箱验证和密码重置</li>
            </ul>
          </div>
          <div
            class="type-card"
            :class="{ active: form.type === 'view' }"
            @click="form.type = 'view'"
          >
            <div class="type-icon">👁️</div>
            <div class="type-title">视图集合 (View)</div>
            <div class="type-desc">基于 SQL 查询的只读虚拟集合</div>
            <ul class="type-features">
              <li>✓ 基于自定义 SQL 创建</li>
              <li>✓ 只读，不可直接修改</li>
              <li>✓ 支持数据聚合统计</li>
              <li>✓ 可关联多个基础集合</li>
            </ul>
          </div>
          <div
            class="type-card"
            :class="{ active: form.type === 'transaction' }"
            @click="form.type = 'transaction'"
          >
            <div class="type-icon">⚡</div>
            <div class="type-title">事务集合 (Transaction)</div>
            <div class="type-desc">用于执行复杂业务事务操作</div>
            <ul class="type-features">
              <li>✓ 支持多步骤原子操作</li>
              <li>✓ 查询、验证、更新、插入</li>
              <li>✓ 条件判断与数据引用</li>
              <li>✓ 自动事务回滚</li>
            </ul>
          </div>
        </div>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        style="margin-top: 20px;"
      >
        <el-form-item label="集合名称" prop="name">
          <el-input
            v-model="form.name"
            placeholder="请输入英文表名，如：products"
            :disabled="isEdit"
          />
          <div class="form-tip">英文表名，只能包含小写字母、数字和下划线，且以字母开头</div>
        </el-form-item>

        <el-form-item label="中文别名" prop="label">
          <el-input
            v-model="form.label"
            placeholder="请输入中文别名，如：产品管理"
          />
          <div class="form-tip">中文别名将显示在左侧菜单中</div>
        </el-form-item>

        <el-form-item label="类型" prop="type" v-if="isEdit">
          <el-select v-model="form.type" placeholder="请选择类型">
            <el-option label="基础集合 (Base)" value="base" />
            <el-option label="认证集合 (Auth)" value="auth" />
            <el-option label="视图集合 (View)" value="view" />
            <el-option label="事务集合 (Transaction)" value="transaction" />
          </el-select>
        </el-form-item>

        <!-- 视图集合额外选项 -->
        <el-form-item label="SQL 查询" v-if="form.type === 'view'">
          <el-input
            v-model="form.viewSQL"
            type="textarea"
            :rows="4"
            placeholder="SELECT * FROM users WHERE status = 'active'"
          />
          <div class="form-tip">输入创建视图的 SQL 查询语句</div>
        </el-form-item>

        <!-- 事务集合额外选项 -->
        <template v-if="form.type === 'transaction'">
          <el-form-item label="事务步骤">
            <div class="transaction-steps">
              <div
                v-for="(step, index) in form.transactionSteps"
                :key="index"
                class="transaction-step"
              >
                <div class="step-header">
                  <span class="step-number">步骤 {{ index + 1 }}</span>
                  <el-button link type="danger" @click="removeStep(index)">
                    删除
                  </el-button>
                </div>
                <el-form-item label="步骤名称">
                  <el-input v-model="step.name" placeholder="如：查询订单" />
                </el-form-item>
                <el-form-item label="步骤类型">
                  <el-select v-model="step.type" placeholder="选择步骤类型">
                    <el-option label="查询 (Query)" value="query" />
                    <el-option label="验证 (Validate)" value="validate" />
                    <el-option label="更新 (Update)" value="update" />
                    <el-option label="插入 (Insert)" value="insert" />
                    <el-option label="删除 (Delete)" value="delete" />
                  </el-select>
                </el-form-item>
                <el-form-item label="操作表" v-if="step.type !== 'validate'">
                  <el-input v-model="step.table" placeholder="集合名称，如：orders" />
                </el-form-item>
                <el-form-item label="别名" v-if="step.type === 'query'">
                  <el-input v-model="step.alias" placeholder="用于后续步骤引用，如：order" />
                </el-form-item>
                <el-form-item label="查询条件" v-if="step.type === 'query' || step.type === 'update' || step.type === 'delete'">
                  <div class="conditions">
                    <div
                      v-for="(cond, condIdx) in step.conditions"
                      :key="condIdx"
                      class="condition-row"
                    >
                      <el-input v-model="cond.field" placeholder="字段" style="width: 120px;" />
                      <el-select v-model="cond.operator" placeholder="操作符" style="width: 100px;">
                        <el-option label="=" value="=" />
                        <el-option label="!=" value="!=" />
                        <el-option label=">" value=">" />
                        <el-option label="<" value="<" />
                        <el-option label=">=" value=">=" />
                        <el-option label="<=" value="<=" />
                        <el-option label="IN" value="in" />
                        <el-option label="LIKE" value="like" />
                      </el-select>
                      <el-input v-model="cond.value" placeholder="值" style="width: 120px;" />
                      <el-input v-model="cond.valueFrom" placeholder="或引用，如：params.id" style="width: 150px;" />
                      <el-button link type="danger" @click="removeCondition(step, condIdx)">×</el-button>
                    </div>
                    <el-button link type="primary" @click="addCondition(step)">+ 添加条件</el-button>
                  </div>
                </el-form-item>
                <el-form-item label="数据" v-if="step.type === 'insert' || step.type === 'update'">
                  <el-input
                    v-model="step.dataJson"
                    type="textarea"
                    :rows="3"
                    placeholder='{"status": "paid", "paidAt": "${now}"}'
                  />
                  <div class="form-tip">支持变量：${params.xxx}, ${user.id}, ${alias.field}</div>
                </el-form-item>
                <el-form-item label="验证条件" v-if="step.type === 'validate'">
                  <el-input v-model="step.validateCondition" placeholder="如：alias.balance >= params.amount" />
                  <div class="form-tip">表达式返回 true 表示验证通过</div>
                </el-form-item>
                <el-form-item label="错误消息">
                  <el-input v-model="step.error" placeholder="验证失败时的错误提示" />
                </el-form-item>
                <el-form-item label="失败处理">
                  <el-select v-model="step.onError" placeholder="选择处理方式">
                    <el-option label="失败并回滚 (fail)" value="fail" />
                    <el-option label="跳过继续 (skip)" value="skip" />
                  </el-select>
                </el-form-item>
              </div>
              <el-button type="primary" link @click="addStep">+ 添加步骤</el-button>
            </div>
          </el-form-item>
          <el-alert
            title="事务集合说明"
            type="info"
            :closable="false"
            show-icon
            style="margin-top: 12px;"
          >
            <template #default>
              事务集合用于执行多步骤的原子操作。<br>
              • 查询步骤：从数据库查询数据，可设置别名供后续引用<br>
              • 验证步骤：使用表达式验证数据是否满足条件<br>
              • 更新/插入/删除步骤：执行数据修改操作<br>
              • 所有步骤在事务中执行，任一步骤失败会自动回滚<br>
              • 使用 ${params.xxx} 引用传入参数，${user.id} 引用当前用户，${alias.field} 引用查询结果
            </template>
          </el-alert>
        </template>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCollections, createCollection, updateCollection, deleteCollection as deleteCollectionApi } from '@/api/collection'

const router = useRouter()

const loading = ref(false)
const submitting = ref(false)
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)
const selectedCollectionId = ref(null)

const collections = ref([])

const pagination = reactive({
  page: 1,
  perPage: 30,
  total: 0
})

const form = reactive({
  name: '',
  label: '',
  type: 'base',
  viewSQL: '',
  transactionSteps: []
})

const rules = {
  name: [
    { required: true, message: '请输入集合名称', trigger: 'blur' },
    { pattern: /^[a-z][a-z0-9_]*$/, message: '只能包含小写字母、数字和下划线，且以字母开头', trigger: 'blur' },
    { validator: validateTableName, trigger: 'blur' }
  ],
  label: [
    { required: true, message: '请输入中文别名', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择类型', trigger: 'change' }
  ]
}

// 验证表名是否符合数据库规则
function validateTableName(rule, value, callback) {
  if (!value) {
    callback()
    return
  }
  // 检查是否是保留字
  const reservedWords = ['admins', 'collections', 'settings', 'logs', 'files', 'realtime', 'user', 'users', 'order', 'group', 'index', 'key', 'value', 'status', 'type', 'name']
  if (reservedWords.includes(value.toLowerCase())) {
    callback(new Error('该名称是系统保留字，请使用其他名称'))
    return
  }
  // 检查长度
  if (value.length > 64) {
    callback(new Error('表名长度不能超过64个字符'))
    return
  }
  callback()
}

function getTypeTag(type) {
  const map = {
    base: 'info',
    auth: 'success',
    view: 'warning',
    transaction: 'danger'
  }
  return map[type] || 'info'
}

function getTypeName(type) {
  const map = {
    base: 'Base',
    auth: 'Auth',
    view: 'View',
    transaction: 'Transaction'
  }
  return map[type] || type
}

// 事务步骤管理
function addStep() {
  form.transactionSteps.push({
    name: '',
    type: 'query',
    table: '',
    alias: '',
    conditions: [],
    dataJson: '',
    validateCondition: '',
    error: '',
    onError: 'fail'
  })
}

function removeStep(index) {
  form.transactionSteps.splice(index, 1)
}

function addCondition(step) {
  step.conditions.push({
    field: '',
    operator: '=',
    value: '',
    valueFrom: ''
  })
}

function removeCondition(step, index) {
  step.conditions.splice(index, 1)
}

function showCreateDialog() {
  isEdit.value = false
  form.name = ''
  form.label = ''
  form.type = 'base'
  form.viewSQL = ''
  form.transactionSteps = []
  dialogVisible.value = true
}

function editCollection(row) {
  isEdit.value = true
  selectedCollectionId.value = row.id
  form.name = row.name
  form.label = row.label || row.name
  form.type = row.type
  // 加载视图集合的 SQL
  form.viewSQL = row.viewQuery || ''
  // 加载事务集合的步骤
  if (row.transactionSteps && row.transactionSteps.length > 0) {
    form.transactionSteps = row.transactionSteps.map(step => {
      const processedStep = { ...step }
      if (step.data) {
        try {
          processedStep.dataJson = JSON.stringify(step.data, null, 2)
        } catch (e) {
          processedStep.dataJson = ''
        }
      }
      return processedStep
    })
  } else {
    form.transactionSteps = []
  }
  dialogVisible.value = true
}

async function submitForm() {
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    if (!isEdit.value) {
      const requestData = {
        name: form.name,
        label: form.label,
        type: form.type,
        viewQuery: form.type === 'view' ? form.viewSQL : undefined
      }

      // 处理事务集合步骤
      if (form.type === 'transaction' && form.transactionSteps.length > 0) {
        // 转换数据格式：将 dataJson 字符串转为对象
        requestData.transactionSteps = form.transactionSteps.map(step => {
          const processedStep = { ...step }
          if (step.dataJson) {
            try {
              processedStep.data = JSON.parse(step.dataJson)
            } catch (e) {
              processedStep.data = {}
            }
          }
          delete processedStep.dataJson
          return processedStep
        })
      }

      await createCollection(requestData)
      ElMessage.success('创建成功')
    } else {
      // 编辑模式
      const requestData = {
        label: form.label
      }

      // 视图集合更新 SQL
      if (form.type === 'view') {
        requestData.viewQuery = form.viewSQL
      }

      // 事务集合更新步骤
      if (form.type === 'transaction' && form.transactionSteps.length > 0) {
        requestData.transactionSteps = form.transactionSteps.map(step => {
          const processedStep = { ...step }
          if (step.dataJson) {
            try {
              processedStep.data = JSON.parse(step.dataJson)
            } catch (e) {
              processedStep.data = {}
            }
          }
          delete processedStep.dataJson
          return processedStep
        })
      }

      await updateCollection(selectedCollectionId.value, requestData)
      ElMessage.success('更新成功')
    }
    dialogVisible.value = false
    loadCollections()
  } catch (error) {
    // error handled by interceptor
  } finally {
    submitting.value = false
  }
}

async function deleteCollection(row) {
  await ElMessageBox.confirm(
    `确定要删除集合 "${row.name}" 吗？此操作不可恢复！`,
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  )
  
  try {
    await deleteCollectionApi(row.id)
    ElMessage.success('删除成功')
    loadCollections()
  } catch (error) {
    // error handled by interceptor
  }
}

function goToDetail(name) {
  router.push(`/collections/${name}/detail`)
}

async function loadCollections() {
  loading.value = true
  try {
    const res = await getCollections({
      page: pagination.page,
      perPage: pagination.perPage
    })
    collections.value = res.data.items || []
    pagination.total = res.data.totalItems || 0
  } catch (error) {
    // error handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCollections()
})
</script>

<style lang="scss" scoped>
.collections-page {
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

  .create-guide {
    .type-cards {
      display: flex;
      gap: 12px;
    }

    .type-card {
      flex: 1;
      padding: 16px;
      border: 2px solid #ebeef5;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s;

      &:hover {
        border-color: #409eff;
      }

      &.active {
        border-color: #409eff;
        background-color: #ecf5ff;
      }

      .type-icon {
        font-size: 28px;
        margin-bottom: 8px;
      }

      .type-title {
        font-size: 14px;
        font-weight: 600;
        margin-bottom: 8px;
        color: #303133;
      }

      .type-desc {
        font-size: 12px;
        color: #909399;
        margin-bottom: 12px;
        line-height: 1.5;
      }

      .type-features {
        margin: 0;
        padding-left: 0;
        list-style: none;
        font-size: 12px;
        color: #606266;

        li {
          margin-bottom: 4px;
        }
      }
    }
  }

  .form-tip {
    font-size: 12px;
    color: #909399;
    margin-top: 4px;
  }

  .transaction-steps {
    width: 100%;

    .transaction-step {
      background: #f5f7fa;
      border: 1px solid #e4e7ed;
      border-radius: 8px;
      padding: 16px;
      margin-bottom: 16px;

      .step-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;

        .step-number {
          font-weight: 600;
          color: #303133;
        }
      }

      .conditions {
        .condition-row {
          display: flex;
          gap: 8px;
          align-items: center;
          margin-bottom: 8px;
        }
      }
    }
  }
}
</style>

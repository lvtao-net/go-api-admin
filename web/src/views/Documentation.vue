<template>
  <div class="api-docs-page" v-loading="loading">
    <div class="docs-container">
      <!-- 左侧菜单 -->
      <div class="docs-sidebar">
        <div class="sidebar-header">
          <h3>API 文档</h3>
          <el-input
            v-model="searchText"
            placeholder="搜索集合..."
            :prefix-icon="Search"
            size="small"
            clearable
          />
        </div>
        
        <div class="sidebar-content">
          <el-scrollbar>
            <!-- 错误提示 -->
            <div v-if="errorMsg" class="error-msg">
              <el-alert :title="errorMsg" type="error" show-icon />
            </div>
            
            <!-- 无数据提示 -->
            <div v-else-if="collections.length === 0 && !loading" class="empty-collections">
              <el-empty description="暂无集合数据" :image-size="80" />
            </div>
            
            <!-- 集合列表 -->
            <div v-else class="collection-list">
              <!-- 公共接口 -->
              <div v-if="publicEndpoints.length > 0" class="collection-group">
                <div
                  class="collection-header"
                  :class="{ active: selectedCollection === '_public' }"
                  @click="toggleCollection('_public')"
                >
                  <el-icon class="expand-icon" :class="{ expanded: expandedCollections.includes('_public') }">
                    <ArrowRight />
                  </el-icon>
                  <span class="collection-name">公共接口</span>
                  <el-tag size="small" type="success">Public</el-tag>
                </div>
                
                <transition name="slide">
                  <div v-if="expandedCollections.includes('_public')" class="api-list">
                    <div
                      v-for="api in publicEndpoints"
                      :key="api.method + api.path"
                      class="api-item"
                      :class="{ active: selectedAPI?.method === api.method && selectedAPI?.path === api.path }"
                      @click="selectPublicAPI(api)"
                    >
                      <span class="api-method" :class="api.method.toLowerCase()">{{ api.method }}</span>
                      <span class="api-name">{{ api.name }}</span>
                    </div>
                  </div>
                </transition>
              </div>
              
              <!-- 集合接口 -->
              <div
                v-for="collection in filteredCollections"
                :key="collection.name"
                class="collection-group"
              >
                <div
                  class="collection-header"
                  :class="{ active: selectedCollection === collection.name }"
                  @click="toggleCollection(collection.name)"
                >
                  <el-icon class="expand-icon" :class="{ expanded: expandedCollections.includes(collection.name) }">
                    <ArrowRight />
                  </el-icon>
                  <span class="collection-name">
                    {{ collection.label || collection.name }}
                  </span>
                  <el-tag size="small" type="info">{{ collection.name }}</el-tag>
                </div>
                
                <transition name="slide">
                  <div v-if="expandedCollections.includes(collection.name)" class="api-list">
                    <div
                      v-for="api in getCollectionAPIs(collection)"
                      :key="api.method + api.path"
                      class="api-item"
                      :class="{ active: selectedAPI?.method === api.method && selectedAPI?.path === api.path }"
                      @click="selectAPI(collection, api)"
                    >
                      <span class="api-method" :class="api.method.toLowerCase()">{{ api.method }}</span>
                      <span class="api-name">{{ api.name }}</span>
                    </div>
                  </div>
                </transition>
              </div>
            </div>
          </el-scrollbar>
        </div>
      </div>

      <!-- 右侧内容 -->
      <div class="docs-content">
        <template v-if="selectedAPI && selectedCollection">
          <div class="api-header">
            <div class="api-title">
              <span class="api-method-large" :class="selectedAPI.method.toLowerCase()">
                {{ selectedAPI.method }}
              </span>
              <span class="api-path">{{ selectedAPI.path }}</span>
            </div>
            <h2>{{ selectedAPI.name }}</h2>
            <p class="api-desc">{{ selectedAPI.description }}</p>
          </div>

          <el-divider />

          <!-- 认证信息 -->
          <div class="api-section">
            <h4>
              <el-icon><Lock /></el-icon>
              认证要求
            </h4>
            <div class="auth-info">
              <el-tag :type="currentAPIAuth.required ? 'warning' : 'success'">
                {{ currentAPIAuth.required ? '需要认证' : '无需认证' }}
              </el-tag>
              <span v-if="currentAPIAuth.description" class="auth-desc">
                {{ currentAPIAuth.description }}
              </span>
            </div>
          </div>

          <!-- 请求参数 -->
          <div v-if="currentAPIParams.length > 0" class="api-section">
            <h4>
              <el-icon><Document /></el-icon>
              请求参数
            </h4>
            <el-table :data="currentAPIParams" border size="small">
              <el-table-column label="参数名" width="180">
                <template #default="{ row }">
                  <div>
                    <span class="param-name">{{ row.name }}</span>
                    <el-tag v-if="row.required" size="small" type="danger">必填</el-tag>
                  </div>
                  <div v-if="row.label && row.label !== row.name" class="param-label">{{ row.label }}</div>
                </template>
              </el-table-column>
              <el-table-column prop="type" label="类型" width="90">
                <template #default="{ row }">
                  <el-tag size="small" type="info">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="rules" label="验证规则" width="150">
                <template #default="{ row }">
                  <span v-if="row.rules" class="validation-rules">{{ row.rules }}</span>
                  <span v-else style="color: #909399;">-</span>
                </template>
              </el-table-column>
              <el-table-column prop="description" label="说明" />
              <el-table-column prop="example" label="示例" width="120">
                <template #default="{ row }">
                  <code v-if="row.example">{{ row.example }}</code>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <!-- 请求示例 -->
          <div class="api-section">
            <h4>
              <el-icon><Cpu /></el-icon>
              请求示例
            </h4>
            <div class="code-block">
              <div class="code-header">
                <span>cURL</span>
                <el-button size="small" text @click="copyCode(curlExample)">
                  <el-icon><CopyDocument /></el-icon>
                  复制
                </el-button>
              </div>
              <pre>{{ curlExample }}</pre>
            </div>
          </div>

          <!-- 响应示例 -->
          <div class="api-section">
            <h4>
              <el-icon><SuccessFilled /></el-icon>
              响应示例
            </h4>
            <el-tabs v-model="responseTab">
              <el-tab-pane label="成功响应" name="success">
                <div class="code-block success">
                  <div class="code-header">
                    <span>JSON</span>
                    <el-button size="small" text @click="copyCode(successResponse)">
                      <el-icon><CopyDocument /></el-icon>
                      复制
                    </el-button>
                  </div>
                  <pre>{{ successResponse }}</pre>
                </div>
              </el-tab-pane>
              <el-tab-pane label="错误响应" name="error">
                <div class="code-block error">
                  <div class="code-header">
                    <span>JSON</span>
                    <el-button size="small" text @click="copyCode(errorResponse)">
                      <el-icon><CopyDocument /></el-icon>
                      复制
                    </el-button>
                  </div>
                  <pre>{{ errorResponse }}</pre>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
        </template>

        <template v-else>
          <div class="empty-state">
            <el-empty description="请从左侧选择一个 API 查看" />
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { ArrowRight, Lock, Document, Cpu, SuccessFilled, CopyDocument, Search } from '@element-plus/icons-vue'
import axios from 'axios'

const collections = ref([])
const endpoints = ref([])
const searchText = ref('')
const expandedCollections = ref([])
const selectedCollection = ref('')
const selectedAPI = ref(null)
const responseTab = ref('success')
const loading = ref(false)
const errorMsg = ref('')

// 过滤集合（view 和 transaction 类型用于前台 API 调用，也需要显示）
const filteredCollections = computed(() => {
  if (!searchText.value) return collections.value
  const search = searchText.value.toLowerCase()
  return collections.value.filter(c =>
    c.name.toLowerCase().includes(search) ||
    (c.label && c.label.toLowerCase().includes(search))
  )
})

// 获取集合的API列表（从后端返回的endpoints中筛选）
function getCollectionAPIs(collection) {
  return endpoints.value
    .filter(ep => ep.collection === collection.name)
    .map(ep => ({
      method: ep.method,
      name: ep.summary,
      path: ep.path,
      description: ep.description,
      auth: ep.auth,
      parameters: ep.parameters,
      requestBody: ep.requestBody,
      responses: ep.responses
    }))
}

// 获取公共接口列表
const publicEndpoints = computed(() => {
  return endpoints.value
    .filter(ep => ep.collection === '_public')
    .map(ep => ({
      method: ep.method,
      name: ep.summary,
      path: ep.path,
      description: ep.description,
      auth: ep.auth,
      parameters: ep.parameters,
      requestBody: ep.requestBody,
      responses: ep.responses
    }))
})

// 切换集合展开/折叠
function toggleCollection(name) {
  const index = expandedCollections.value.indexOf(name)
  if (index > -1) {
    expandedCollections.value.splice(index, 1)
  } else {
    expandedCollections.value.push(name)
  }
}

// 选择API
function selectAPI(collection, api) {
  selectedCollection.value = collection.name
  selectedAPI.value = api
}

// 选择公共API
function selectPublicAPI(api) {
  selectedCollection.value = '_public'
  selectedAPI.value = api
}

// 当前API的认证信息
const currentAPIAuth = computed(() => {
  if (!selectedAPI.value) {
    return { required: false, description: '' }
  }
  
  const auth = selectedAPI.value.auth
  if (!auth || auth === '公开') {
    return { required: false, description: '该接口无需认证即可访问' }
  }
  
  if (auth === '需要认证') {
    return { required: true, description: '需要登录认证，请在请求头中携带 Authorization: Bearer {token}' }
  }
  
  if (auth === '已禁用') {
    return { required: false, description: '该接口已被禁用' }
  }
  
  return { required: true, description: auth }
})

// 当前API的参数
const currentAPIParams = computed(() => {
  if (!selectedAPI.value) return []
  
  // 公共接口没有集合字段
  if (selectedCollection.value === '_public') {
    // 如果后端返回了参数，使用后端的参数
    if (selectedAPI.value.parameters && selectedAPI.value.parameters.length > 0) {
      return selectedAPI.value.parameters.map(p => ({
        name: p.name,
        type: p.type,
        required: p.required,
        description: p.description,
        example: p.example,
        rules: ''
      }))
    }
    return []
  }
  
  // Auth 接口不使用集合字段（登录、注册、刷新token等）
  const authEndpoints = ['/register', '/auth-with-password', '/auth-refresh', '/request-otp', '/reset-password']
  if (authEndpoints.some(ep => selectedAPI.value.path.includes(ep))) {
    // 如果后端返回了参数，使用后端的参数
    if (selectedAPI.value.parameters && selectedAPI.value.parameters.length > 0) {
      return selectedAPI.value.parameters.map(p => ({
        name: p.name,
        type: p.type,
        required: p.required,
        description: p.description,
        example: p.example,
        rules: ''
      }))
    }
    // 注册接口特殊处理
    if (selectedAPI.value.path.includes('/register')) {
      return [
        { name: 'email', type: 'string', required: true, description: '邮箱地址', example: 'user@example.com', rules: '' },
        { name: 'password', type: 'string', required: true, description: '登录密码', example: 'your_password', rules: '' },
        { name: 'code', type: 'string', required: false, description: '邮箱验证码（可选，如需验证先调用 /request-otp）', example: '123456', rules: '' }
      ]
    }
    return []
  }
  
  // GET 单条记录 - 只显示路由参数 id
  if (selectedAPI.value.method === 'GET' && selectedAPI.value.path.includes(':id')) {
    return [
      { name: 'id', type: 'string', required: true, description: '记录ID', example: '123', rules: '' }
    ]
  }
  
  // GET 列表接口参数
  if (selectedAPI.value.method === 'GET' && selectedAPI.value.path.includes('/records')) {
    return [
      { name: 'page', type: 'number', required: false, description: '页码，默认为 1', example: '1', rules: '' },
      { name: 'perPage', type: 'number', required: false, description: '每页数量，默认为 30', example: '30', rules: '' },
      { name: 'sort', type: 'string', required: false, description: '排序字段，-前缀表示倒序', example: '-created', rules: '' },
      { name: 'filter', type: 'string', required: false, description: '过滤条件', example: "status='active'", rules: '' }
    ]
  }
  
  // DELETE 接口 - 只显示路由参数 id
  if (selectedAPI.value.method === 'DELETE') {
    return [
      { name: 'id', type: 'string', required: true, description: '记录ID', example: '123', rules: '' }
    ]
  }
  
  const collection = collections.value.find(c => c.name === selectedCollection.value)
  if (!collection) return []
  
  // POST 接口参数 - 使用集合字段（创建）
  if (selectedAPI.value.method === 'POST' && selectedAPI.value.path.includes('/records') && collection.fields) {
    return collection.fields.map(f => {
      const rulesDisplay = formatValidationRules(f.validationRules)
      let description = f.description || ''
      if (f.dictionary) {
        description = description ? `${description} | 关联字典: ${f.dictionary}` : `关联字典: ${f.dictionary}`
      }
      return {
        name: f.name,
        label: f.label || f.name,
        type: f.type,
        required: f.required || false,
        description: description,
        example: getExampleValue(f),
        rules: rulesDisplay
      }
    })
  }
  
  // PATCH 接口参数 - 使用集合字段（更新），但标记为可选
  if (selectedAPI.value.method === 'PATCH' && selectedAPI.value.path.includes(':id') && collection.fields) {
    return collection.fields.map(f => {
      const rulesDisplay = formatValidationRules(f.validationRules)
      let description = f.description || ''
      if (f.dictionary) {
        description = description ? `${description} | 关联字典: ${f.dictionary}` : `关联字典: ${f.dictionary}`
      }
      return {
        name: f.name,
        label: f.label || f.name,
        type: f.type,
        required: false, // 更新时所有字段都是可选的
        description: description,
        example: getExampleValue(f),
        rules: rulesDisplay
      }
    })
  }
  
  return []
})

// 验证规则标签映射（包含描述）
const validationRuleLabels = {
  'required': { label: '必填', description: '字段不能为空' },
  'email': { label: '邮箱格式', description: '有效的邮箱地址' },
  'phone': { label: '手机号', description: '中国大陆手机号码' },
  'url': { label: 'URL格式', description: '有效的URL地址' },
  'idcard': { label: '身份证号', description: '中国大陆身份证号码' },
  'ip': { label: 'IP地址', description: 'IPv4或IPv6地址' },
  'ipv4': { label: 'IPv4地址', description: 'IPv4地址' },
  'ipv6': { label: 'IPv6地址', description: 'IPv6地址' },
  'number': { label: '数字', description: '有效的数字' },
  'integer': { label: '整数', description: '整数' },
  'positive': { label: '正数', description: '大于0的数' },
  'negative': { label: '负数', description: '小于0的数' },
  'alpha': { label: '纯字母', description: '只包含字母' },
  'alphanum': { label: '字母数字', description: '只包含字母和数字' },
  'chinese': { label: '中文', description: '只包含中文字符' },
  'date': { label: '日期格式', description: '日期格式 YYYY-MM-DD' },
  'datetime': { label: '日期时间格式', description: '日期时间格式' },
  'min_length': { label: '最小长度', description: '最小字符长度' },
  'max_length': { label: '最大长度', description: '最大字符长度' },
  'range_length': { label: '长度范围', description: '字符长度范围' },
  'min_value': { label: '最小值', description: '数字最小值' },
  'max_value': { label: '最大值', description: '数字最大值' },
  'range_value': { label: '值范围', description: '数字值范围' },
  'password_strength': { label: '密码强度', description: '至少8位，包含大小写字母和数字' },
  'credit_card': { label: '信用卡号', description: '有效的信用卡号' },
  'wechat': { label: '微信号', description: '微信号格式' },
  'qq': { label: 'QQ号', description: 'QQ号码' },
  'bank_card': { label: '银行卡号', description: '银行卡号格式' },
  'no_space': { label: '不含空格', description: '不能包含空格字符' },
  'no_special_char': { label: '不含特殊字符', description: '不能包含特殊字符' }
}

// 格式化验证规则显示
function formatValidationRules(rules) {
  if (!rules || rules.length === 0) return ''

  return rules.map(rule => {
    // 支持两种格式：对象格式（从API返回）和字符串格式（本地数据）
    let ruleName, ruleLabel, ruleDescription, ruleParams

    if (typeof rule === 'object' && rule !== null) {
      // 对象格式：{ name, label, description, params }
      ruleName = rule.name
      ruleLabel = rule.label
      ruleDescription = rule.description
      ruleParams = rule.params
    } else if (typeof rule === 'string') {
      // 字符串格式：可能是 "rule_name" 或 "rule_name:param=value"
      const colonIndex = rule.indexOf(':')
      ruleName = colonIndex > -1 ? rule.substring(0, colonIndex) : rule

      // 如果有参数，解析参数
      if (colonIndex > -1) {
        const paramsStr = rule.substring(colonIndex + 1)
        ruleParams = {}
        paramsStr.split(',').forEach(p => {
          const [key, value] = p.split('=')
          if (key && value) {
            ruleParams[key] = value
          }
        })
      }

      const ruleInfo = validationRuleLabels[ruleName] || { label: ruleName, description: '' }
      ruleLabel = ruleInfo.label
      ruleDescription = ruleInfo.description
    } else {
      return ''
    }

    // 如果有参数，显示参数信息
    if (ruleParams && Object.keys(ruleParams).length > 0) {
      if (ruleName === 'min_length' && ruleParams.min) {
        return `最小${ruleParams.min}字符`
      }
      if (ruleName === 'max_length' && ruleParams.max) {
        return `最大${ruleParams.max}字符`
      }
      if (ruleName === 'range_length' && ruleParams.min && ruleParams.max) {
        return `长度${ruleParams.min}-${ruleParams.max}字符`
      }
      if (ruleName === 'min_value' && ruleParams.min) {
        return `最小值${ruleParams.min}`
      }
      if (ruleName === 'max_value' && ruleParams.max) {
        return `最大值${ruleParams.max}`
      }
      if (ruleName === 'range_value' && ruleParams.min && ruleParams.max) {
        return `值范围${ruleParams.min}-${ruleParams.max}`
      }
    }

    // 显示规则名称和描述
    if (ruleDescription) {
      return `${ruleLabel}(${ruleDescription})`
    }
    return ruleLabel
  }).join('、')
}

// 获取示例值
function getExampleValue(field) {
  switch (field.type) {
    case 'text': return '"示例文本"'
    case 'number': return '123'
    case 'bool': return 'true'
    case 'email': return '"user@example.com"'
    case 'url': return '"https://example.com"'
    case 'date': return '"2024-01-01T00:00:00Z"'
    case 'json': return '{}'
    default: return ''
  }
}

// cURL 示例
const curlExample = computed(() => {
  if (!selectedAPI.value) return ''
  
  const baseUrl = import.meta.env.VITE_API_BASE_URL || ''
  const url = baseUrl + selectedAPI.value.path.replace('{id}', '123')
  
  let curl = `curl -X ${selectedAPI.value.method} '${url}' \\\n`
  
  if (currentAPIAuth.value.required) {
    curl += `  -H 'Authorization: Bearer YOUR_TOKEN' \\\n`
  }
  
  curl += `  -H 'Content-Type: application/json'`
  
  if (selectedAPI.value.method === 'POST' || selectedAPI.value.method === 'PATCH') {
    // 优先使用 requestBody 中的 example
    if (selectedAPI.value.requestBody?.content?.['application/json']?.example) {
      const exampleData = selectedAPI.value.requestBody.content['application/json'].example
      curl += ` \\\n  -d '${JSON.stringify(exampleData, null, 4)}'`
    } else {
      // 否则使用集合字段
      const collection = collections.value.find(c => c.name === selectedCollection.value)
      const fields = collection?.fields?.slice(0, 3) || []
      const exampleData = {}
      fields.forEach(f => {
        if (f.type === 'number') exampleData[f.name] = 123
        else if (f.type === 'bool') exampleData[f.name] = true
        else exampleData[f.name] = '示例'
      })
      if (Object.keys(exampleData).length > 0) {
        curl += ` \\\n  -d '${JSON.stringify(exampleData, null, 4)}'`
      }
    }
  }
  
  return curl
})

// 成功响应示例
const successResponse = computed(() => {
  if (!selectedAPI.value || !selectedCollection.value) return '{}'
  
  const collection = collections.value.find(c => c.name === selectedCollection.value)
  
  // Auth 接口响应
  if (selectedAPI.value.path.includes('/auth-with-password')) {
    return JSON.stringify({
      code: 0,
      message: 'success',
      data: {
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
        refreshToken: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...',
        record: {
          id: '1',
          email: 'user@example.com',
          created: '2024-01-01 10:00:00'
        }
      }
    }, null, 2)
  }
  
  if (selectedAPI.value.path.includes('/register')) {
    return JSON.stringify({
      code: 0,
      message: 'success',
      data: {
        id: '1',
        email: 'user@example.com',
        created: '2024-01-01 10:00:00'
      }
    }, null, 2)
  }
  
  if (selectedAPI.value.path.includes('/auth-refresh')) {
    return JSON.stringify({
      code: 0,
      message: 'success',
      data: {
        token: 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
      }
    }, null, 2)
  }
  
  if (selectedAPI.value.path.includes('/request-otp')) {
    return JSON.stringify({
      code: 0,
      message: 'success',
      data: {
        message: '验证码已发送'
      }
    }, null, 2)
  }
  
  if (selectedAPI.value.path.includes('/reset-password')) {
    return JSON.stringify({
      code: 0,
      message: 'success',
      data: {
        message: '密码重置成功'
      }
    }, null, 2)
  }
  
  if (selectedAPI.value.method === 'GET' && selectedAPI.value.path.includes('{id}')) {
    // 单条记录
    const example = {
      id: 'RECORD_ID',
      created: '2024-01-01 10:00:00',
      updated: '2024-01-01 10:00:00'
    }
    collection?.fields?.slice(0, 3).forEach(f => {
      if (f.type === 'number') example[f.name] = 123
      else if (f.type === 'bool') example[f.name] = true
      else example[f.name] = '示例'
    })
    return JSON.stringify(example, null, 2)
  }
  
  if (selectedAPI.value.method === 'GET') {
    // 列表
    return JSON.stringify({
      page: 1,
      perPage: 30,
      totalItems: 100,
      totalPages: 4,
      items: [
        { id: 'RECORD_ID_1', created: '2024-01-01 10:00:00', updated: '2024-01-01 10:00:00' },
        { id: 'RECORD_ID_2', created: '2024-01-01 11:00:00', updated: '2024-01-01 11:00:00' }
      ]
    }, null, 2)
  }
  
  if (selectedAPI.value.method === 'POST') {
    return JSON.stringify({
      id: 'NEW_RECORD_ID',
      created: '2024-01-01 10:00:00',
      updated: '2024-01-01 10:00:00'
    }, null, 2)
  }
  
  return JSON.stringify({ success: true }, null, 2)
})

// 错误响应示例
const errorResponse = computed(() => {
  return JSON.stringify({
    code: 400,
    message: '请求参数错误',
    errors: {
      field: '该字段是必填项'
    }
  }, null, 2)
})

// 复制代码
function copyCode(code) {
  navigator.clipboard.writeText(code).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}

// 加载集合列表（使用公开的API文档接口）
async function loadCollections() {
  loading.value = true
  errorMsg.value = ''
  try {
    const res = await axios.get('/api/doc')
    if (res.data.code === 0 && res.data.data) {
      collections.value = res.data.data.collections || []
      endpoints.value = res.data.data.endpoints || []
      // 默认展开第一个集合
      if (collections.value.length > 0) {
        expandedCollections.value = [collections.value[0].name]
      }
    }
  } catch (error) {
    console.error('加载集合失败:', error)
    errorMsg.value = error.message || '加载集合失败'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadCollections()
})
</script>

<style lang="scss" scoped>
.api-docs-page {
  height: 100%;
  background: #fff;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.docs-container {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.docs-sidebar {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  background: #fafafa;
  overflow: hidden;

  .sidebar-header {
    padding: 16px;
    border-bottom: 1px solid #e4e7ed;

    h3 {
      margin: 0 0 12px;
      font-size: 16px;
    }
  }

  .sidebar-content {
    flex: 1;
    overflow: hidden;
  }

  .error-msg {
    padding: 12px;
  }

  .empty-collections {
    padding: 24px;
    display: flex;
    justify-content: center;
  }

  .collection-list {
    padding: 8px;
  }

  .collection-group {
    margin-bottom: 4px;
  }

  .collection-header {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s;

    &:hover {
      background: #e9ecf0;
    }

    &.active {
      background: #e6f0ff;
    }

    .expand-icon {
      margin-right: 8px;
      transition: transform 0.2s;

      &.expanded {
        transform: rotate(90deg);
      }
    }

    .collection-name {
      flex: 1;
      font-weight: 500;
    }
  }

  .api-list {
    padding-left: 20px;
  }

  .api-item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    cursor: pointer;
    border-radius: 4px;
    margin: 2px 0;

    &:hover {
      background: #e9ecf0;
    }

    &.active {
      background: #e6f0ff;
    }

    .api-method {
      font-size: 11px;
      font-weight: 600;
      padding: 2px 6px;
      border-radius: 3px;
      margin-right: 8px;
      min-width: 45px;
      text-align: center;

      &.get { background: #e8f5e9; color: #2e7d32; }
      &.post { background: #e3f2fd; color: #1565c0; }
      &.patch { background: #fff3e0; color: #ef6c00; }
      &.delete { background: #ffebee; color: #c62828; }
    }

    .api-name {
      font-size: 13px;
      color: #606266;
    }
  }
}

.docs-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;

  .api-header {
    .api-title {
      display: flex;
      align-items: center;
      margin-bottom: 12px;
    }

    .api-method-large {
      font-size: 14px;
      font-weight: 600;
      padding: 4px 10px;
      border-radius: 4px;
      margin-right: 12px;

      &.get { background: #e8f5e9; color: #2e7d32; }
      &.post { background: #e3f2fd; color: #1565c0; }
      &.patch { background: #fff3e0; color: #ef6c00; }
      &.delete { background: #ffebee; color: #c62828; }
    }

    .api-path {
      font-family: 'Monaco', 'Menlo', monospace;
      font-size: 16px;
      color: #303133;
    }

    h2 {
      margin: 0 0 8px;
      font-size: 20px;
    }

    .api-desc {
      margin: 0;
      color: #909399;
    }
  }

  .api-section {
    margin-bottom: 24px;

    h4 {
      display: flex;
      align-items: center;
      gap: 8px;
      margin: 0 0 12px;
      font-size: 15px;
      color: #303133;
    }
  }

  .auth-info {
    display: flex;
    align-items: center;
    gap: 12px;

    .auth-desc {
      color: #909399;
      font-size: 13px;
    }
  }

  .param-name {
    font-family: 'Monaco', 'Menlo', monospace;
    margin-right: 8px;
  }

  .code-block {
    background: #1e1e1e;
    border-radius: 6px;
    overflow: hidden;

    &.success {
      background: #1a2e1a;
    }

    &.error {
      background: #2e1a1a;
    }

    .code-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 12px;
      background: rgba(255, 255, 255, 0.05);
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);

      span {
        color: #888;
        font-size: 12px;
      }
    }

    pre {
      margin: 0;
      padding: 12px;
      color: #d4d4d4;
      font-family: 'Monaco', 'Menlo', monospace;
      font-size: 13px;
      line-height: 1.5;
      overflow-x: auto;
    }
  }

  .empty-state {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.2s ease;
}

.slide-enter-from,
.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

code {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 12px;
}

.validation-rules {
  font-size: 12px;
  color: #e6a23c;
  background: #fdf6ec;
  padding: 2px 6px;
  border-radius: 3px;
}

.param-label {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}
</style>

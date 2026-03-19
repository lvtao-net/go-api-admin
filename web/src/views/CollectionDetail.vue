<template>
  <div class="collection-detail">
    <el-page-header @back="goBack">
      <template #content>
        <span class="page-title">{{ collectionName }}</span>
        <el-tag size="small" style="margin-left: 8px;">{{ collectionType }}</el-tag>
      </template>
    </el-page-header>
    
    <el-tabs v-model="activeTab" class="detail-tabs">
      <!-- 字段设置 - 仅 base/auth 类型显示 -->
      <el-tab-pane v-if="!isNoDataCollection" label="字段" name="fields">
        <div class="tab-header">
          <el-button type="primary" @click="showAddFieldDialog">
            <el-icon><Plus /></el-icon>
            添加字段
          </el-button>
        </div>

        <el-table :data="fields" stripe>
          <el-table-column prop="name" label="字段名" width="150">
            <template #default="{ row }">
              <div>{{ row.name }}</div>
              <div v-if="row.label && row.label !== row.name" class="field-label-hint">{{ row.label }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="unique" label="唯一值" width="80">
            <template #default="{ row }">
              <el-tag :type="row.unique ? 'warning' : 'info'" size="small">
                {{ row.unique ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="验证规则" width="200">
            <template #default="{ row }">
              <template v-if="row.validationRules && row.validationRules.length > 0">
                <el-tag
                  v-for="rule in row.validationRules.slice(0, 3)"
                  :key="rule"
                  size="small"
                  style="margin-right: 4px; margin-bottom: 2px;"
                >
                  {{ getRuleLabel(rule) }}
                </el-tag>
                <el-tag v-if="row.validationRules.length > 3" size="small" type="info">
                  +{{ row.validationRules.length - 3 }}
                </el-tag>
              </template>
              <span v-else style="color: #909399;">-</span>
            </template>
          </el-table-column>
          <el-table-column prop="defaultValue" label="默认值" />
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button link type="primary" @click="editField(row)">编辑</el-button>
              <el-button link type="danger" @click="deleteField(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 记录列表 - 仅 base/auth 类型显示 -->
      <el-tab-pane v-if="!isNoDataCollection" label="记录" name="records">
        <div class="tab-header">
          <el-button type="primary" @click="showCreateRecordDialog">
            <el-icon><Plus /></el-icon>
            新建记录
          </el-button>
          <el-button @click="loadRecords">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>

        <el-table :data="records" v-loading="loading" stripe>
          <el-table-column v-for="col in displayColumns" :key="col" :prop="col" :label="col" min-width="120" />
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="viewRecord(row)">
                详情
              </el-button>
              <el-button link type="primary" @click="editRecord(row)">
                编辑
              </el-button>
              <el-button link type="danger" @click="deleteRecordConfirm(row)">
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
            @size-change="loadRecords"
            @current-change="loadRecords"
          />
        </div>
      </el-tab-pane>

      <!-- API 规则 - 所有类型都需要 -->
      <el-tab-pane label="API 规则" name="rules">
        <div class="rules-container">
          <!-- 权限预设说明 -->
          <div class="permission-presets-info">
            <el-alert type="info" :closable="false" show-icon>
              <template #title>
                <strong>权限预设说明</strong>
              </template>
              <div class="preset-descriptions">
                <p><strong>公开</strong> - 无需登录，所有人可访问</p>
                <p><strong>登录用户</strong> - 需要登录认证</p>
                <p><strong>仅自己</strong> - 只能操作自己的记录</p>
                <p><strong>管理员</strong> - 仅管理员可操作</p>
                <p><strong>禁止</strong> - 完全禁止访问</p>
              </div>
            </el-alert>
          </div>

          <el-divider />

          <!-- 每个操作的独立权限选择 -->
          <div class="operation-permissions">
            <el-table :data="operationRules" border>
              <el-table-column label="操作" width="120">
                <template #default="{ row }">
                  <el-tag :type="row.tagType" size="large">{{ row.label }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="权限预设" width="300">
                <template #default="{ row }">
                  <el-radio-group v-model="row.preset" @change="(val) => applyOperationPreset(row.key, val)">
                    <el-radio-button value="public">公开</el-radio-button>
                    <el-radio-button value="auth">登录用户</el-radio-button>
                    <el-radio-button value="owner">仅自己</el-radio-button>
                    <el-radio-button value="admin">管理员</el-radio-button>
                    <el-radio-button value="disabled">禁止</el-radio-button>
                    <el-radio-button value="custom">自定义</el-radio-button>
                  </el-radio-group>
                </template>
              </el-table-column>
              <el-table-column label="规则表达式">
                <template #default="{ row }">
                  <el-input
                    v-model="row.rule"
                    type="textarea"
                    :rows="2"
                    placeholder="自定义规则表达式"
                    :disabled="row.preset !== 'custom'"
                    @change="onRuleChange(row.key)"
                  />
                  <div class="rule-hint" v-if="row.preset === 'custom'">
                    可用变量: @request.auth.id, @request.auth.email, @request.auth.admin
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>

          <el-divider />

          <!-- 字段级别 API 权限 - 仅 base/auth 类型显示 -->
          <div class="field-api-permissions" v-if="!isNoDataCollection && fieldAPISettings.length > 0">
            <div class="section-title">字段 API 权限</div>
            <el-alert type="warning" :closable="false" style="margin-bottom: 16px;">
              控制字段在 API 中的可见性和可编辑性
            </el-alert>
            <el-table :data="fieldAPISettings" border size="small">
              <el-table-column prop="name" label="字段名" width="150" />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="API 禁止" width="100" align="center">
                <template #default="{ row }">
                  <el-switch v-model="row.apiDisabled" @change="updateFieldAPISetting(row)" />
                </template>
              </el-table-column>
              <el-table-column label="只读" width="100" align="center">
                <template #default="{ row }">
                  <el-switch v-model="row.apiReadOnly" @change="updateFieldAPISetting(row)" />
                </template>
              </el-table-column>
              <el-table-column label="只写" width="100" align="center">
                <template #default="{ row }">
                  <el-switch v-model="row.apiWriteOnly" @change="updateFieldAPISetting(row)" />
                </template>
              </el-table-column>
              <el-table-column label="列表隐藏" width="100" align="center">
                <template #default="{ row }">
                  <el-switch v-model="row.apiHiddenList" @change="updateFieldAPISetting(row)" />
                </template>
              </el-table-column>
              <el-table-column label="详情隐藏" width="100" align="center">
                <template #default="{ row }">
                  <el-switch v-model="row.apiHiddenView" @change="updateFieldAPISetting(row)" />
                </template>
              </el-table-column>
            </el-table>
          </div>

          <el-divider />

          <!-- 关联字段权限配置 -->
          <div class="field-permissions" v-if="relationFields.length > 0">
            <div class="section-title">关联字段权限</div>
            <el-form label-width="120px" size="small">
              <el-form-item
                v-for="field in relationFields"
                :key="field.name"
                :label="field.name + ' 字段'"
              >
                <el-select
                  v-model="fieldPermissions[field.name]"
                  placeholder="选择权限"
                  @change="updateFieldPermission(field.name)"
                >
                  <el-option label="公开" value="" />
                  <el-option label="登录用户可访问" value="@request.auth.id != ''" />
                  <el-option label="仅创建者可访问" value="@request.auth.id = @request.body.' + field.name + '.id" />
                </el-select>
              </el-form-item>
            </el-form>
          </div>

          <div style="margin-top: 20px;">
            <el-button type="primary" @click="saveRules" :loading="savingRules">保存所有规则</el-button>
          </div>
        </div>
      </el-tab-pane>
      
      <!-- 设置 -->
      <el-tab-pane label="设置" name="settings">
        <el-form ref="settingsFormRef" :model="settingsForm" :rules="settingsRules" label-width="100px" style="max-width: 500px;">
          <el-form-item label="集合名称">
            <el-input v-model="collectionName" disabled />
          </el-form-item>
          <el-form-item label="中文别名" prop="label">
            <el-input v-model="settingsForm.label" placeholder="请输入中文别名" />
          </el-form-item>
          <el-form-item label="菜单图标">
            <el-select v-model="settingsForm.icon" placeholder="选择图标" clearable style="width: 200px;">
              <el-option
                v-for="icon in availableIcons"
                :key="icon.value"
                :label="icon.label"
                :value="icon.value"
              >
                <div style="display: flex; align-items: center; gap: 8px;">
                  <el-icon><component :is="getIconComponent(icon.value)" /></el-icon>
                  <span>{{ icon.label }}</span>
                </div>
              </el-option>
            </el-select>
            <el-icon v-if="settingsForm.icon" style="margin-left: 12px; font-size: 20px;">
              <component :is="getIconComponent(settingsForm.icon)" />
            </el-icon>
          </el-form-item>
          <el-form-item label="集合类型">
            <el-input v-model="collectionType" disabled />
          </el-form-item>

          <!-- Base/Auth 集合：主键和查找字段配置 -->
          <template v-if="!isNoDataCollection">
            <el-divider content-position="left">API 参数配置</el-divider>

            <el-form-item label="主键字段">
              <el-select v-model="settingsForm.primaryKeyField" placeholder="选择主键字段" style="width: 200px;">
                <el-option label="id (默认)" value="id" />
                <el-option
                  v-for="field in fields.filter(f => f.type === 'number' || f.type === 'text')"
                  :key="field.name"
                  :label="field.label || field.name"
                  :value="field.name"
                />
              </el-select>
              <div class="form-tip">主键字段用于 API 路由中的 :id 参数</div>
            </el-form-item>

            <el-form-item label="可查找字段">
              <div class="lookup-fields-config">
                <div class="lookup-header">
                  <span>配置可通过 URL 查找记录的字段</span>
                </div>
                <div
                  v-for="(lookup, index) in settingsForm.lookupFields"
                  :key="index"
                  class="lookup-field-item"
                >
                  <el-select v-model="lookup.field" placeholder="选择字段" style="width: 150px;">
                    <el-option
                      v-for="field in fields.filter(f => f.unique || f.type === 'text' || f.type === 'number')"
                      :key="field.name"
                      :label="field.label || field.name"
                      :value="field.name"
                    />
                  </el-select>
                  <el-checkbox v-model="lookup.required" style="margin-left: 8px;">必填</el-checkbox>
                  <el-button link type="danger" @click="removeLookupField(index)" style="margin-left: 8px;">删除</el-button>
                </div>
                <el-button type="primary" link @click="addLookupField">+ 添加查找字段</el-button>
                <div class="form-tip">
                  添加后可通过 /api/collections/{collection}/records/by/{field}/{value} 查找记录
                </div>
              </div>
            </el-form-item>
          </template>

          <!-- 视图/事务集合：路由参数配置 -->
          <template v-if="isNoDataCollection">
            <el-divider content-position="left">路由参数配置</el-divider>

            <el-form-item label="路由参数">
              <div class="route-params-config">
                <div
                  v-for="(param, index) in settingsForm.routeParams"
                  :key="index"
                  class="route-param-item"
                >
                  <el-input v-model="param.name" placeholder="参数名" style="width: 120px;" />
                  <el-select v-model="param.type" placeholder="类型" style="width: 100px;">
                    <el-option label="字符串" value="string" />
                    <el-option label="数字" value="number" />
                    <el-option label="布尔" value="bool" />
                  </el-select>
                  <el-select v-model="param.source" placeholder="来源" style="width: 100px;">
                    <el-option label="路径" value="path" />
                    <el-option label="查询" value="query" />
                    <el-option label="请求体" value="body" />
                  </el-select>
                  <el-checkbox v-model="param.required" style="margin-left: 8px;">必填</el-checkbox>
                  <el-input v-model="param.default" placeholder="默认值" style="width: 100px;" />
                  <el-button link type="danger" @click="removeRouteParam(index)" style="margin-left: 8px;">×</el-button>
                </div>
                <el-button type="primary" link @click="addRouteParam">+ 添加参数</el-button>
                <div class="form-tip">
                  配置 API 调用时需要的参数，如：year, month, status 等
                </div>
              </div>
            </el-form-item>
          </template>

          <!-- 视图集合：SQL 查询可编辑 -->
          <el-form-item v-if="collectionType === 'view' || collectionType === 'View'" label="视图SQL">
            <el-input v-model="settingsForm.viewSQL" type="textarea" :rows="4" placeholder="SELECT * FROM users WHERE status = 'active'" />
            <div class="form-tip">输入创建视图的 SQL 查询语句</div>
          </el-form-item>
          <!-- 事务集合：事务步骤配置 -->
          <el-form-item v-if="collectionType === 'transaction' || collectionType === 'Transaction'" label="事务步骤">
            <div class="transaction-steps">
              <div
                v-for="(step, index) in settingsForm.transactionSteps"
                :key="index"
                class="transaction-step"
              >
                <div class="step-header">
                  <span class="step-number">步骤 {{ index + 1 }}</span>
                  <el-button link type="danger" @click="removeSettingStep(index)">
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
                      <el-button link type="danger" @click="removeSettingCondition(step, condIdx)">×</el-button>
                    </div>
                    <el-button link type="primary" @click="addSettingCondition(step)">+ 添加条件</el-button>
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
              <el-button type="primary" link @click="addSettingStep">+ 添加步骤</el-button>
            </div>
          </el-form-item>
          <el-form-item label="菜单隐藏">
            <el-switch v-model="settingsForm.menuHidden" />
            <div class="form-tip">开启后在左侧菜单中隐藏此集合</div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveSettings" :loading="savingSettings">保存设置</el-button>
            <el-button type="danger" @click="deleteCollectionConfirm">删除集合</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>
    </el-tabs>
    
    <!-- 记录对话框 -->
    <el-dialog
      v-model="recordDialogVisible"
      :title="isEditRecord ? '编辑记录' : '新建记录'"
      width="600px"
    >
      <el-form ref="recordFormRef" :model="recordForm" label-width="100px">
        <el-form-item
          v-for="field in fields"
          :key="field.name"
          :label="field.name"
          :prop="field.name"
        >
          <el-input
            v-if="field.type === 'text'"
            v-model="recordForm[field.name]"
            type="textarea"
            :rows="3"
          />
          <el-input-number
            v-else-if="field.type === 'number'"
            v-model="recordForm[field.name]"
          />
          <el-switch
            v-else-if="field.type === 'bool'"
            v-model="recordForm[field.name]"
          />
          <el-date-picker
            v-else-if="field.type === 'date'"
            v-model="recordForm[field.name]"
            type="datetime"
            placeholder="选择日期时间"
          />
          <el-select
            v-else-if="field.type === 'select'"
            v-model="recordForm[field.name]"
            placeholder="请选择"
          >
            <el-option
              v-for="opt in (field.options?.values || [])"
              :key="opt"
              :label="opt"
              :value="opt"
            />
          </el-select>
          <el-radio-group
            v-else-if="field.type === 'radio'"
            v-model="recordForm[field.name]"
          >
            <el-radio
              v-for="opt in getFieldOptions(field)"
              :key="opt.value"
              :value="opt.value"
            >
              {{ opt.label }}
            </el-radio>
          </el-radio-group>
          <el-checkbox-group
            v-else-if="field.type === 'checkbox'"
            v-model="recordForm[field.name]"
          >
            <el-checkbox
              v-for="opt in getFieldOptions(field)"
              :key="opt.value"
              :label="opt.value"
            >
              {{ opt.label }}
            </el-checkbox>
          </el-checkbox-group>
          <el-input
            v-else
            v-model="recordForm[field.name]"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="recordDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitRecord" :loading="submittingRecord">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 记录详情对话框 -->
    <el-dialog
      v-model="viewDialogVisible"
      title="记录详情"
      width="600px"
    >
      <el-descriptions :column="2" border v-if="currentRecord">
        <el-descriptions-item label="ID">
          {{ currentRecord.id }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ formatDate(currentRecord.created) }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDate(currentRecord.updated) }}
        </el-descriptions-item>
      </el-descriptions>
      <el-divider />
      <div class="record-detail-fields">
        <div v-for="(value, key) in displayRecordData" :key="key" class="detail-item">
          <span class="detail-label">{{ key }}:</span>
          <span class="detail-value">{{ formatValue(value) }}</span>
        </div>
      </div>
      <template #footer>
        <el-button @click="viewDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="editCurrentRecord">
          编辑
        </el-button>
      </template>
    </el-dialog>

    <!-- 字段对话框 -->
    <el-dialog
      v-model="fieldDialogVisible"
      :title="isEditField ? '编辑字段' : '添加字段'"
      width="500px"
    >
      <el-form ref="fieldFormRef" :model="fieldForm" :rules="fieldRules" label-width="80px">
        <el-form-item label="字段名" prop="name">
          <el-input v-model="fieldForm.name" placeholder="数据库字段名，如 title、user_id" :disabled="isEditField" />
          <div class="form-tip">字段名用于数据库存储和API交互，只能包含字母、数字和下划线</div>
        </el-form-item>
        <el-form-item label="显示名称" prop="label">
          <el-input v-model="fieldForm.label" placeholder="如 标题、用户ID" />
          <div class="form-tip">显示名称用于前端界面展示，方便用户理解字段含义</div>
        </el-form-item>
        <el-form-item label="字段说明">
          <el-input v-model="fieldForm.description" type="textarea" :rows="2" placeholder="字段用途说明，将显示在API文档中" />
        </el-form-item>
        <el-form-item label="类型" prop="type">
          <el-select v-model="fieldForm.type" placeholder="请选择类型" @change="onTypeChange">
            <el-option label="文本 (text)" value="text" />
            <el-option label="数字 (number)" value="number" />
            <el-option label="布尔 (bool)" value="bool" />
            <el-option label="邮箱 (email)" value="email" />
            <el-option label="URL (url)" value="url" />
            <el-option label="日期 (date)" value="date" />
            <el-option label="单选 (radio)" value="radio" />
            <el-option label="多选 (checkbox)" value="checkbox" />
            <el-option label="选择器 (select)" value="select" />
            <el-option label="关联关系 (relation)" value="relation" />
            <el-option label="文件 (file)" value="file" />
            <el-option label="富文本 (editor)" value="editor" />
            <el-option label="JSON (json)" value="json" />
            <el-option label="密码 (password)" value="password" />
          </el-select>
        </el-form-item>
        <el-form-item label="唯一值">
          <el-switch v-model="fieldForm.unique" />
          <div class="form-tip">开启后，该字段的值不能重复（如邮箱、账号等）</div>
        </el-form-item>
        <el-form-item label="默认值">
          <el-input v-model="fieldForm.defaultValue" placeholder="默认值" />
        </el-form-item>
        <!-- 字典关联选项：单选、多选、下拉只能关联字典 -->
        <el-form-item v-if="['radio', 'checkbox', 'select'].includes(fieldForm.type)" label="关联字典" prop="dictionary">
          <el-select v-model="fieldForm.dictionary" placeholder="请选择字典" @change="onDictionaryChange">
            <el-option v-for="dict in dictionaries" :key="dict.id" :label="dict.label || dict.name" :value="dict.name" />
          </el-select>
          <div class="form-tip" v-if="fieldForm.dictionary">
            已关联字典：{{ getDictionaryLabel(fieldForm.dictionary) }}
          </div>
          <div class="form-tip" v-else style="color: #e6a23c;">
            请先创建字典，然后在字典管理中添加选项
          </div>
        </el-form-item>
        <el-form-item v-if="fieldForm.type === 'relation'" label="关联集合">
          <el-input v-model="fieldForm.relationCollection" placeholder="关联的集合名称" />
        </el-form-item>
        <el-form-item v-if="fieldForm.type === 'relation'" label="级联删除">
          <el-switch v-model="fieldForm.relationCascade" />
        </el-form-item>
        <el-form-item v-if="fieldForm.type === 'relation'" label="多对多">
          <el-switch v-model="fieldForm.relationMultiple" @change="onRelationMultipleChange" />
        </el-form-item>
        <!-- 验证规则 -->
        <el-form-item label="验证规则">
          <el-select
            v-model="fieldForm.validationRules"
            multiple
            collapse-tags
            collapse-tags-tooltip
            placeholder="选择验证规则"
            style="width: 100%;"
            @change="onValidationRulesChange"
          >
            <el-option-group
              v-for="(rules, category) in validationRulesByCategory"
              :key="category"
              :label="category"
            >
              <el-option
                v-for="rule in rules"
                :key="rule.name"
                :label="rule.label"
                :value="rule.name"
              >
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span>{{ rule.label }}</span>
                  <span style="color: #909399; font-size: 12px;">{{ rule.description }}</span>
                </div>
              </el-option>
            </el-option-group>
          </el-select>
          <div class="form-tip">
            选择验证规则后，数据提交时会自动验证
          </div>
        </el-form-item>
        <!-- 验证规则参数设置 -->
        <el-form-item
          v-for="ruleName in fieldForm.validationRules.filter(r => ruleNeedsParams(r))"
          :key="ruleName"
          :label="getRuleLabel(ruleName) + '参数'"
        >
          <div class="validation-params">
            <template v-for="param in getRuleParamsDef(ruleName)" :key="param.key">
              <div class="param-item">
                <span class="param-label">{{ param.label }}:</span>
                <el-input-number
                  v-model="fieldForm.validationParams[ruleName][param.key]"
                  :min="param.key === 'min' ? 0 : undefined"
                  :max="param.key === 'max' ? 999999 : undefined"
                  size="small"
                  style="width: 120px;"
                />
              </div>
            </template>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="fieldDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitField" :loading="submittingField">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Document, Folder, User, Setting, Lock, View, Grid, DataBoard, 
  Collection, Tickets, Key, Calendar, ChatDotRound, Star, 
  Location, Phone, Message, ShoppingCart, Bell, Timer,
  Plus, Refresh, Edit, Delete, Search, Upload, Download,
  Picture, VideoCamera, Headset, Wallet, TrendCharts
} from '@element-plus/icons-vue'
import { getCollection, updateCollection, deleteCollection as deleteCollectionApi, checkDeleteCollection, getRecords, createRecord, updateRecord, deleteRecord as deleteRecordApi } from '@/api/collection'
import { getDictionaries, getDictionaryByName } from '@/api/dictionary'

const route = useRoute()
const router = useRouter()

const collectionName = computed(() => route.params.name)
const collectionType = ref('Base')

// 判断是否是无数据集合（view/transaction类型没有实际表）
const isNoDataCollection = computed(() => {
  return collectionType.value === 'view' || collectionType.value === 'transaction' || 
         collectionType.value === 'View' || collectionType.value === 'Transaction'
})

const activeTab = ref('fields')
const loading = ref(false)
const records = ref([])
const fields = ref([])
const displayColumns = ref(['id', 'created', 'updated'])

const pagination = reactive({
  page: 1,
  perPage: 30,
  total: 0
})

const rules = reactive({
  listRule: '',
  viewRule: '',
  createRule: '',
  updateRule: '',
  deleteRule: ''
})
const savingRules = ref(false)

// 设置表单
const settingsFormRef = ref(null)
const savingSettings = ref(false)
const settingsForm = reactive({
  label: '',
  viewSQL: '',
  menuHidden: false,
  icon: '',
  primaryKeyField: 'id',
  lookupFields: [],
  routeParams: [],
  transactionSteps: []
})
const settingsRules = {
  label: [{ required: true, message: '请输入中文别名', trigger: 'blur' }]
}

// 可用图标列表
const availableIcons = [
  { value: 'Document', label: '文档' },
  { value: 'Folder', label: '文件夹' },
  { value: 'User', label: '用户' },
  { value: 'Setting', label: '设置' },
  { value: 'Lock', label: '锁' },
  { value: 'View', label: '视图' },
  { value: 'Grid', label: '网格' },
  { value: 'DataBoard', label: '仪表盘' },
  { value: 'Collection', label: '集合' },
  { value: 'Tickets', label: '票据' },
  { value: 'Key', label: '钥匙' },
  { value: 'Calendar', label: '日历' },
  { value: 'ChatDotRound', label: '聊天' },
  { value: 'Star', label: '收藏' },
  { value: 'Location', label: '位置' },
  { value: 'Phone', label: '电话' },
  { value: 'Message', label: '消息' },
  { value: 'ShoppingCart', label: '购物车' },
  { value: 'Bell', label: '通知' },
  { value: 'Timer', label: '计时器' },
  { value: 'Edit', label: '编辑' },
  { value: 'Delete', label: '删除' },
  { value: 'Search', label: '搜索' },
  { value: 'Upload', label: '上传' },
  { value: 'Download', label: '下载' },
  { value: 'Picture', label: '图片' },
  { value: 'VideoCamera', label: '视频' },
  { value: 'Headset', label: '音频' },
  { value: 'Wallet', label: '钱包' },
  { value: 'TrendCharts', label: '图表' }
]

// 图标映射
const iconComponents = {
  Document, Folder, User, Setting, Lock, View, Grid, DataBoard,
  Collection, Tickets, Key, Calendar, ChatDotRound, Star,
  Location, Phone, Message, ShoppingCart, Bell, Timer,
  Edit, Delete, Search, Upload, Download,
  Picture, VideoCamera, Headset, Wallet, TrendCharts
}

function getIconComponent(iconName) {
  return iconComponents[iconName] || Document
}

// 快捷配置
const quickPreset = ref('')
const relationFields = ref([])
const fieldPermissions = reactive({})

// 字典相关
const dictionaries = ref([])
const dictionaryMap = reactive({})
const dictionaryCache = reactive({})
const fieldOptionsCache = reactive({})

// 操作规则列表（根据集合类型动态生成）
const operationRules = ref([])

// 基础操作规则配置
const baseOperationConfigs = [
  { key: 'listRule', label: '列表', tagType: 'success' },
  { key: 'viewRule', label: '详情', tagType: 'info' },
  { key: 'createRule', label: '创建', tagType: 'primary' },
  { key: 'updateRule', label: '更新', tagType: 'warning' },
  { key: 'deleteRule', label: '删除', tagType: 'danger' }
]

// 视图集合只支持查询操作
const viewOperationConfigs = [
  { key: 'listRule', label: '列表', tagType: 'success' },
  { key: 'viewRule', label: '详情', tagType: 'info' }
]

// 事务集合只支持执行操作
const transactionOperationConfigs = [
  { key: 'createRule', label: '执行', tagType: 'primary' }
]

// 字段 API 设置
const fieldAPISettings = ref([])

// 单个操作的预设规则
const operationPresetRules = {
  public: '',                              // 公开 - 无限制
  auth: '@request.auth.id != ""',          // 登录用户
  owner: '@request.auth.id = id',          // 仅自己（需要在运行时替换 id）
  admin: '@request.auth.admin = true',     // 仅管理员
  disabled: 'false',                       // 禁止
}

// 应用操作预设
function applyOperationPreset(key, preset) {
  const op = operationRules.value.find(o => o.key === key)
  if (!op) return

  if (preset === 'custom') {
    // 自定义时保持当前规则
    return
  }

  let rule = operationPresetRules[preset] || ''
  
  // 特殊处理 owner 预设
  if (preset === 'owner') {
    if (key === 'listRule') {
      rule = '@request.auth.id != ""'  // 列表：登录用户可看自己的
    } else if (key === 'createRule') {
      rule = '@request.auth.id != ""'  // 创建：登录用户可创建
    } else {
      rule = '@request.auth.id = id'   // 其他：只能操作自己的
    }
  }

  op.rule = rule
  rules[key] = rule
}

// 规则变化时
function onRuleChange(key) {
  const op = operationRules.value.find(o => o.key === key)
  if (op) {
    rules[key] = op.rule
    op.preset = 'custom'
  }
}

// 更新字段 API 设置
function updateFieldAPISetting(field) {
  // 更新 fields 数组中对应字段的设置
  const fieldIndex = fields.value.findIndex(f => f.name === field.name)
  if (fieldIndex !== -1) {
    fields.value[fieldIndex] = {
      ...fields.value[fieldIndex],
      apiDisabled: field.apiDisabled,
      apiReadOnly: field.apiReadOnly,
      apiWriteOnly: field.apiWriteOnly,
      apiHiddenList: field.apiHiddenList,
      apiHiddenView: field.apiHiddenView
    }
  }
}

// 预设规则（保留旧版兼容）
const presetRules = {
  // 公开 - 所有人都可以访问
  public: {
    listRule: '',
    viewRule: '',
    createRule: '',
    updateRule: '',
    deleteRule: ''
  },
  // 登录用户 - 需要登录才能访问
  auth: {
    listRule: '@request.auth.id != ""',
    viewRule: '@request.auth.id != ""',
    createRule: '@request.auth.id != ""',
    updateRule: '@request.auth.id != ""',
    deleteRule: '@request.auth.id != ""'
  },
  // 创建者 - 只有记录创建者可以访问
  owner: {
    listRule: '@request.auth.id != ""',
    viewRule: '@request.auth.id = id',
    createRule: '@request.auth.id != ""',
    updateRule: '@request.auth.id = id',
    deleteRule: '@request.auth.id = id'
  },
  // 仅管理员 - 只有管理员可以访问
  admin: {
    listRule: '@request.auth.admin = true',
    viewRule: '@request.auth.admin = true',
    createRule: '@request.auth.admin = true',
    updateRule: '@request.auth.admin = true',
    deleteRule: '@request.auth.admin = true'
  }
}

function applyQuickPreset(preset) {
  quickPreset.value = preset
  if (preset === 'custom') {
    return
  }
  const presetRule = presetRules[preset]
  if (presetRule) {
    rules.listRule = presetRule.listRule || ''
    rules.viewRule = presetRule.viewRule || ''
    rules.createRule = presetRule.createRule || ''
    rules.updateRule = presetRule.updateRule || ''
    rules.deleteRule = presetRule.deleteRule || ''
  }
}

function updateFieldPermission(fieldName) {
  // 关联字段权限会在保存规则时一起保存
}

function detectRelationFields() {
  relationFields.value = []
  for (const field of fields.value) {
    if (field.type === 'relation') {
      relationFields.value.push({
        name: field.name,
        collection: field.relationCollection || ''
      })
    }
  }
}

// 记录对话框
const recordDialogVisible = ref(false)
const isEditRecord = ref(false)
const submittingRecord = ref(false)
const recordFormRef = ref(null)
const recordForm = reactive({})

// 记录详情对话框
const viewDialogVisible = ref(false)
const currentRecord = ref(null)
const displayRecordData = ref({})

// 字段对话框
const fieldDialogVisible = ref(false)
const isEditField = ref(false)
const submittingField = ref(false)
const fieldFormRef = ref(null)
const fieldForm = reactive({
  name: '',
  label: '',           // 显示名称
  description: '',     // 字段说明
  type: 'text',
  unique: false,
  defaultValue: null,
  dictionary: '',
  relationCollection: '',
  relationCascade: false,
  relationMultiple: false,
  validationRules: [],
  validationParams: {},  // 存储规则参数 { min_length: { min: 8 }, range_value: { min: 0, max: 100 } }
  validationMessages: {}
})
const fieldRules = {
  name: [{ required: true, message: '请输入字段名', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  dictionary: [
    {
      validator: (rule, value, callback) => {
        if (['radio', 'checkbox', 'select'].includes(fieldForm.type) && !value) {
          callback(new Error('请选择关联字典'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
}

// 内置验证规则列表
const builtInValidationRules = [
  { name: 'required', label: '必填', description: '字段不能为空', category: '基础' },
  { name: 'email', label: '邮箱', description: '有效的邮箱地址', category: '格式' },
  { name: 'phone', label: '手机号', description: '中国大陆手机号码', category: '格式' },
  { name: 'url', label: 'URL', description: '有效的URL地址', category: '格式' },
  { name: 'idcard', label: '身份证号', description: '中国大陆身份证号码', category: '格式' },
  { name: 'ip', label: 'IP地址', description: 'IPv4或IPv6地址', category: '格式' },
  { name: 'ipv4', label: 'IPv4', description: 'IPv4地址', category: '格式' },
  { name: 'ipv6', label: 'IPv6', description: 'IPv6地址', category: '格式' },
  { name: 'number', label: '数字', description: '有效的数字', category: '类型' },
  { name: 'integer', label: '整数', description: '整数', category: '类型' },
  { name: 'positive', label: '正数', description: '大于0的数', category: '类型' },
  { name: 'negative', label: '负数', description: '小于0的数', category: '类型' },
  { name: 'alpha', label: '纯字母', description: '只包含字母', category: '格式' },
  { name: 'alphanum', label: '字母数字', description: '只包含字母和数字', category: '格式' },
  { name: 'chinese', label: '中文', description: '只包含中文字符', category: '格式' },
  { name: 'date', label: '日期', description: '日期格式 YYYY-MM-DD', category: '格式' },
  { name: 'datetime', label: '日期时间', description: '日期时间格式', category: '格式' },
  { name: 'min_length', label: '最小长度', description: '最小字符长度', category: '长度', hasParams: true, params: [{ key: 'min', label: '最小长度', type: 'number', default: 1 }] },
  { name: 'max_length', label: '最大长度', description: '最大字符长度', category: '长度', hasParams: true, params: [{ key: 'max', label: '最大长度', type: 'number', default: 255 }] },
  { name: 'range_length', label: '长度范围', description: '字符长度范围', category: '长度', hasParams: true, params: [{ key: 'min', label: '最小长度', type: 'number', default: 1 }, { key: 'max', label: '最大长度', type: 'number', default: 255 }] },
  { name: 'min_value', label: '最小值', description: '数字最小值', category: '范围', hasParams: true, params: [{ key: 'min', label: '最小值', type: 'number', default: 0 }] },
  { name: 'max_value', label: '最大值', description: '数字最大值', category: '范围', hasParams: true, params: [{ key: 'max', label: '最大值', type: 'number', default: 100 }] },
  { name: 'range_value', label: '值范围', description: '数字值范围', category: '范围', hasParams: true, params: [{ key: 'min', label: '最小值', type: 'number', default: 0 }, { key: 'max', label: '最大值', type: 'number', default: 100 }] },
  { name: 'password_strength', label: '密码强度', description: '至少8位，包含大小写字母和数字', category: '安全' },
  { name: 'credit_card', label: '信用卡号', description: '有效的信用卡号', category: '格式' },
  { name: 'wechat', label: '微信号', description: '微信号格式', category: '格式' },
  { name: 'qq', label: 'QQ号', description: 'QQ号码', category: '格式' },
  { name: 'bank_card', label: '银行卡号', description: '银行卡号格式', category: '格式' },
  { name: 'hex_color', label: '十六进制颜色', description: '如 #FFF 或 #FFFFFF', category: '格式' },
  { name: 'json', label: 'JSON格式', description: '有效的JSON格式', category: '格式' },
  { name: 'uuid', label: 'UUID', description: 'UUID格式', category: '格式' },
  { name: 'no_space', label: '不含空格', description: '不能包含空格字符', category: '格式' },
  { name: 'no_special_char', label: '不含特殊字符', description: '不能包含特殊字符', category: '格式' },
]

// 需要参数的规则名称列表
const rulesWithParams = ['min_length', 'max_length', 'range_length', 'min_value', 'max_value', 'range_value']

// 按类别分组的验证规则
const validationRulesByCategory = computed(() => {
  const result = {}
  for (const rule of builtInValidationRules) {
    if (!result[rule.category]) {
      result[rule.category] = []
    }
    result[rule.category].push(rule)
  }
  return result
})

// 获取验证规则标签（支持解析带参数的规则，显示详细描述）
function getRuleLabel(ruleName) {
  // 支持两种格式：对象格式和字符串格式
  let baseName, params

  if (typeof ruleName === 'object' && ruleName !== null) {
    // 对象格式：{ name, label, description, params }
    baseName = ruleName.name
    params = ruleName.params || {}
  } else if (typeof ruleName === 'string') {
    // 字符串格式
    const parsed = parseRule(ruleName)
    baseName = parsed.name
    params = parsed.params
  } else {
    return String(ruleName)
  }

  const rule = builtInValidationRules.find(r => r.name === baseName)

  if (!rule) {
    return baseName
  }

  // 如果有参数，显示参数信息
  if (Object.keys(params).length > 0) {
    // 根据规则类型格式化参数显示
    if (baseName === 'range_value') {
      return `${rule.label}(${params.min ?? 0} ~ ${params.max ?? 100})`
    } else if (baseName === 'range_length') {
      return `${rule.label}(${params.min ?? 1} ~ ${params.max ?? 255}字符)`
    } else if (baseName === 'min_value') {
      return `${rule.label}(≥${params.min ?? 0})`
    } else if (baseName === 'max_value') {
      return `${rule.label}(≤${params.max ?? 100})`
    } else if (baseName === 'min_length') {
      return `${rule.label}(≥${params.min ?? 1}字符)`
    } else if (baseName === 'max_length') {
      return `${rule.label}(≤${params.max ?? 255}字符)`
    }
    // 其他带参数的规则
    const paramStr = Object.entries(params).map(([k, v]) => `${k}=${v}`).join(', ')
    return `${rule.label}(${paramStr})`
  }

  // 没有参数但有描述的规则，显示描述
  if (rule.description) {
    return `${rule.label}(${rule.description})`
  }

  return rule.label
}

// 解析规则字符串，返回 { name, params }
function parseRule(ruleStr) {
  if (!ruleStr) return { name: '', params: {} }
  const parts = ruleStr.split(':')
  const name = parts[0]
  const params = {}
  if (parts[1]) {
    const paramPairs = parts[1].split(',')
    for (const pair of paramPairs) {
      const [key, value] = pair.split('=')
      if (key && value !== undefined) {
        params[key] = isNaN(value) ? value : Number(value)
      }
    }
  }
  return { name, params }
}

// 构建规则字符串（带参数）
function buildRuleString(name, params) {
  if (!params || Object.keys(params).length === 0) {
    return name
  }
  const paramStr = Object.entries(params)
    .map(([k, v]) => `${k}=${v}`)
    .join(',')
  return `${name}:${paramStr}`
}

// 获取规则的基础名称（不带参数）
function getRuleBaseName(ruleStr) {
  return ruleStr ? ruleStr.split(':')[0] : ''
}

// 检查规则是否需要参数
function ruleNeedsParams(ruleName) {
  return rulesWithParams.includes(ruleName)
}

// 获取规则参数定义
function getRuleParamsDef(ruleName) {
  const rule = builtInValidationRules.find(r => r.name === ruleName)
  return rule?.params || []
}

// 获取规则标签（显示用）
function getRuleDisplayLabel(ruleName) {
  const rule = builtInValidationRules.find(r => r.name === ruleName)
  return rule ? rule.label : ruleName
}

// 从已选规则中解析参数到 validationParams
function initValidationParams() {
  fieldForm.validationParams = {}
  for (const ruleStr of fieldForm.validationRules) {
    const { name, params } = parseRule(ruleStr)
    if (Object.keys(params).length > 0) {
      fieldForm.validationParams[name] = params
    }
  }
}

// 当验证规则选择变化时
function onValidationRulesChange(selectedRules) {
  // 初始化新选择规则的默认参数
  for (const ruleStr of selectedRules) {
    const baseName = getRuleBaseName(ruleStr)
    if (ruleNeedsParams(baseName) && !fieldForm.validationParams[baseName]) {
      const paramsDef = getRuleParamsDef(baseName)
      const defaultParams = {}
      for (const p of paramsDef) {
        defaultParams[p.key] = p.default
      }
      fieldForm.validationParams[baseName] = defaultParams
    }
  }
  // 清理已删除规则的参数
  const selectedBaseNames = selectedRules.map(r => getRuleBaseName(r))
  for (const key of Object.keys(fieldForm.validationParams)) {
    if (!selectedBaseNames.includes(key)) {
      delete fieldForm.validationParams[key]
    }
  }
}

// 加载字典列表
async function loadDictionaries() {
  try {
    const res = await getDictionaries({ perPage: 100 })
    dictionaries.value = res.data.items || []
    dictionaries.value.forEach(dict => {
      dictionaryMap[dict.name] = dict.label || dict.name
    })
  } catch (error) {
    console.error('加载字典失败:', error)
  }
}

function getDictionaryLabel(name) {
  return dictionaryMap[name] || name
}

// 获取字段选项（从缓存中获取）
function getFieldOptions(field) {
  if (fieldOptionsCache[field.name]) {
    return fieldOptionsCache[field.name]
  }
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

function onDictionaryChange(value) {
  // 字典选择变更
}

function goBack() {
  router.push('/collections')
}

// 事务步骤管理（设置页面）
function addSettingStep() {
  settingsForm.transactionSteps.push({
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

function removeSettingStep(index) {
  settingsForm.transactionSteps.splice(index, 1)
}

function addSettingCondition(step) {
  step.conditions.push({
    field: '',
    operator: '=',
    value: '',
    valueFrom: ''
  })
}

function removeSettingCondition(step, index) {
  step.conditions.splice(index, 1)
}

// 查找字段管理
function addLookupField() {
  settingsForm.lookupFields.push({
    field: '',
    required: false,
    validation: ''
  })
}

function removeLookupField(index) {
  settingsForm.lookupFields.splice(index, 1)
}

// 路由参数管理
function addRouteParam() {
  settingsForm.routeParams.push({
    name: '',
    type: 'string',
    source: 'query',
    required: false,
    default: '',
    description: ''
  })
}

function removeRouteParam(index) {
  settingsForm.routeParams.splice(index, 1)
}

function showCreateRecordDialog() {
  isEditRecord.value = false
  Object.keys(recordForm).forEach(key => delete recordForm[key])
  fields.value.forEach(f => {
    if (f.type === 'checkbox') {
      recordForm[f.name] = []
    } else {
      recordForm[f.name] = f.defaultValue || null
    }
  })
  recordDialogVisible.value = true
}

function editRecord(row) {
  isEditRecord.value = true
  Object.assign(recordForm, row)
  // 处理 checkbox 类型的数据（可能是 JSON 字符串）
  fields.value.forEach(f => {
    if (f.type === 'checkbox' && typeof recordForm[f.name] === 'string') {
      try {
        recordForm[f.name] = JSON.parse(recordForm[f.name])
      } catch (e) {
        recordForm[f.name] = recordForm[f.name] ? [recordForm[f.name]] : []
      }
    }
  })
  recordDialogVisible.value = true
}

function viewRecord(row) {
  currentRecord.value = row
  // 过滤掉系统字段，只显示业务数据
  const data = { ...row }
  delete data.id
  delete data.created
  delete data.updated
  displayRecordData.value = data
  viewDialogVisible.value = true
}

function editCurrentRecord() {
  if (currentRecord.value) {
    viewDialogVisible.value = false
    editRecord(currentRecord.value)
  }
}

function formatDate(date) {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

function formatValue(value) {
  if (value === null || value === undefined) return '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

async function submitRecord() {
  submittingRecord.value = true
  try {
    if (isEditRecord.value) {
      const id = recordForm.id
      delete recordForm.id
      delete recordForm.created
      delete recordForm.updated
      await updateRecord(collectionName.value, id, recordForm)
      ElMessage.success('更新成功')
    } else {
      await createRecord(collectionName.value, recordForm)
      ElMessage.success('创建成功')
    }
    recordDialogVisible.value = false
    loadRecords()
  } catch (error) {
    // handled
  } finally {
    submittingRecord.value = false
  }
}

async function deleteRecordConfirm(row) {
  await ElMessageBox.confirm('确定要删除此记录吗？', '警告', { type: 'warning' })
  try {
    await deleteRecordApi(collectionName.value, row.id)
    ElMessage.success('删除成功')
    loadRecords()
  } catch (error) {
    // handled
  }
}

function showAddFieldDialog() {
  isEditField.value = false
  Object.assign(fieldForm, {
    name: '',
    label: '',
    description: '',
    type: 'text',
    unique: false,
    defaultValue: null,
    dictionary: '',
    relationCollection: '',
    relationCascade: false,
    relationMultiple: false,
    validationRules: [],
    validationParams: {},
    validationMessages: {}
  })
  fieldDialogVisible.value = true
}

function editField(row) {
  isEditField.value = true

  // 解析验证规则中的参数
  const validationRules = []
  const validationParams = {}
  for (const ruleStr of (row.validationRules || [])) {
    const { name, params } = parseRule(ruleStr)
    validationRules.push(name)
    // 对于需要参数的规则，确保参数对象存在
    if (ruleNeedsParams(name)) {
      const paramsDef = getRuleParamsDef(name)
      const defaultParams = {}
      for (const p of paramsDef) {
        defaultParams[p.key] = params[p.key] !== undefined ? params[p.key] : p.default
      }
      validationParams[name] = defaultParams
    } else if (Object.keys(params).length > 0) {
      validationParams[name] = params
    }
  }

  Object.assign(fieldForm, {
    name: row.name,
    label: row.label || '',
    description: row.description || '',
    type: row.type,
    unique: row.unique || false,
    defaultValue: row.defaultValue,
    dictionary: row.dictionary || '',
    relationCollection: row.relationCollection || '',
    relationCascade: row.relationCascade || false,
    relationMultiple: row.relationMax > 1,
    validationRules,
    validationParams,
    validationMessages: row.validationMessages || {}
  })
  fieldDialogVisible.value = true
}

function onTypeChange() {
  if (!['radio', 'checkbox', 'select'].includes(fieldForm.type)) {
    fieldForm.dictionary = ''
  }
  if (fieldForm.type !== 'relation') {
    fieldForm.relationCollection = ''
    fieldForm.relationCascade = false
    fieldForm.relationMultiple = false
  }
}

function onRelationMultipleChange(val) {
  if (val) {
    fieldForm.relationMax = 10
  } else {
    fieldForm.relationMax = 1
  }
}

async function submitField() {
  const valid = await fieldFormRef.value.validate().catch(() => false)
  if (!valid) return

  // 单选、多选、下拉必须关联字典
  if (['radio', 'checkbox', 'select'].includes(fieldForm.type) && !fieldForm.dictionary) {
    ElMessage.warning('请选择关联字典')
    return
  }

  // 处理关联字段的多对多设置
  if (fieldForm.type === 'relation') {
    fieldForm.relationMax = fieldForm.relationMultiple ? 10 : 1
  }

  // 将验证规则和参数合并为带参数的规则字符串
  const finalValidationRules = fieldForm.validationRules.map(ruleName => {
    const params = fieldForm.validationParams[ruleName]
    return buildRuleString(ruleName, params)
  })

  submittingField.value = true
  try {
    // 显式构建字段数据，确保所有字段都包含
    const fieldData = {
      name: fieldForm.name,
      label: fieldForm.label,
      description: fieldForm.description,
      type: fieldForm.type,
      unique: fieldForm.unique,
      defaultValue: fieldForm.defaultValue,
      dictionary: fieldForm.dictionary,
      relationCollection: fieldForm.relationCollection,
      relationCascade: fieldForm.relationCascade,
      relationMax: fieldForm.relationMax,
      validationRules: finalValidationRules,
      validationMessages: fieldForm.validationMessages
    }

    const newFields = isEditField.value
      ? fields.value.map(f => f.name === fieldForm.name ? { ...f, ...fieldData } : f)
      : [...fields.value, fieldData]

    await updateCollection(collectionName.value, { fields: newFields })
    ElMessage.success(isEditField.value ? '更新成功' : '添加成功')
    fieldDialogVisible.value = false
    loadCollection()
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error(error.message || '操作失败')
  } finally {
    submittingField.value = false
  }
}

async function deleteField(row) {
  await ElMessageBox.confirm(`确定要删除字段 "${row.name}" 吗？`, '警告', { type: 'warning' })
  try {
    const newFields = fields.value.filter(f => f.name !== row.name)
    await updateCollection(collectionName.value, { fields: newFields })
    ElMessage.success('删除成功')
    loadCollection()
  } catch (error) {
    // handled
  }
}

async function saveRules() {
  savingRules.value = true
  try {
    // 构建更新后的字段列表
    const updatedFields = fields.value.map((f, index) => {
      const apiSetting = fieldAPISettings.value[index] || {}
      return {
        ...f,
        apiDisabled: apiSetting.apiDisabled || false,
        apiReadOnly: apiSetting.apiReadOnly || false,
        apiWriteOnly: apiSetting.apiWriteOnly || false,
        apiHiddenList: apiSetting.apiHiddenList || false,
        apiHiddenView: apiSetting.apiHiddenView || false
      }
    })

    // 只保存当前集合类型支持的规则
    const rulesData = {}
    for (const op of operationRules.value) {
      rulesData[op.key] = rules[op.key] || null
    }

    await updateCollection(collectionName.value, {
      ...rulesData,
      fields: updatedFields
    })
    ElMessage.success('规则保存成功')
    loadCollection()
  } catch (error) {
    // handled
  } finally {
    savingRules.value = false
  }
}

async function saveSettings() {
  const valid = await settingsFormRef.value?.validate().catch(() => false)
  if (!valid) return

  savingSettings.value = true
  try {
    const updateData = {
      label: settingsForm.label,
      menuHidden: settingsForm.menuHidden,
      icon: settingsForm.icon
    }

    // Base/Auth 集合：保存主键和查找字段配置
    if (!isNoDataCollection.value) {
      updateData.primaryKeyField = settingsForm.primaryKeyField || 'id'
      updateData.lookupFields = settingsForm.lookupFields.filter(l => l.field)
    }

    // 视图/事务集合：保存路由参数配置
    if (isNoDataCollection.value) {
      updateData.routeParams = settingsForm.routeParams.filter(p => p.name)
    }

    // 视图集合更新 SQL
    if (collectionType.value === 'view' || collectionType.value === 'View') {
      updateData.viewQuery = settingsForm.viewSQL
    }

    // 事务集合更新步骤
    if ((collectionType.value === 'transaction' || collectionType.value === 'Transaction') && settingsForm.transactionSteps.length > 0) {
      updateData.transactionSteps = settingsForm.transactionSteps.map(step => {
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

    await updateCollection(collectionName.value, updateData)
    ElMessage.success('设置保存成功')
  } catch (error) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    savingSettings.value = false
  }
}

async function deleteCollectionConfirm() {
  // 先检查删除条件
  try {
    const checkRes = await checkDeleteCollection(collectionName.value)
    const checkData = checkRes.data

    if (!checkData.canDelete) {
      let msg = `集合 "${collectionName.value}" 无法删除：\n`

      if (checkData.hasData) {
        msg += `\n⚠️ 该集合当前包含 ${checkData.recordCount} 条数据`
      }

      if (checkData.relatedCount > 0) {
        msg += `\n⚠️ 该集合被以下集合关联：${checkData.related.join(', ')}`
      }

      ElMessage.warning(msg)
      return
    }

    // 可以删除，显示确认对话框
    await ElMessageBox.confirm(
      `确定要删除集合 "${collectionName.value}" 吗？此操作不可恢复！`,
      '警告',
      { type: 'error' }
    )

    await deleteCollectionApi(collectionName.value)
    ElMessage.success('删除成功')
    router.push('/collections')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

async function loadRecords() {
  loading.value = true
  try {
    const res = await getRecords(collectionName.value, {
      page: pagination.page,
      perPage: pagination.perPage
    })
    records.value = res.data.items || []
    pagination.total = res.data.totalItems || 0
    
    // 更新显示列
    if (records.value.length > 0) {
      const firstRecord = records.value[0]
      displayColumns.value = ['id', 'created', 'updated', ...Object.keys(firstRecord).filter(k => !['id', 'created', 'updated'].includes(k))]
    }
  } catch (error) {
    // handled
  } finally {
    loading.value = false
  }
}

async function loadCollection() {
  try {
    const res = await getCollection(collectionName.value)
    const data = res.data
    collectionType.value = data.type || 'Base'
    fields.value = data.fields || []

    // 对于 view/transaction 类型，默认选中 API 规则标签页
    const isNoData = data.type === 'view' || data.type === 'transaction' ||
                     data.type === 'View' || data.type === 'Transaction'
    if (isNoData) {
      activeTab.value = 'rules'
    }

    // 加载设置
    settingsForm.label = data.label || ''
    settingsForm.viewSQL = data.viewSQL || data.viewQuery || ''
    settingsForm.menuHidden = data.menuHidden || false
    settingsForm.icon = data.icon || ''
    settingsForm.primaryKeyField = data.primaryKeyField || 'id'
    settingsForm.lookupFields = data.lookupFields || []
    settingsForm.routeParams = data.routeParams || []

    // 加载事务步骤
    if (data.transactionSteps && data.transactionSteps.length > 0) {
      settingsForm.transactionSteps = data.transactionSteps.map(step => {
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
      settingsForm.transactionSteps = []
    }

    // 加载规则
    rules.listRule = data.listRule || ''
    rules.viewRule = data.viewRule || ''
    rules.createRule = data.createRule || ''
    rules.updateRule = data.updateRule || ''
    rules.deleteRule = data.deleteRule || ''

    // 根据集合类型初始化操作规则
    let configs
    if (data.type === 'view' || data.type === 'View') {
      configs = viewOperationConfigs
    } else if (data.type === 'transaction' || data.type === 'Transaction') {
      configs = transactionOperationConfigs
    } else {
      configs = baseOperationConfigs
    }

    operationRules.value = configs.map(config => ({
      ...config,
      preset: detectPreset(config.key, data[config.key] || ''),
      rule: data[config.key] || ''
    }))

    // 初始化字段 API 设置
    fieldAPISettings.value = fields.value.map(f => ({
      name: f.name,
      type: f.type,
      apiDisabled: f.apiDisabled || false,
      apiReadOnly: f.apiReadOnly || false,
      apiWriteOnly: f.apiWriteOnly || false,
      apiHiddenList: f.apiHiddenList || false,
      apiHiddenView: f.apiHiddenView || false
    }))

    // 检测当前预设
    detectQuickPreset()

    // 加载关联字段
    detectRelationFields()

    // 加载字段选项（字典关联）
    await loadFieldOptions()

    // 加载记录（只有有数据集合才加载）
    if (!isNoData) {
      loadRecords()
    }
  } catch (error) {
    // handled
  }
}

// 检测单个操作的预设
function detectPreset(key, rule) {
  if (rule === '') return 'public'
  if (rule === '@request.auth.id != ""') return 'auth'
  if (rule === '@request.auth.admin = true') return 'admin'
  if (rule === 'false') return 'disabled'
  if (rule === '@request.auth.id = id' || rule === '@request.auth.id = @request.body.id') return 'owner'
  return 'custom'
}

function detectQuickPreset() {
  const ruleStr = JSON.stringify({
    listRule: rules.listRule,
    viewRule: rules.viewRule,
    createRule: rules.createRule,
    updateRule: rules.updateRule,
    deleteRule: rules.deleteRule
  })

  for (const [preset, presetRule] of Object.entries(presetRules)) {
    const presetStr = JSON.stringify(presetRule)
    if (presetStr === ruleStr) {
      quickPreset.value = preset
      return
    }
  }
  quickPreset.value = 'custom'
}

onMounted(() => {
  loadCollection()
  loadDictionaries()
})
</script>

<style lang="scss" scoped>
.collection-detail {
  .page-title {
    font-size: 18px;
    font-weight: 500;
  }
  
  .detail-tabs {
    margin-top: 20px;
  }
  
  .tab-header {
    margin-bottom: 16px;
  }
  
  .pagination-container {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
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

  .field-label-hint {
    font-size: 12px;
    color: #909399;
    margin-top: 2px;
  }

  .lookup-fields-config,
  .route-params-config {
    width: 100%;

    .lookup-header {
      font-size: 13px;
      color: #606266;
      margin-bottom: 12px;
    }

    .lookup-field-item,
    .route-param-item {
      display: flex;
      align-items: center;
      margin-bottom: 8px;
      gap: 8px;
    }
  }

  .validation-params {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    
    .param-item {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .param-label {
        font-size: 13px;
        color: #606266;
        white-space: nowrap;
      }
    }
  }

  .rules-container {
    .section-title {
      font-size: 16px;
      font-weight: 500;
      margin-bottom: 16px;
    }

    .section-subtitle {
      font-size: 14px;
      font-weight: 500;
      margin: 16px 0;
      color: #606266;
    }

    .permission-presets-info {
      .preset-descriptions {
        margin-top: 8px;
        font-size: 13px;
        p {
          margin: 4px 0;
        }
      }
    }

    .operation-permissions {
      margin-bottom: 20px;

      .rule-hint {
        font-size: 11px;
        color: #909399;
        margin-top: 4px;
      }
    }

    .field-api-permissions {
      margin-bottom: 20px;
    }

    .quick-config {
      margin-bottom: 20px;

      .quick-presets {
        margin-bottom: 20px;

        .el-button-group {
          flex-wrap: wrap;
        }
      }

      .field-permissions {
        margin-top: 20px;
        padding: 16px;
        background: #f5f7fa;
        border-radius: 4px;
      }
    }
  }
}
</style>

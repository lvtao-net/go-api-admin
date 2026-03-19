import request from './request'

// 获取操作日志列表
export function getLogs(params) {
  return request.get('/manage/logs', { params })
}

// 获取日志统计
export function getLogStats(params) {
  return request.get('/manage/logs/stats', { params })
}

// 清理旧日志
export function deleteOldLogs(params) {
  return request.delete('/manage/logs', { params })
}

// 获取邮件模板列表
export function getEmailTemplates() {
  return request.get('/manage/email-templates')
}

// 获取单个邮件模板
export function getEmailTemplate(type) {
  return request.get(`/manage/email-templates/${type}`)
}

// 更新邮件模板
export function updateEmailTemplate(type, data) {
  return request.patch(`/manage/email-templates/${type}`, data)
}

// 测试邮件模板
export function testEmailTemplate(type, email) {
  return request.post(`/manage/email-templates/${type}/test`, { email })
}

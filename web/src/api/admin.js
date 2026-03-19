import request from './request'

// 管理员登录
export function adminLogin(email, password) {
  return request.post('/admins/auth-with-password', {
    identity: email,
    password,
  })
}

// 刷新 Token
export function refreshToken() {
  return request.post('/admins/auth-refresh')
}

// 获取管理员信息
export function getProfile() {
  return request.get('/admins/profile')
}

// 登出
export function adminLogout() {
  return Promise.resolve()
}

// 获取认证方式
export function getAuthMethods() {
  return request.get('/admins/auth-methods')
}

// 获取统计数据
export function getStats() {
  return request.get('/admins/stats')
}

// 获取管理员列表
export function getAdmins(params) {
  return request.get('/admins', { params })
}

// 创建管理员
export function createAdmin(data) {
  return request.post('/admins', data)
}

// 更新管理员
export function updateAdmin(id, data) {
  return request.patch(`/admins/${id}`, data)
}

// 删除管理员
export function deleteAdmin(id) {
  return request.delete(`/admins/${id}`)
}

// 修改当前管理员密码
export function updateAdminPassword(data) {
  return request.post('/admins/change-password', data)
}

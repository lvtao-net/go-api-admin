import request from './request'

// 获取集合列表
export function getCollections(params = {}) {
  return request.get('/manage/collections', { params })
}

// 获取单个集合
export function getCollection(id) {
  return request.get(`/manage/collections/${id}`)
}

// 创建集合
export function createCollection(data) {
  return request.post('/manage/collections', data)
}

// 更新集合
export function updateCollection(id, data) {
  return request.patch(`/manage/collections/${id}`, data)
}

// 删除集合
export function deleteCollection(id) {
  return request.delete(`/manage/collections/${id}`)
}

// 检查删除集合
export function checkDeleteCollection(id) {
  return request.get(`/manage/collections/${id}/check-delete`)
}

// 获取记录列表
export function getRecords(collection, params = {}) {
  return request.get(`/manage/collections/${collection}/records`, { params })
}

// 获取单条记录
export function getRecord(collection, id) {
  return request.get(`/manage/collections/${collection}/records/${id}`)
}

// 创建记录
export function createRecord(collection, data) {
  return request.post(`/manage/collections/${collection}/records`, data)
}

// 更新记录
export function updateRecord(collection, id, data) {
  return request.patch(`/manage/collections/${collection}/records/${id}`, data)
}

// 删除记录
export function deleteRecord(collection, id) {
  return request.delete(`/manage/collections/${collection}/records/${id}`)
}

// 批量删除记录
export function batchDeleteRecords(collection, ids) {
  return request.post(`/manage/collections/${collection}/records/batch-delete`, { ids })
}

// 获取集合字段
export function getCollectionFields(collection) {
  return request.get(`/manage/collections/${collection}/fields`)
}

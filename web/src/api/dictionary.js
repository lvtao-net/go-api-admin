import request from './request'

// 获取字典列表
export function getDictionaries(params = {}) {
  return request.get('/manage/dictionaries', { params })
}

// 获取单个字典
export function getDictionary(id) {
  return request.get(`/manage/dictionaries/${id}`)
}

// 根据名称获取字典
export function getDictionaryByName(name) {
  return request.get(`/manage/dictionaries/name/${name}`)
}

// 创建字典
export function createDictionary(data) {
  return request.post('/manage/dictionaries', data)
}

// 更新字典
export function updateDictionary(id, data) {
  return request.patch(`/manage/dictionaries/${id}`, data)
}

// 删除字典
export function deleteDictionary(id) {
  return request.delete(`/manage/dictionaries/${id}`)
}

// 获取字典项列表
export function getDictionaryItems(id) {
  return request.get(`/manage/dictionaries/${id}/items`)
}

// 创建字典项
export function createDictionaryItem(id, data) {
  return request.post(`/manage/dictionaries/${id}/items`, data)
}

// 更新字典项
export function updateDictionaryItem(id, itemId, data) {
  return request.patch(`/manage/dictionaries/${id}/items/${itemId}`, data)
}

// 删除字典项
export function deleteDictionaryItem(id, itemId) {
  return request.delete(`/manage/dictionaries/${id}/items/${itemId}`)
}

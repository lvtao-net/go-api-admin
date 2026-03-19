import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { adminLogin, adminLogout, getProfile } from '@/api/admin'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('admin_token') || '')
  const admin = ref(JSON.parse(localStorage.getItem('admin_info') || 'null'))

  const isLoggedIn = computed(() => !!token.value)

  async function login(email, password) {
    try {
      const res = await adminLogin(email, password)
      token.value = res.data.token
      admin.value = res.data.admin
      localStorage.setItem('admin_token', res.data.token)
      localStorage.setItem('admin_info', JSON.stringify(res.data.admin))
      return { success: true }
    } catch (error) {
      return { success: false, message: error.response?.data?.message || '登录失败' }
    }
  }

  async function logout() {
    try {
      await adminLogout()
    } catch (e) {
      // ignore
    }
    token.value = ''
    admin.value = null
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_info')
  }

  async function fetchProfile() {
    try {
      const res = await getProfile()
      admin.value = res.data
      localStorage.setItem('admin_info', JSON.stringify(res.data))
    } catch (error) {
      logout()
    }
  }

  return {
    token,
    admin,
    isLoggedIn,
    login,
    logout,
    fetchProfile
  }
})

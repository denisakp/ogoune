import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authService from '@/services/authService'
import { message } from 'ant-design-vue'

const TOKEN_KEY = 'pulseguard_auth_token'
const EMAIL_KEY = 'pulseguard_user_email'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const email = ref<string | null>(localStorage.getItem(EMAIL_KEY))
  const isLoading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value)

  // Actions
  async function login(emailInput: string, password: string): Promise<boolean> {
    isLoading.value = true
    try {
      const response = await authService.login(emailInput, password)
      
      // Store token and email
      token.value = response.token
      email.value = response.email
      
      localStorage.setItem(TOKEN_KEY, response.token)
      localStorage.setItem(EMAIL_KEY, response.email)
      
      message.success('Successfully logged in!')
      return true
    } catch (error) {
      // Error is already handled by axios interceptor
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function verify(): Promise<boolean> {
    if (!token.value) {
      return false
    }

    try {
      const response = await authService.verify()
      email.value = response.email
      localStorage.setItem(EMAIL_KEY, response.email)
      return true
    } catch (error) {
      // Token is invalid, clear auth state
      logout()
      return false
    }
  }

  function logout() {
    token.value = null
    email.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EMAIL_KEY)
    message.info('You have been logged out')
  }

  function getToken(): string | null {
    return token.value
  }

  return {
    // State
    token,
    email,
    isLoading,
    
    // Getters
    isAuthenticated,
    
    // Actions
    login,
    verify,
    logout,
    getToken,
  }
})

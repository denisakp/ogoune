import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import authService from '@/services/authService'
import { message } from 'ant-design-vue'
import type { User } from '@/types'

const TOKEN_KEY = 'pulseguard_auth_token'
const EMAIL_KEY = 'pulseguard_user_email'
const USER_ID_KEY = 'pulseguard_user_id'

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(localStorage.getItem(TOKEN_KEY))
  const email = ref<string | null>(localStorage.getItem(EMAIL_KEY))
  const userId = ref<string | null>(localStorage.getItem(USER_ID_KEY))
  const user = ref<User | null>(null)
  const isLoading = ref(false)
  const requiresPasswordInit = ref(false)
  const requires2FA = ref(false)
  const pending2FAEmail = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!token.value)

  // Actions
  async function login(emailInput: string, password: string): Promise<boolean> {
    isLoading.value = true
    try {
      const response = await authService.login(emailInput, password)

      // Reset transient flags
      requiresPasswordInit.value = false
      requires2FA.value = false
      pending2FAEmail.value = null

      // Store email for downstream steps
      email.value = response.email
      localStorage.setItem(EMAIL_KEY, response.email)

      // Check if password initialization or 2FA is required
      if (response.force_password_change) {
        requiresPasswordInit.value = true
        message.warning('You must set up your password before continuing')
        return false
      }

      if (response.requires_2fa) {
        requires2FA.value = true
        pending2FAEmail.value = response.email
        token.value = null
        localStorage.removeItem(TOKEN_KEY)
        message.info('Please verify with 2FA')
        return false
      }

      // Store token when fully authenticated
      token.value = response.token
      localStorage.setItem(TOKEN_KEY, response.token)

      message.success('Successfully logged in!')
      return true
    } catch (error) {
      // Error is already handled by axios interceptor
      return false
    } finally {
      isLoading.value = false
    }
  }

  async function verifyTwoFactor(otp: string): Promise<boolean> {
    if (!pending2FAEmail.value) {
      message.error('Session expired. Please log in again.')
      return false
    }

    isLoading.value = true
    try {
      const response = await authService.verify2FA({ email: pending2FAEmail.value, otp })

      token.value = response.token
      email.value = response.email
      localStorage.setItem(TOKEN_KEY, response.token)
      localStorage.setItem(EMAIL_KEY, response.email)

      requires2FA.value = false
      pending2FAEmail.value = null

      // Fetch user details for completeness
      await verify()

      message.success('2FA verified successfully!')
      return true
    } catch (error) {
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
      userId.value = response.user_id
      user.value = response as User

      localStorage.setItem(EMAIL_KEY, response.email)
      localStorage.setItem(USER_ID_KEY, response.user_id)
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
    userId.value = null
    user.value = null
    requiresPasswordInit.value = false
    requires2FA.value = false
    pending2FAEmail.value = null
    localStorage.removeItem(TOKEN_KEY)
    localStorage.removeItem(EMAIL_KEY)
    localStorage.removeItem(USER_ID_KEY)
    message.info('You have been logged out')
  }

  function getToken(): string | null {
    return token.value
  }

  function clearPasswordInitRequired() {
    requiresPasswordInit.value = false
  }

  function clear2FARequired() {
    requires2FA.value = false
    pending2FAEmail.value = null
  }

  return {
    // State
    token,
    email,
    userId,
    user,
    isLoading,
    requiresPasswordInit,
    requires2FA,
    pending2FAEmail,

    // Getters
    isAuthenticated,

    // Actions
    login,
    verify,
    verifyTwoFactor,
    logout,
    getToken,
    clearPasswordInitRequired,
    clear2FARequired,
  }
})

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'

import authService, { type SignupRequest, type ResetPasswordRequest } from '@/services/authService'
import { ValidationError } from '@/core/errors'
import type { User } from '@/types'

const TK = 'ogoune_auth_token', EK = 'ogoune_user_email', UK = 'ogoune_user_id'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem(TK))
  const email = ref<string | null>(localStorage.getItem(EK))
  const userId = ref<string | null>(localStorage.getItem(UK))
  const user = ref<User | null>(null)
  const isLoading = ref(false)
  const requiresPasswordInit = ref(false)
  const requires2FA = ref(false)
  const pending2FAEmail = ref<string | null>(null)
  const isAuthenticated = computed(() => !!token.value)

  function setAuth(t: string, e: string) {
    token.value = t; email.value = e
    localStorage.setItem(TK, t); localStorage.setItem(EK, e)
  }

  async function login(emailInput: string, password: string): Promise<boolean> {
    isLoading.value = true
    try {
      const r = await authService.login(emailInput, password)
      requiresPasswordInit.value = false; requires2FA.value = false; pending2FAEmail.value = null
      email.value = r.email; localStorage.setItem(EK, r.email)
      if (r.force_password_change) {
        requiresPasswordInit.value = true; message.warning('You must set up your password before continuing'); return false
      }
      if (r.requires_2fa) {
        requires2FA.value = true; pending2FAEmail.value = r.email
        token.value = null; localStorage.removeItem(TK); message.info('Please verify with 2FA'); return false
      }
      setAuth(r.token, r.email); message.success('Successfully logged in!'); return true
    } catch (e) {
      if (e instanceof ValidationError) throw e
      return false
    } finally { isLoading.value = false }
  }

  async function verifyTwoFactor(otp: string): Promise<boolean> {
    if (!pending2FAEmail.value) { message.error('Session expired. Please log in again.'); return false }
    isLoading.value = true
    try {
      const r = await authService.verify2FA({ email: pending2FAEmail.value, otp })
      setAuth(r.token, r.email); requires2FA.value = false; pending2FAEmail.value = null
      await verify(); message.success('2FA verified successfully!'); return true
    } catch { return false } finally { isLoading.value = false }
  }

  async function verify(): Promise<boolean> {
    if (!token.value) return false
    try {
      const r = await authService.verify()
      email.value = r.email; userId.value = r.user_id; user.value = r as User
      localStorage.setItem(EK, r.email); localStorage.setItem(UK, r.user_id); return true
    } catch { logout(); return false }
  }

  async function signUp(input: SignupRequest): Promise<boolean> {
    isLoading.value = true
    try {
      const r = await authService.signUp(input)
      setAuth(r.token, r.email)
      message.success('Account created')
      return true
    } catch (e) {
      if (e instanceof ValidationError) throw e
      return false
    } finally { isLoading.value = false }
  }

  async function resetPasswordWithToken(input: ResetPasswordRequest): Promise<boolean> {
    isLoading.value = true
    try {
      const r = await authService.resetPasswordWithToken(input)
      setAuth(r.token, r.email)
      return true
    } catch (e) {
      if (e instanceof ValidationError) throw e
      return false
    } finally { isLoading.value = false }
  }

  function logout() {
    token.value = null; email.value = null; userId.value = null; user.value = null
    requiresPasswordInit.value = false; requires2FA.value = false; pending2FAEmail.value = null
    localStorage.removeItem(TK); localStorage.removeItem(EK); localStorage.removeItem(UK)
    message.info('You have been logged out')
  }

  return {
    token, email, userId, user, isLoading, requiresPasswordInit, requires2FA, pending2FAEmail,
    isAuthenticated, login, signUp, resetPasswordWithToken, verify, verifyTwoFactor, logout,
    getToken: () => token.value,
    clearPasswordInitRequired: () => { requiresPasswordInit.value = false },
    clear2FARequired: () => { requires2FA.value = false; pending2FAEmail.value = null },
  }
})

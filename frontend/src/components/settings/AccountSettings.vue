<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import type { Rule } from 'ant-design-vue/es/form'
import {
  UserOutlined,
  MailOutlined,
  LockOutlined,
  SafetyOutlined,
  LoadingOutlined,
} from '@ant-design/icons-vue'
import QRCode from 'qrcode'
import accountService from '@/services/accountService'

const isLoading = ref(false)

// Profile state
const profileForm = reactive({
  name: '',
  email: '',
})

// Change password state
const changePasswordForm = reactive({
  newPassword: '',
  confirmPassword: '',
})

const currentPasswordForm = reactive({
  currentPassword: '',
})

const isChangePasswordModalOpen = ref(false)
const isChangePasswordSubmitting = ref(false)

// 2FA state
const is2FAEnabled = ref(false)
const enable2FALoading = ref(false)
const showQRCode = ref(false)
const qrCode = ref('')
const pending2FASecret = ref('')
const confirm2FAForm = reactive({
  otp: '',
})

const profileRules: Record<string, Rule[]> = {
  name: [{ required: true, message: 'Please enter your name!', trigger: 'blur' }],
  email: [
    { required: true, message: 'Please enter your email!', trigger: 'blur' },
    { type: 'email', message: 'Please enter a valid email!', trigger: 'blur' },
  ],
}

const changePasswordRules: Record<string, Rule[]> = {
  newPassword: [
    { required: true, message: 'Please enter your new password!', trigger: 'blur' },
    { min: 8, message: 'Password must be at least 8 characters!', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: 'Please confirm your password!', trigger: 'blur' },
    {
      validator: (_rule: Rule, value: string) => {
        if (value && value !== changePasswordForm.newPassword) {
          return Promise.reject(new Error('Passwords do not match!'))
        }
        return Promise.resolve()
      },
      trigger: 'blur',
    },
  ],
}

const confirm2FARules: Record<string, Rule[]> = {
  otp: [
    { required: true, message: 'Please enter the 6-digit code!', trigger: 'blur' },
    { len: 6, message: 'OTP must be 6 digits!', trigger: 'blur' },
  ],
}

onMounted(async () => {
  await loadProfile()
})

async function loadProfile() {
  isLoading.value = true
  try {
    const profile = await accountService.getProfile()
    profileForm.name = profile.name
    profileForm.email = profile.email
    is2FAEnabled.value = profile.two_factor_enabled
  } catch (error) {
    message.error('Failed to load profile')
  } finally {
    isLoading.value = false
  }
}

async function handleUpdateProfile() {
  try {
    const updatedProfile = await accountService.updateProfile(profileForm.name, profileForm.email)
    profileForm.name = updatedProfile.name
    profileForm.email = updatedProfile.email
    message.success('Profile updated successfully')
  } catch (error) {
    message.error('Failed to update profile')
  }
}

function openChangePasswordModal() {
  if (!changePasswordForm.newPassword || !changePasswordForm.confirmPassword) {
    message.error('Please fill in the new password fields')
    return
  }

  if (changePasswordForm.newPassword !== changePasswordForm.confirmPassword) {
    message.error('Passwords do not match')
    return
  }

  isChangePasswordModalOpen.value = true
}

async function submitChangePassword() {
  if (!currentPasswordForm.currentPassword) {
    message.error('Please enter your current password')
    return
  }

  isChangePasswordSubmitting.value = true
  try {
    await accountService.changePassword(
      currentPasswordForm.currentPassword,
      changePasswordForm.newPassword,
    )
    currentPasswordForm.currentPassword = ''
    changePasswordForm.newPassword = ''
    changePasswordForm.confirmPassword = ''
    isChangePasswordModalOpen.value = false
    message.success('Password changed successfully')
  } catch (error) {
    message.error('Failed to change password')
  } finally {
    isChangePasswordSubmitting.value = false
  }
}

async function handleEnable2FA() {
  enable2FALoading.value = true
  try {
    const response = await accountService.enable2FA()
    if (response.qr_code) {
      qrCode.value = await QRCode.toDataURL(response.qr_code)
      pending2FASecret.value = response.secret
      showQRCode.value = true
      message.info('Please scan the QR code with your authenticator app')
    } else {
      message.error('No QR code received from server')
    }
  } catch (error) {
    message.error('Failed to enable 2FA')
  } finally {
    enable2FALoading.value = false
  }
}

async function handleConfirm2FA() {
  if (!confirm2FAForm.otp) {
    message.error('Please enter the OTP')
    return
  }

  if (!pending2FASecret.value) {
    message.error('Missing secret, please re-enable 2FA')
    return
  }

  try {
    await accountService.confirm2FA({ otp: confirm2FAForm.otp, secret: pending2FASecret.value })
    is2FAEnabled.value = true
    showQRCode.value = false
    confirm2FAForm.otp = ''
    pending2FASecret.value = ''
    message.success('Two-factor authentication enabled successfully')
  } catch (error) {
    message.error('Invalid OTP. Please try again')
  }
}

async function handleDisable2FA() {
  const otp = prompt('Enter your 6-digit authenticator code:')
  if (!otp) {
    return
  }

  try {
    await accountService.disable2FA(otp)
    is2FAEnabled.value = false
    message.success('Two-factor authentication disabled')
  } catch (error) {
    message.error('Failed to disable 2FA')
  }
}
</script>

<template>
  <div class="account-settings">
    <a-card :bordered="false" title="Account Settings" class="settings-card">
      <a-skeleton v-if="isLoading" active />

      <div v-else class="settings-content">
        <section class="settings-section">
          <div class="section-header">
            <h3>Profile</h3>
            <p class="section-description">Update your personal information.</p>
          </div>

          <a-form :model="profileForm" :rules="profileRules" layout="vertical">
            <a-row :gutter="16">
              <a-col :xs="24" :md="12">
                <a-form-item label="Name" name="name">
                  <a-input v-model:value="profileForm.name" placeholder="Your name" size="large">
                    <template #prefix>
                      <UserOutlined />
                    </template>
                  </a-input>
                </a-form-item>
              </a-col>

              <a-col :xs="24" :md="12">
                <a-form-item label="Email" name="email">
                  <a-input
                    v-model:value="profileForm.email"
                    placeholder="your@email.com"
                    size="large"
                  >
                    <template #prefix>
                      <MailOutlined />
                    </template>
                  </a-input>
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item>
              <a-button type="primary" size="large" @click="handleUpdateProfile">
                Update Profile
              </a-button>
            </a-form-item>
          </a-form>
        </section>

        <a-divider />

        <section class="settings-section">
          <div class="section-header">
            <h3>Password</h3>
            <p class="section-description">Change your password to keep your account secure.</p>
          </div>

          <a-form :model="changePasswordForm" :rules="changePasswordRules" layout="vertical">
            <a-row :gutter="16">
              <a-col :xs="24" :md="12">
                <a-form-item label="New Password" name="newPassword">
                  <a-input-password
                    v-model:value="changePasswordForm.newPassword"
                    placeholder="Enter your new password"
                    size="large"
                  >
                    <template #prefix>
                      <LockOutlined />
                    </template>
                  </a-input-password>
                </a-form-item>
              </a-col>

              <a-col :xs="24" :md="12">
                <a-form-item label="Confirm New Password" name="confirmPassword">
                  <a-input-password
                    v-model:value="changePasswordForm.confirmPassword"
                    placeholder="Confirm your new password"
                    size="large"
                  >
                    <template #prefix>
                      <LockOutlined />
                    </template>
                  </a-input-password>
                </a-form-item>
              </a-col>
            </a-row>

            <a-form-item>
              <a-button type="primary" size="large" @click="openChangePasswordModal">
                Change Password
              </a-button>
            </a-form-item>
          </a-form>

          <a-modal
            v-model:open="isChangePasswordModalOpen"
            title="Confirm Current Password"
            :confirm-loading="isChangePasswordSubmitting"
            ok-text="Confirm"
            @ok="submitChangePassword"
            @cancel="isChangePasswordModalOpen = false"
          >
            <p class="section-description">Enter your current password to apply the new one.</p>
            <a-form layout="vertical">
              <a-form-item label="Current Password">
                <a-input-password
                  v-model:value="currentPasswordForm.currentPassword"
                  placeholder="Enter your current password"
                  size="large"
                >
                  <template #prefix>
                    <LockOutlined />
                  </template>
                </a-input-password>
              </a-form-item>
            </a-form>
          </a-modal>
        </section>

        <a-divider />

        <section class="settings-section">
          <div class="section-header">
            <h3>Two-Factor Authentication</h3>
            <p class="section-description">
              Add an extra layer of security with a one-time code from your authenticator app.
            </p>
          </div>

          <a-alert
            v-if="is2FAEnabled"
            message="Two-factor authentication is enabled"
            type="success"
            show-icon
            class="twofa-alert"
          />
          <a-alert
            v-else
            message="Two-factor authentication is disabled"
            type="warning"
            show-icon
            class="twofa-alert"
          />

          <div v-if="!is2FAEnabled" class="enable-2fa">
            <p class="description">
              Two-factor authentication adds an extra layer of security to your account.
            </p>
            <a-button
              type="primary"
              size="large"
              :loading="enable2FALoading"
              @click="handleEnable2FA"
            >
              <SafetyOutlined />
              Enable 2FA
            </a-button>

            <div v-if="showQRCode" class="qr-code-section">
              <p>Scan this QR code with your authenticator app:</p>
              <div class="qr-code-placeholder">
                <img v-if="qrCode" :src="qrCode" alt="2FA QR Code" />
                <div v-else class="qr-loading">
                  <LoadingOutlined />
                </div>
              </div>

              <a-form :model="confirm2FAForm" :rules="confirm2FARules" layout="vertical">
                <a-form-item label="Verification Code" name="otp">
                  <a-input
                    v-model:value="confirm2FAForm.otp"
                    placeholder="000000"
                    size="large"
                    maxlength="6"
                  />
                </a-form-item>

                <a-form-item>
                  <a-button type="primary" size="large" @click="handleConfirm2FA">
                    Verify and Enable
                  </a-button>
                </a-form-item>
              </a-form>
            </div>
          </div>

          <div v-else class="disable-2fa">
            <p class="description">Your account is protected with two-factor authentication.</p>
            <a-button type="primary" danger size="large" @click="handleDisable2FA">
              Disable 2FA
            </a-button>
          </div>
        </section>
      </div>
    </a-card>
  </div>
</template>

<style scoped>
.account-settings {
  padding: 0;
  background: transparent;
}

.settings-card {
  box-shadow: none;
  border: 1px solid #f0f0f0;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.settings-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.section-description {
  margin: 4px 0 0;
  color: rgba(0, 0, 0, 0.65);
  line-height: 1.5;
}

.twofa-alert {
  margin: 8px 0 16px;
}

.description {
  color: #666;
  margin: 0 0 12px;
  line-height: 1.6;
}

.enable-2fa,
.disable-2fa {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.qr-code-section {
  margin-top: 8px;
  padding: 20px;
  background: #f5f5f5;
  border-radius: 8px;
}

.qr-code-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 220px;
  height: 220px;
  margin: 12px auto 16px;
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.qr-code-placeholder img {
  max-width: 100%;
  max-height: 100%;
}

.qr-loading {
  font-size: 28px;
  color: #999;
}

:deep(.ant-form) {
  max-width: none;
  width: 100%;
}

@media (max-width: 600px) {
  .account-settings {
    padding: 16px;
  }

  .qr-code-placeholder {
    width: 200px;
    height: 200px;
  }
}
</style>

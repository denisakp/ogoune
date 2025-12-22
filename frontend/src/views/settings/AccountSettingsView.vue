<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { message } from 'ant-design-vue'
import {
  UserOutlined,
  MailOutlined,
  LockOutlined,
  SafetyOutlined,
  LoadingOutlined,
} from '@ant-design/icons-vue'
import accountService from '@/services/accountService'
import type { Rule } from 'ant-design-vue/es/form'

const isLoading = ref(false)
const activeTab = ref('profile')

// Profile state
const profileForm = reactive({
  name: '',
  email: '',
})

// Change password state
const changePasswordForm = reactive({
  currentPassword: '',
  newPassword: '',
  confirmPassword: '',
})

// 2FA state
const is2FAEnabled = ref(false)
const enable2FALoading = ref(false)
const showQRCode = ref(false)
const qrCode = ref('')
//const backupCodes = ref<string[]>([])
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
  currentPassword: [
    { required: true, message: 'Please enter your current password!', trigger: 'blur' },
  ],
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

// Load profile
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

async function handleChangePassword() {
  if (!changePasswordForm.currentPassword || !changePasswordForm.newPassword) {
    message.error('Please fill in all fields')
    return
  }

  if (changePasswordForm.newPassword !== changePasswordForm.confirmPassword) {
    message.error('Passwords do not match')
    return
  }

  try {
    await accountService.changePassword(
      changePasswordForm.currentPassword,
      changePasswordForm.newPassword,
    )
    changePasswordForm.currentPassword = ''
    changePasswordForm.newPassword = ''
    changePasswordForm.confirmPassword = ''
    message.success('Password changed successfully')
  } catch (error) {
    message.error('Failed to change password')
  }
}

async function handleEnable2FA() {
  enable2FALoading.value = true
  try {
    const response = await accountService.enable2FA()
    qrCode.value = response.qr_code
    showQRCode.value = true
    message.info('Please scan the QR code with your authenticator app')
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

  try {
    await accountService.confirm2FA(confirm2FAForm.otp)
    is2FAEnabled.value = true
    showQRCode.value = false
    confirm2FAForm.otp = ''
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
  <div class="account-settings-container">
    <a-card title="Account Settings" class="settings-card">
      <a-skeleton v-if="isLoading" active />

      <a-tabs v-else v-model:activeKey="activeTab" type="card">
        <!-- Profile Tab -->
        <a-tab-pane key="profile" tab="Profile">
          <div class="profile-section">
            <a-form :model="profileForm" :rules="profileRules" layout="vertical">
              <a-form-item label="Name" name="name">
                <a-input v-model:value="profileForm.name" placeholder="Your name" size="large">
                  <template #prefix>
                    <UserOutlined />
                  </template>
                </a-input>
              </a-form-item>

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

              <a-form-item>
                <a-button type="primary" size="large" @click="handleUpdateProfile">
                  Update Profile
                </a-button>
              </a-form-item>
            </a-form>
          </div>
        </a-tab-pane>

        <!-- Password Tab -->
        <a-tab-pane key="password" tab="Password">
          <div class="password-section">
            <a-form :model="changePasswordForm" :rules="changePasswordRules" layout="vertical">
              <a-form-item label="Current Password" name="currentPassword">
                <a-input-password
                  v-model:value="changePasswordForm.currentPassword"
                  placeholder="Enter your current password"
                  size="large"
                >
                  <template #prefix>
                    <LockOutlined />
                  </template>
                </a-input-password>
              </a-form-item>

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

              <a-form-item>
                <a-button type="primary" size="large" @click="handleChangePassword">
                  Change Password
                </a-button>
              </a-form-item>
            </a-form>
          </div>
        </a-tab-pane>

        <!-- 2FA Tab -->
        <a-tab-pane key="2fa" tab="Two-Factor Authentication">
          <div class="twofa-section">
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
              <a-button type="danger" size="large" @click="handleDisable2FA">
                Disable 2FA
              </a-button>
            </div>
          </div>
        </a-tab-pane>
      </a-tabs>
    </a-card>
  </div>
</template>
<style scoped>
.account-settings-container {
  padding: 24px;
  max-width: 900px;
  margin: 0 auto;
}

.settings-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.profile-section,
.password-section,
.twofa-section {
  padding: 24px 0;
}

.twofa-alert {
  margin-bottom: 24px;
}

.description {
  color: #666;
  margin-bottom: 16px;
  line-height: 1.6;
}

.enable-2fa,
.disable-2fa {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.qr-code-section {
  margin-top: 24px;
  padding: 24px;
  background: #f5f5f5;
  border-radius: 8px;
}

.qr-code-placeholder {
  display: flex;
  justify-content: center;
  align-items: center;
  width: 250px;
  height: 250px;
  margin: 16px auto;
  background: white;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.qr-code-placeholder img {
  max-width: 100%;
  max-height: 100%;
}

.qr-loading {
  font-size: 32px;
  color: #999;
}

:deep(.ant-form) {
  max-width: 400px;
}

@media (max-width: 600px) {
  .account-settings-container {
    padding: 16px;
  }

  .qr-code-placeholder {
    width: 200px;
    height: 200px;
  }
}
</style>

import { reactive, ref, watch } from 'vue'
import type {
  IntegrationType,
  IntegrationConfig,
  SlackConfig,
  DiscordConfig,
  GoogleChatConfig,
  WebhookConfig,
} from '@/types'

/**
 * Composable to manage integration-specific configurations
 */
export function useIntegrationConfig() {
  const integrationType = ref<IntegrationType>('slack')

  // Slack config state
  const slackConfig = reactive<SlackConfig>({
    type: 'slack',
    webhook_url: '',
    channel: '',
    username: '',
  })

  // Discord config state
  const discordConfig = reactive<DiscordConfig>({
    type: 'discord',
    webhook_url: '',
    channel: '',
  })

  // Google Chat config state
  const googleChatConfig = reactive<GoogleChatConfig>({
    type: 'googlechat',
    webhook_url: '',
    thread_key: '',
  })

  // Webhook config state
  const webhookConfig = reactive<WebhookConfig>({
    type: 'webhook',
    url: '',
    method: 'POST',
    headers: {},
    auth_type: 'none',
    auth_token: '',
  })

  /**
   * Get current configuration based on integration type
   */
  const getCurrentConfig = (): IntegrationConfig => {
    switch (integrationType.value) {
      case 'slack':
        return { ...slackConfig }
      case 'discord':
        return { ...discordConfig }
      case 'googlechat':
        return { ...googleChatConfig }
      case 'webhook':
        return { ...webhookConfig }
      default:
        return { type: 'slack' } as IntegrationConfig
    }
  }

  /**
   * Set configuration from existing integration
   */
  const setConfigFromIntegration = (config: IntegrationConfig) => {
    const baseType = config.type
    integrationType.value = baseType as IntegrationType

    switch (baseType) {
      case 'slack': {
        const slackCfg = config as SlackConfig
        Object.assign(slackConfig, slackCfg)
        break
      }
      case 'discord': {
        const discordCfg = config as DiscordConfig
        Object.assign(discordConfig, discordCfg)
        break
      }
      case 'googlechat': {
        const googleChatCfg = config as GoogleChatConfig
        Object.assign(googleChatConfig, googleChatCfg)
        break
      }
      case 'webhook': {
        const webhookCfg = config as WebhookConfig
        Object.assign(webhookConfig, webhookCfg)
        break
      }
    }
  }

  /**
   * Reset all configurations
   */
  const resetConfig = () => {
    slackConfig.webhook_url = ''
    slackConfig.channel = ''
    slackConfig.username = ''

    discordConfig.webhook_url = ''
    discordConfig.channel = ''

    googleChatConfig.webhook_url = ''
    googleChatConfig.thread_key = ''

    webhookConfig.url = ''
    webhookConfig.method = 'POST'
    webhookConfig.headers = {}
    webhookConfig.auth_type = 'none'
    webhookConfig.auth_token = ''
  }

  /**
   * Validate current configuration
   */
  const validateConfig = (): { valid: boolean; errors: string[] } => {
    const errors: string[] = []

    switch (integrationType.value) {
      case 'slack': {
        if (!slackConfig.webhook_url.trim()) {
          errors.push('Slack webhook URL is required')
        }
        if (!slackConfig.webhook_url.startsWith('https://hooks.slack.com/')) {
          errors.push('Invalid Slack webhook URL format')
        }
        break
      }
      case 'discord': {
        if (!discordConfig.webhook_url.trim()) {
          errors.push('Discord webhook URL is required')
        }
        if (!discordConfig.webhook_url.startsWith('https://')) {
          errors.push('Invalid Discord webhook URL format')
        }
        break
      }
      case 'googlechat': {
        if (!googleChatConfig.webhook_url.trim()) {
          errors.push('Google Chat webhook URL is required')
        }
        if (!googleChatConfig.webhook_url.startsWith('https://')) {
          errors.push('Invalid Google Chat webhook URL format')
        }
        break
      }
      case 'webhook': {
        if (!webhookConfig.url.trim()) {
          errors.push('Webhook URL is required')
        }
        if (!webhookConfig.url.startsWith('http')) {
          errors.push('Webhook URL must start with http or https')
        }
        if (webhookConfig.auth_type !== 'none' && !webhookConfig.auth_token?.trim()) {
          errors.push('Authentication token is required when auth type is not "none"')
        }
        break
      }
    }

    return {
      valid: errors.length === 0,
      errors,
    }
  }

  return {
    integrationType,
    slackConfig,
    discordConfig,
    googleChatConfig,
    webhookConfig,
    getCurrentConfig,
    setConfigFromIntegration,
    resetConfig,
    validateConfig,
  }
}

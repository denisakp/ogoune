import { z } from 'zod'

/**
 * Spec 059 US3 — discriminated channel schema.
 * Backend dispatch (internal/monitoring/incident_service.go) currently
 * handles `smtp` + `webhook`; `slack` is recognised by the domain enum but
 * not yet routed. Discord and Telegram are spec-idealised — deferred.
 */

const baseFields = {
  name: z.string().trim().min(1, 'Required').max(80, 'At most 80 characters'),
  is_default: z.boolean().default(false),
  is_active: z.boolean().default(true),
}

export const smtpChannelSchema = z.object({
  type: z.literal('smtp'),
  ...baseFields,
  config: z.object({
    host: z.string().trim().min(1, 'SMTP host required'),
    port: z.coerce.number().int().min(1).max(65535),
    username: z.string().trim().min(1, 'Username required'),
    password: z.string().min(1, 'Password required'),
    sender: z.string().trim().email('Sender must be an email'),
    recipient: z.string().trim().email('Recipient must be an email'),
  }),
})

export const slackChannelSchema = z.object({
  type: z.literal('slack'),
  ...baseFields,
  config: z.object({
    webhook_url: z
      .string()
      .url('Must be a valid URL')
      .startsWith('https://hooks.slack.com/', 'Must be a Slack incoming-webhook URL'),
    channel: z
      .string()
      .trim()
      .regex(/^#?[a-z0-9-_]+$/, 'Lowercase letters, digits, dashes, underscores'),
    display_name: z.string().max(80).optional(),
  }),
})

export const webhookChannelSchema = z.object({
  type: z.literal('webhook'),
  ...baseFields,
  config: z.object({
    url: z.string().url('Must be a valid URL'),
    method: z.enum(['POST', 'PUT']).default('POST'),
    headers: z
      .array(
        z.object({
          name: z.string().trim().min(1).max(80),
          value: z.string().min(1).max(1024),
        }),
      )
      .max(20, 'At most 20 headers')
      .default([]),
  }),
})

export const notificationChannelSchema = z.discriminatedUnion('type', [
  smtpChannelSchema,
  slackChannelSchema,
  webhookChannelSchema,
])

export type NotificationChannelInput = z.infer<typeof notificationChannelSchema>
export type ChannelType = NotificationChannelInput['type']

export const CHANNEL_TYPES: { value: ChannelType; label: string; icon: string }[] = [
  { value: 'smtp', label: 'Email (SMTP)', icon: 'i-lucide-mail' },
  { value: 'slack', label: 'Slack', icon: 'i-lucide-message-square' },
  { value: 'webhook', label: 'Webhook', icon: 'i-lucide-webhook' },
]

export function emptyConfigForType(type: ChannelType): NotificationChannelInput['config'] {
  switch (type) {
    case 'smtp':
      return { host: '', port: 587, username: '', password: '', sender: '', recipient: '' }
    case 'slack':
      return { webhook_url: '', channel: '', display_name: '' }
    case 'webhook':
      return { url: '', method: 'POST', headers: [] }
  }
}

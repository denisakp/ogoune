import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import BrandingSection from '../BrandingSection.vue'
import type { StatusPageSettingsResponse } from '@/types'

vi.mock('@/services/statusPageSettingsService', () => ({
  uploadStatusPageLogo: vi.fn(),
  deleteStatusPageLogo: vi.fn(),
}))
vi.mock('ant-design-vue', () => ({
  message: { success: vi.fn(), error: vi.fn() },
}))

import * as svc from '@/services/statusPageSettingsService'

function mkSettings(overrides: Partial<StatusPageSettingsResponse> = {}): StatusPageSettingsResponse {
  return {
    id: 'sp-1',
    name: 'Acme',
    homepage_url: '',
    custom_domain: '',
    google_analytics_id: '',
    enable_details_page: true,
    show_uptime_percentage: true,
    hide_paused_monitors: true,
    show_incident_history: true,
    custom_domain_status: 'pending',
    custom_domain_ssl_status: 'none',
    custom_domain_dns_records: [],
    logo_url_light: '',
    logo_url_dark: '',
    favicon_url: '',
    primary_color: '#4f46e5',
    theme_overrides: {},
    created_at: '',
    updated_at: '',
    ...overrides,
  }
}

function render(overrides: Partial<StatusPageSettingsResponse> = {}) {
  const settings = mkSettings(overrides)
  return mount(BrandingSection, {
    props: {
      settings,
      primaryColor: settings.primary_color,
      themeOverrides: settings.theme_overrides,
    },
  })
}

describe('BrandingSection — US5', () => {
  beforeEach(() => vi.clearAllMocks())

  it('renders 3 logo slots, primary color picker, and theme overrides editor', () => {
    const w = render()
    expect(w.findAll('[data-slot]')).toHaveLength(3)
    expect(w.find('[data-testid="primary-color-picker"]').exists()).toBe(true)
    expect(w.find('[data-testid="theme-overrides-editor"]').exists()).toBe(true)
  })

  it('picking a primary swatch emits update:primaryColor', async () => {
    const w = render()
    await w.get('[data-testid="swatch-#10b981"]').trigger('click')
    expect(w.emitted('update:primaryColor')).toBeTruthy()
    expect(w.emitted('update:primaryColor')?.[0]).toEqual(['#10b981'])
  })

  it('uploading a valid PNG calls uploadStatusPageLogo and emits settings-refreshed', async () => {
    const next = mkSettings({ logo_url_light: '/static/uploads/statuspage/light-x.png' })
    vi.mocked(svc.uploadStatusPageLogo).mockResolvedValue(next)
    const w = render()
    const file = new File([new Uint8Array([1, 2, 3])], 'logo.png', { type: 'image/png' })
    const lightInput = w.findAll('[data-slot]')[0]?.find('[data-testid="file-input"]')
    Object.defineProperty(lightInput!.element, 'files', { value: [file] })
    await lightInput!.trigger('change')
    await flushPromises()
    expect(svc.uploadStatusPageLogo).toHaveBeenCalledWith('light', file)
    expect(w.emitted('settings-refreshed')?.[0]).toEqual([next])
  })

  it('changing a theme color emits update:themeOverrides', async () => {
    const w = render()
    const colorInput = w.get('[data-testid="color---status-up"]')
    await colorInput.setValue('#10b981')
    await colorInput.trigger('input')
    expect(w.emitted('update:themeOverrides')).toBeTruthy()
    const last = w.emitted('update:themeOverrides')?.at(-1)
    expect(last?.[0]).toMatchObject({ '--status-up': '#10b981' })
  })

  it('invalid hex in custom field does not emit until valid', async () => {
    const w = render()
    const input = w.get('[data-testid="hex-input"]')
    await input.setValue('not-a-hex')
    await input.trigger('input')
    // No emission because regex fails.
    expect((w.emitted('update:primaryColor') ?? []).length).toBe(0)
    await input.setValue('#abcdef')
    await input.trigger('input')
    expect(w.emitted('update:primaryColor')?.at(-1)).toEqual(['#abcdef'])
  })
})

import { describe, expect, it, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import AppLayout from './AppLayout.vue'
import { useAnnouncementStore } from '@/stores/announcementStore'
import { __setAnnouncementsFeedForTests } from '@/services/announcementsService'

function mountLayout() {
  return mount(AppLayout, {
    slots: { default: '<div data-test="slot">page-content</div>' },
    global: {
      stubs: {
        AppSidebar: { template: '<aside data-test="sidebar" />' },
        AppTopbar: { template: '<header data-test="topbar" />' },
        OnboardingWizardModal: true,
      },
    },
  })
}

describe('AppLayout', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    // Deterministic boot-fetch: no remote banners unless a test publishes.
    __setAnnouncementsFeedForTests({ fetchActive: async () => [] })
  })

  it('mounts and renders the default slot inside the main column', () => {
    const wrapper = mountLayout()
    expect(wrapper.find('[data-test="sidebar"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="topbar"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="slot"]').text()).toBe('page-content')
  })

  it('shows no announcement banner by default', () => {
    const wrapper = mountLayout()
    expect(wrapper.find('[data-testid="announcement-banner"]').exists()).toBe(false)
  })

  it('renders the active announcement banner and dismisses it', async () => {
    const wrapper = mountLayout()
    const store = useAnnouncementStore()
    store.publish({
      id: 'maint-1',
      severity: 'warning',
      title: 'Scheduled maintenance',
      dismissible: true,
    })
    await nextTick()

    const banner = wrapper.find('[data-testid="announcement-banner"]')
    expect(banner.exists()).toBe(true)
    expect(banner.text()).toContain('Scheduled maintenance')

    // UAlert renders a close button (aria-label="Close") when :close is set.
    await banner.find('button[aria-label="Close"]').trigger('click')
    await nextTick()
    expect(wrapper.find('[data-testid="announcement-banner"]').exists()).toBe(false)
    expect(store.dismissed.has('maint-1')).toBe(true)
  })
})

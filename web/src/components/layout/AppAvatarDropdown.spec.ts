import { describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { mount } from '@vue/test-utils'
import AppAvatarDropdown from './AppAvatarDropdown.vue'
import { useAuthStore } from '@/stores/authStore'

const pushMock = vi.fn()
vi.mock('vue-router', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-router')>()
  return {
    ...actual,
    useRouter: () => ({ push: pushMock }),
  }
})

describe('AppAvatarDropdown', () => {
  it('mounts and exposes user initials from the auth store', () => {
    setActivePinia(createPinia())
    const store = useAuthStore()
    store.email = 'user@example.com'

    const wrapper = mount(AppAvatarDropdown, {
      global: { stubs: { UDropdownMenu: { template: '<div><slot /></div>' } } },
    })
    expect(wrapper.find('[aria-label^="Open user menu"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('US')
  })

  it('declares the menu entries in documented order', () => {
    setActivePinia(createPinia())
    const wrapper = mount(AppAvatarDropdown, {
      global: { stubs: { UDropdownMenu: { template: '<div />' } } },
    })
    const exposed = wrapper.vm as unknown as {
      getItems: () => Array<Array<{ label: string }>>
    }
    const labels = exposed
      .getItems()
      .flat()
      .map((i) => i.label)
    expect(labels).toEqual([
      'Profile',
      'Keyboard shortcuts',
      'Documentation',
      "What's new",
      'Sign out',
    ])
  })
})

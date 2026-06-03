import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn(), resolve: () => ({ href: '#' }) }),
  useRoute: () => ({ query: {}, params: {}, path: '/overview', name: 'overview' }),
  useLink: () => ({ href: { value: '#' }, navigate: vi.fn(), isActive: { value: false } }),
  RouterLink: { template: '<a><slot /></a>' },
}))

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    resources: [],
    loading: false,
    loadResources: vi.fn().mockResolvedValue(undefined),
  }),
}))

vi.mock('@/services/statsService', () => ({
  fetchStatsSummary: vi.fn().mockResolvedValue({}),
}))
vi.mock('@/services/activityService', () => ({
  fetchActivities: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/components/overview/HeroCard.vue', () => ({
  default: { name: 'HeroCard', template: '<div class="hero" />' },
}))
vi.mock('@/components/overview/SecondaryStats.vue', () => ({
  default: { name: 'SecondaryStats', template: '<div class="secondary" />' },
}))
vi.mock('@/components/overview/StatusBreakdown.vue', () => ({
  default: { name: 'StatusBreakdown', template: '<div class="breakdown" />' },
}))
vi.mock('@/components/overview/RecentActivity.vue', () => ({
  default: { name: 'RecentActivity', template: '<div class="activity" />' },
}))
vi.mock('@/components/ResponseTimeChart.vue', () => ({
  default: { name: 'ResponseTimeChart', template: '<div class="chart" />' },
}))

import OverviewView from './OverviewView.vue'

const stubs = {
  UButton: { template: '<button><slot /></button>' },
  UIcon: { template: '<span />' },
}

function build() {
  setActivePinia(createPinia())
  return mount(OverviewView, { global: { stubs } })
}

beforeEach(() => {})

describe('OverviewView', () => {
  it('renders the 5 documented sections', () => {
    const w = build()
    expect(w.findComponent({ name: 'HeroCard' }).exists()).toBe(true)
    expect(w.findComponent({ name: 'SecondaryStats' }).exists()).toBe(true)
    expect(w.findComponent({ name: 'ResponseTimeChart' }).exists()).toBe(true)
    expect(w.findComponent({ name: 'StatusBreakdown' }).exists()).toBe(true)
    expect(w.findComponent({ name: 'RecentActivity' }).exists()).toBe(true)
  })

  it('renders page header title + subtitle + 2 quick actions', () => {
    const w = build()
    expect(w.text()).toContain('Overview')
    expect(w.text()).toContain('Live view across all resources')
    expect(w.findAll('button').length).toBeGreaterThanOrEqual(2)
  })

  it('dark-mode artifact check: root carries bg-default token (FR-017)', () => {
    document.documentElement.classList.add('dark')
    const w = build()
    expect(w.find('.bg-default').exists()).toBe(true)
    document.documentElement.classList.remove('dark')
  })
})

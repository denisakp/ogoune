import { defineComponent, ref } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import ResourceView from '@/views/resources/ResourceView.vue'
import type { Resource } from '@/types'

const { loadResourceWithResponseTimesMock, pauseResourceMock, messageErrorMock } = vi.hoisted(
  () => ({
    loadResourceWithResponseTimesMock: vi.fn(),
    pauseResourceMock: vi.fn(),
    messageErrorMock: vi.fn(),
  }),
)

vi.mock('vue-router', () => ({
  useRouter: () => ({ back: vi.fn() }),
  useRoute: () => ({ params: { id: 'hb-1' } }),
}))

vi.mock('ant-design-vue', () => ({
  message: {
    error: messageErrorMock,
    success: vi.fn(),
  },
}))

vi.mock('@/libs/date-time.helper', () => ({
  getTimeRangeCutoff: () => new Date('2025-01-01T00:00:00.000Z'),
  timeAgo: () => '',
}))

vi.mock('@/stores/resourceStore', () => ({
  useResourceStore: () => ({
    fetchLoading: ref(false),
    pauseMonitoring: pauseResourceMock,
    loadResourceWithResponseTimes: loadResourceWithResponseTimesMock,
    $id: 'resource',
  }),
}))

vi.mock('@/composables/useMonitorLive', () => ({
  useMonitorLive: () => ({
    liveData: ref(null),
    isLoading: ref(false),
    lastUpdated: ref(null),
    error: ref(null),
    isTerminated: ref(false),
    refresh: vi.fn(),
    startPolling: vi.fn(),
    stopPolling: vi.fn(),
  }),
}))

vi.mock('@/components/resources/ResourceModal.vue', () => ({
  default: defineComponent({
    name: 'ResourceModal',
    template: '<div data-testid="resource-modal" />',
  }),
}))

vi.mock('@/components/ResponseTimeChart.vue', () => ({
  default: defineComponent({
    name: 'ResponseTimeChart',
    template: '<div data-testid="response-chart" />',
  }),
}))

const buildHeartbeat = (overrides: Partial<Resource> = {}): Resource => ({
  id: 'hb-1',
  name: 'Nightly Backup',
  type: 'heartbeat',
  target: '',
  interval: 300,
  timeout: 10,
  status: 'waiting',
  is_active: true,
  failure_count: 0,
  confirmation_checks: 1,
  confirmation_interval: 30,
  created_at: '2026-01-01T00:00:00.000Z',
  updated_at: '2026-01-01T00:00:00.000Z',
  tags: [],
  incidents: [],
  response_times: [],
  waiting: true,
  last_ping_at: null,
  heartbeat_interval: 300,
  heartbeat_grace: 60,
  heartbeat_slug: '550e8400-e29b-41d4-a716-446655440000',
  ...overrides,
})

const mountView = () =>
  mount(ResourceView, {
    global: {
      stubs: {
        ResourceModal: true,
        ResponseTimeChart: true,
      },
    },
  })

describe('ResourceView — heartbeat badge and last ping display (T043)', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    loadResourceWithResponseTimesMock.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows HEARTBEAT type in header', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('HEARTBEAT')
    wrapper.unmount()
  })

  it('shows waiting alert when monitor has never been pinged', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="heartbeat-waiting-alert"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Waiting for first ping')
    wrapper.unmount()
  })

  it('does not show waiting alert when monitor is up', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildHeartbeat({ status: 'up', waiting: false, last_ping_at: '2026-03-31T10:00:00.000Z' }),
    )
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="heartbeat-waiting-alert"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows last ping at when available', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildHeartbeat({ status: 'up', waiting: false, last_ping_at: '2026-03-31T10:00:00.000Z' }),
    )
    const wrapper = mountView()
    await flushPromises()

    const lastPing = wrapper.find('[data-testid="last-ping-at"]')
    expect(lastPing.exists()).toBe(true)
    expect(lastPing.text()).not.toBe('Never')
    wrapper.unmount()
  })

  it('shows Never for last ping when monitor is in waiting state', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    const lastPing = wrapper.find('[data-testid="last-ping-at"]')
    expect(lastPing.exists()).toBe(true)
    expect(lastPing.text()).toBe('Never')
    wrapper.unmount()
  })

  it('shows down status text for a down heartbeat monitor', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildHeartbeat({ status: 'down', waiting: false, last_ping_at: '2026-03-30T10:00:00.000Z' }),
    )
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('Down')
    wrapper.unmount()
  })
})

describe('ResourceView — heartbeat integration snippet visible by default (T045)', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    loadResourceWithResponseTimesMock.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders the heartbeat integration card by default for heartbeat monitors', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="heartbeat-integration-card"]').exists()).toBe(true)
    wrapper.unmount()
  })

  it('shows the bash integration snippet containing the ping URL', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    const snippet = wrapper.find('[data-testid="heartbeat-snippet"]')
    expect(snippet.exists()).toBe(true)
    expect(snippet.text()).toContain('curl')
    expect(snippet.text()).toContain('550e8400-e29b-41d4-a716-446655440000')
    wrapper.unmount()
  })

  it('does not render heartbeat integration card for non-heartbeat monitors', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue({
      id: 'http-1',
      name: 'My Site',
      type: 'http',
      target: 'https://example.com',
      interval: 60,
      timeout: 10,
      status: 'up',
      is_active: true,
      failure_count: 0,
      confirmation_checks: 2,
      confirmation_interval: 30,
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z',
      tags: [],
      incidents: [],
      response_times: [],
    })
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-testid="heartbeat-integration-card"]').exists()).toBe(false)
    wrapper.unmount()
  })

  it('shows the ping URL in the integration card', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    const pingUrlEl = wrapper.find('[data-testid="ping-url"]')
    expect(pingUrlEl.exists()).toBe(true)
    expect(pingUrlEl.text()).toContain('550e8400-e29b-41d4-a716-446655440000')
    wrapper.unmount()
  })
})

describe('ResourceView — next expected ping countdown (T046)', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    loadResourceWithResponseTimesMock.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows countdown when last_ping_at is set', async () => {
    const lastPingAt = new Date(Date.now() - 60_000).toISOString() // 1 min ago
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildHeartbeat({
        status: 'up',
        waiting: false,
        last_ping_at: lastPingAt,
        heartbeat_interval: 300,
        heartbeat_grace: 60,
      }),
    )
    const wrapper = mountView()
    await flushPromises()

    const countdown = wrapper.find('[data-testid="next-ping-countdown"]')
    expect(countdown.exists()).toBe(true)
    // deadline = last_ping_at + 300 + 60 = now - 60s + 360s = now + 300s → ~5m remaining
    expect(countdown.text()).toMatch(/\d+m\s+\d+s|Overdue/)
    wrapper.unmount()
  })

  it('shows Overdue when deadline has passed', async () => {
    const lastPingAt = new Date(Date.now() - 400_000).toISOString() // 400s ago, deadline was 360s
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildHeartbeat({
        status: 'down',
        waiting: false,
        last_ping_at: lastPingAt,
        heartbeat_interval: 300,
        heartbeat_grace: 60,
      }),
    )
    const wrapper = mountView()
    await flushPromises()

    const countdown = wrapper.find('[data-testid="next-ping-countdown"]')
    expect(countdown.exists()).toBe(true)
    expect(countdown.text()).toBe('Overdue')
    wrapper.unmount()
  })

  it('does not render countdown when monitor is waiting (no last_ping_at)', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildHeartbeat())
    const wrapper = mountView()
    await flushPromises()

    // countdown element is v-if="resource.last_ping_at"
    expect(wrapper.find('[data-testid="next-ping-countdown"]').exists()).toBe(false)
    wrapper.unmount()
  })
})

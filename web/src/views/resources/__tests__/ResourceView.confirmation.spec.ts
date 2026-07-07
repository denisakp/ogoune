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
  useRouter: () => ({
    back: vi.fn(),
  }),
  useRoute: () => ({
    params: { id: 'resource-1' },
  }),
}))

vi.mock('@nuxt/ui/composables/useToast', () => ({
  useToast: () => ({
    add: (input: { title?: string; color?: string }) => {
      if (input?.color === 'error') messageErrorMock(input.title)
    },
  }),
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

const buildResource = (overrides: Partial<Resource> = {}): Resource => ({
  id: 'resource-1',
  name: 'Primary API',
  type: 'http',
  target: 'https://example.com/health',
  interval: 60,
  timeout: 10,
  status: 'down',
  is_active: true,
  failure_count: 1,
  confirmation_checks: 3,
  confirmation_interval: 30,
  last_checked: new Date().toISOString(),
  created_at: '2025-01-01T00:00:00.000Z',
  updated_at: '2025-01-01T00:00:00.000Z',
  tags: [],
  incidents: [],
  response_times: [],
  ...overrides,
})

describe('ResourceView confirmation state rendering', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    loadResourceWithResponseTimesMock.mockReset()
    pauseResourceMock.mockReset()
    messageErrorMock.mockReset()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows confirming outage progress and countdown when below threshold', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(buildResource())

    const wrapper = mount(ResourceView, {
      global: {
        stubs: {
          ResourceModal: true,
          ResponseTimeChart: true,
        },
      },
    })

    await flushPromises()

    expect(loadResourceWithResponseTimesMock).toHaveBeenCalledWith('resource-1', 50)
    expect(wrapper.text()).toContain('Confirming outage: 1/3')
    expect(wrapper.text()).toMatch(/Next confirmation check in\s+\d+s/)

    wrapper.unmount()
  })

  it('does not show confirming banner when threshold is already reached', async () => {
    loadResourceWithResponseTimesMock.mockResolvedValue(
      buildResource({ failure_count: 3, confirmation_checks: 3 }),
    )

    const wrapper = mount(ResourceView, {
      global: {
        stubs: {
          ResourceModal: true,
          ResponseTimeChart: true,
        },
      },
    })

    await flushPromises()

    expect(wrapper.text()).not.toContain('Confirming outage:')

    wrapper.unmount()
  })
})

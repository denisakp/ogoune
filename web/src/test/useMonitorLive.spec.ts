import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { useMonitorLive } from '@/composables/useMonitorLive'
import * as liveService from '@/services/liveService'

describe('useMonitorLive', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.spyOn(liveService, 'fetchLiveSnapshot').mockResolvedValue({
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      resource: {} as any,
      stats: {
        uptime_2h: null,
        uptime_24h: null,
        uptime_7d: null,
        uptime_30d: null,
        avg_response_time_24h: null,
        last_response_time: null,
      },
      active_incident: null,
      recent_activities: [],
      fetched_at: new Date().toISOString(),
    })
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  const mountComposable = (interval = 10) => {
    let exposed: ReturnType<typeof useMonitorLive> | null = null
    const TestComponent = defineComponent({
      setup() {
        exposed = useMonitorLive('res-1', interval)
        return () => null
      },
    })

    const wrapper = mount(TestComponent)
    return {
      wrapper,
      get composable() {
        if (!exposed) {
          throw new Error('Composable not initialized')
        }
        return exposed
      },
    }
  }

  it('uses max(interval*1000, 15000) floor', () => {
    const { wrapper, composable } = mountComposable(5)
    expect(composable.pollingIntervalMs.value).toBe(15_000)
    wrapper.unmount()

    const mounted2 = mountComposable(20)
    expect(mounted2.composable.pollingIntervalMs.value).toBe(20_000)
    mounted2.wrapper.unmount()
  })

  it('SC-001 / FR-003: returns 5000ms when isWaiting=true, regardless of heartbeat_interval', () => {
    // Without the fix: max(300*1000, 15_000) = 300_000ms — violates the 5s SC-001 requirement.
    // With the fix: waiting=true overrides to 5_000ms.
    let exposed: ReturnType<typeof useMonitorLive> | null = null
    const TestComponent = defineComponent({
      setup() {
        exposed = useMonitorLive(
          'hb-waiting',
          () => 300,
          () => true,
        )
        return () => null
      },
    })
    const wrapper = mount(TestComponent)
    expect(exposed!.pollingIntervalMs.value).toBe(5_000)
    wrapper.unmount()
  })

  it('SC-001: retains normal interval when isWaiting=false', () => {
    let exposed: ReturnType<typeof useMonitorLive> | null = null
    const TestComponent = defineComponent({
      setup() {
        exposed = useMonitorLive(
          'hb-up',
          () => 60,
          () => false,
        )
        return () => null
      },
    })
    const wrapper = mount(TestComponent)
    expect(exposed!.pollingIntervalMs.value).toBe(60_000)
    wrapper.unmount()
  })

  it('SC-001: retains 15s floor when isWaiting=false and interval is very short', () => {
    let exposed: ReturnType<typeof useMonitorLive> | null = null
    const TestComponent = defineComponent({
      setup() {
        exposed = useMonitorLive(
          'hb-short',
          () => 5,
          () => false,
        )
        return () => null
      },
    })
    const wrapper = mount(TestComponent)
    expect(exposed!.pollingIntervalMs.value).toBe(15_000)
    wrapper.unmount()
  })

  it('SC-001: retains normal interval when isWaiting is omitted (backwards compat)', () => {
    let exposed: ReturnType<typeof useMonitorLive> | null = null
    const TestComponent = defineComponent({
      setup() {
        exposed = useMonitorLive('hb-compat', () => 60)
        return () => null
      },
    })
    const wrapper = mount(TestComponent)
    expect(exposed!.pollingIntervalMs.value).toBe(60_000)
    wrapper.unmount()
  })

  it('cleans up polling timer on unmount', async () => {
    const { wrapper } = mountComposable(15)
    await vi.advanceTimersByTimeAsync(30_000)
    const callsBeforeUnmount = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length

    wrapper.unmount()
    await vi.advanceTimersByTimeAsync(60_000)

    const callsAfterUnmount = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length
    expect(callsAfterUnmount).toBe(callsBeforeUnmount)
  })

  it('keeps liveData when transient poll fails', async () => {
    const { composable, wrapper } = mountComposable(15)

    await Promise.resolve()
    expect(composable.liveData.value).not.toBeNull()

    vi.mocked(liveService.fetchLiveSnapshot).mockRejectedValueOnce(new Error('network'))
    await composable.refresh()

    expect(composable.liveData.value).not.toBeNull()
    expect(composable.isTerminated.value).toBe(false)
    expect(composable.error.value).toContain('Could not refresh')

    wrapper.unmount()
  })

  it('sets isTerminated and stops polling on 404', async () => {
    const { composable, wrapper } = mountComposable(15)

    // Wait for the initial onMounted fetch to complete.
    await Promise.resolve()
    await Promise.resolve()

    // useMonitorLive checks `instanceof NotFoundError` (typed ApiError from
    // the new HTTP client). Mock the rejection with a real NotFoundError.
    const { NotFoundError } = await import('@/core/errors')
    vi.mocked(liveService.fetchLiveSnapshot).mockRejectedValueOnce(new NotFoundError())

    await composable.refresh()
    expect(composable.isTerminated.value).toBe(true)

    const callsBefore = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length
    await vi.advanceTimersByTimeAsync(45_000)
    const callsAfter = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length
    expect(callsAfter).toBe(callsBefore)

    wrapper.unmount()
  })

  it('pauses on hidden and resumes with immediate fetch when visible', async () => {
    const { wrapper } = mountComposable(15)

    await Promise.resolve()
    const beforeHidden = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length

    Object.defineProperty(document, 'hidden', { value: true, configurable: true })
    document.dispatchEvent(new Event('visibilitychange'))
    await vi.advanceTimersByTimeAsync(30_000)

    const whileHidden = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length
    expect(whileHidden).toBe(beforeHidden)

    Object.defineProperty(document, 'hidden', { value: false, configurable: true })
    document.dispatchEvent(new Event('visibilitychange'))
    await Promise.resolve()

    const afterVisible = vi.mocked(liveService.fetchLiveSnapshot).mock.calls.length
    expect(afterVisible).toBeGreaterThan(whileHidden)

    wrapper.unmount()
  })
})

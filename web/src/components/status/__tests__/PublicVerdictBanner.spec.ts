import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import PublicVerdictBanner from '../PublicVerdictBanner.vue'

function mkVerdict(status: 'operational' | 'partial_degradation' | 'major_outage') {
  return {
    status,
    label: status === 'operational' ? 'All Systems Operational' : 'Trouble',
    color: status === 'operational' ? 'green' : 'red',
  } as const
}

describe('PublicVerdictBanner', () => {
  it('sets data-status=operational and emerald palette', () => {
    const w = mount(PublicVerdictBanner, {
      props: {
        verdict: mkVerdict('operational'),
        generatedAt: new Date('2026-06-04T12:00:00Z'),
        secondsAgo: 0,
      },
    })
    expect(w.find('[data-status="operational"]').exists()).toBe(true)
    expect(w.html()).toContain('emerald')
  })

  it('sets data-status=partial_degradation and orange palette', () => {
    const w = mount(PublicVerdictBanner, {
      props: {
        verdict: mkVerdict('partial_degradation'),
        generatedAt: new Date('2026-06-04T12:00:00Z'),
        secondsAgo: 12,
      },
    })
    expect(w.find('[data-status="partial_degradation"]').exists()).toBe(true)
    expect(w.html()).toContain('orange')
  })

  it('sets data-status=major_outage and red palette', () => {
    const w = mount(PublicVerdictBanner, {
      props: {
        verdict: mkVerdict('major_outage'),
        generatedAt: new Date('2026-06-04T12:00:00Z'),
        secondsAgo: 30,
      },
    })
    expect(w.find('[data-status="major_outage"]').exists()).toBe(true)
    expect(w.html()).toContain('red')
  })

  it('formats "Last updated: …" timestamp from generatedAt', () => {
    const w = mount(PublicVerdictBanner, {
      props: {
        verdict: mkVerdict('operational'),
        generatedAt: new Date('2026-06-04T12:38:00Z'),
        secondsAgo: 1,
      },
    })
    expect(w.get('[data-testid="updated-label"]').text()).toContain('Last updated')
    expect(w.get('[data-testid="updated-label"]').text()).toContain('UTC')
  })

  it('falls back to "just now" when generatedAt is null', () => {
    const w = mount(PublicVerdictBanner, {
      props: {
        verdict: mkVerdict('operational'),
        generatedAt: null,
        secondsAgo: null,
      },
    })
    expect(w.get('[data-testid="updated-label"]').text()).toBe('Last updated: just now')
  })
})

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
  it('paints green for operational', () => {
    const w = mount(PublicVerdictBanner, {
      props: { verdict: mkVerdict('operational'), secondsAgo: 0 },
    })
    expect(w.attributes('data-status')).toBe('operational')
    expect(w.classes().some((c) => c.includes('emerald'))).toBe(true)
  })

  it('paints orange for partial degradation', () => {
    const w = mount(PublicVerdictBanner, {
      props: { verdict: mkVerdict('partial_degradation'), secondsAgo: 12 },
    })
    expect(w.attributes('data-status')).toBe('partial_degradation')
    expect(w.classes().some((c) => c.includes('orange'))).toBe(true)
  })

  it('paints red for major outage', () => {
    const w = mount(PublicVerdictBanner, {
      props: { verdict: mkVerdict('major_outage'), secondsAgo: 30 },
    })
    expect(w.attributes('data-status')).toBe('major_outage')
    expect(w.classes().some((c) => c.includes('red'))).toBe(true)
  })

  it('renders "Updated Xs ago" timestamp', () => {
    const w = mount(PublicVerdictBanner, {
      props: { verdict: mkVerdict('operational'), secondsAgo: 23 },
    })
    expect(w.get('[data-testid="updated-label"]').text()).toBe('Updated 23s ago')
  })

  it('renders "Updated just now" when secondsAgo is null or tiny', () => {
    const w = mount(PublicVerdictBanner, {
      props: { verdict: mkVerdict('operational'), secondsAgo: null },
    })
    expect(w.get('[data-testid="updated-label"]').text()).toBe('Updated just now')
  })
})

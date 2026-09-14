import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import CapabilityNotice from './CapabilityNotice.vue'
import type { HostCapabilities } from '@/types'

// Real NuxtUI components render under jsdom; only the icon is stubbed.
const stubs = { UIcon: { template: '<span />' } }

const on = { available: true }
const unreadable = { available: false, reason: 'unreadable' as const }
const settingOff = { available: false, reason: 'setting_off' as const }
const platform = { available: false, reason: 'platform' as const }

const native: HostCapabilities = {
  state: 'declared',
  kmsg: on,
  cgroup_oom: on,
  segfault: on,
  oom_detail: 'with_process',
  declared_at: '2026-09-12T09:00:00Z',
}
const container: HostCapabilities = {
  state: 'declared',
  kmsg: unreadable,
  cgroup_oom: on,
  segfault: unreadable,
  oom_detail: 'without_process',
}
const nativeSettingOff: HostCapabilities = {
  state: 'declared',
  kmsg: on,
  cgroup_oom: on,
  segfault: settingOff,
  oom_detail: 'with_process',
}
const nothing: HostCapabilities = {
  state: 'declared',
  kmsg: unreadable,
  cgroup_oom: unreadable,
  segfault: unreadable,
  oom_detail: 'none',
}
const mac: HostCapabilities = {
  state: 'declared',
  kmsg: platform,
  cgroup_oom: platform,
  segfault: platform,
  oom_detail: 'none',
}

function host(capabilities: HostCapabilities) {
  return mount(CapabilityNotice, { props: { capabilities, variant: 'host' }, global: { stubs } })
}
function incident(capabilities: HostCapabilities) {
  return mount(CapabilityNotice, {
    props: { capabilities, variant: 'incident' },
    global: { stubs },
  })
}
function badges(w: ReturnType<typeof host>) {
  return w.findAll('[data-test="capability-row-badge"]').map((b) => b.text())
}
function rowText(w: ReturnType<typeof host>, key: string) {
  return w.find(`[data-test="capability-row-${key}"] [data-test="capability-row-text"]`).text()
}

describe('CapabilityNotice — host variant', () => {
  it('declared, native, setting on: three available rows', () => {
    const w = host(native)
    expect(rowText(w, 'kmsg')).toBe('Readable')
    expect(rowText(w, 'oom')).toBe('Detected, with the process name')
    expect(rowText(w, 'segfault')).toBe('Available')
    expect(badges(w)).toEqual(['available', 'available', 'available'])
    expect(w.find('[data-test="capability-declared-at"]').exists()).toBe(true)
    expect(w.find('[data-test="capability-row-action"]').exists()).toBe(false)
  })

  it('declared, container: OOM kills without the process name; segfault needs the log, not the setting', () => {
    const w = host(container)
    expect(rowText(w, 'kmsg')).toContain('Not readable in this install')
    expect(rowText(w, 'oom')).toBe('Detected, without the process name')
    expect(rowText(w, 'segfault')).toBe('Unavailable — needs the kernel log')
    expect(w.find('[data-test="capability-row-action"]').exists()).toBe(false)
    expect(badges(w)).toEqual([
      'unavailable',
      'available',
      'unavailable',
    ])
  })

  it('declared, native, setting off: the sysctl line, and "best-effort"', () => {
    const w = host(nativeSettingOff)
    expect(rowText(w, 'segfault')).toContain('not reporting userspace faults')
    expect(rowText(w, 'segfault')).toContain('Best-effort')
    expect(w.find('[data-test="capability-row-action"]').text()).toBe(
      'sysctl -w debug.exception-trace=1',
    )
  })

  it('declared, nothing readable: not detected, not a crash', () => {
    const w = host(nothing)
    expect(rowText(w, 'oom')).toBe('Not detected')
    expect(rowText(w, 'segfault')).toBe('Unavailable — needs the kernel log')
  })

  it('declared, platform: says so, offers nothing to enable', () => {
    const w = host(mac)
    expect(rowText(w, 'kmsg')).toBe('Not supported on this platform')
    expect(rowText(w, 'segfault')).toBe('Not supported on this platform')
    expect(w.find('[data-test="capability-row-action"]').exists()).toBe(false)
  })

  it('not_reported: the upgrade hint, and NO rows', () => {
    const w = host({ state: 'not_reported' })
    const alert = w.find('[data-test="capability-notice-not_reported"]')
    expect(alert.exists()).toBe(true)
    expect(alert.text()).toContain('does not report')
    expect(alert.text()).toContain('Upgrade the agent')
    expect(w.findAll('[data-test^="capability-row-"]')).toHaveLength(0)
    expect(w.text()).not.toContain('unavailable')
  })

  it('not_known: "not yet known", and NO rows -- never rendered as unavailable', () => {
    const w = host({ state: 'not_known' })
    const alert = w.find('[data-test="capability-notice-not_known"]')
    expect(alert.exists()).toBe(true)
    expect(alert.text()).toContain('Not yet known')
    expect(w.findAll('[data-test^="capability-row-"]')).toHaveLength(0)
    expect(w.text()).not.toContain('unavailable')
  })
})

describe('CapabilityNotice — incident variant', () => {
  const sentence = (c: HostCapabilities) =>
    incident(c).find('[data-test="capability-notice-incident"]')

  it('container: only cgroup OOM kills could have been reported', () => {
    expect(sentence(container).text()).toContain('could not read the kernel log')
    expect(sentence(container).text()).toContain('only cgroup out-of-memory kills')
  })

  it('native, all on: could observe everything, none reported', () => {
    expect(sentence(native).text()).toContain('could observe out-of-memory kills and segfaults')
  })

  it('native, setting off: a segfault would not have been seen', () => {
    expect(sentence(nativeSettingOff).text()).toContain('segfault capture was off')
  })

  it('nothing readable: could not observe any', () => {
    expect(sentence(nothing).text()).toContain('could not observe any')
  })

  it('not_reported: says the agent version did not report', () => {
    expect(sentence({ state: 'not_reported' }).text()).toContain('did not report what it could observe')
  })

  it('not_known: says it is not known -- never borrows the present', () => {
    expect(sentence({ state: 'not_known' }).text()).toContain('is not known')
  })

  it('no_machine: renders nothing', () => {
    const w = incident({ state: 'no_machine' })
    expect(w.find('[data-test="capability-notice-incident"]').exists()).toBe(false)
    expect(w.find('[data-test="capability-notice-host"]').exists()).toBe(false)
  })
})

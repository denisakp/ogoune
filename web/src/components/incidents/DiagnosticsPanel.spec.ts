import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import DiagnosticsPanel from './DiagnosticsPanel.vue'

const stubs = { UIcon: { template: '<span />' } }

const diag = (
  overrides: Partial<{
    error_message: string
    response_body: string
    root_cause_hint: string | null
    keyword: string | null
    icmp_available: boolean | null
  }> = {},
) => ({
  id: 'd1',
  incident_id: 'i1',
  request_method: 'GET',
  request_url: 'https://x.test',
  request_timeout: 10,
  http_status_code: 500,
  response_size: 0,
  failure_type: 'http_5xx',
  error_message: 'Internal Server Error',
  error_summary: 'Server returned 500',
  total_duration: 1240,
  dns_duration: 42,
  tls_duration: 88,
  first_byte_duration: 920,
  body_truncated: false,
  body_encoded: false,
  ...overrides,
})

describe('DiagnosticsPanel', () => {
  it('renders structured key-value rows: Cause, Error, Request, HTTP Status, Timing', () => {
    const w = mount(DiagnosticsPanel, { global: { stubs }, props: { diagnostics: diag() } })
    expect(w.text()).toContain('Cause')
    expect(w.text()).toContain('http_5xx')
    expect(w.text()).toContain('Server returned 500')
    expect(w.text()).toContain('Request')
    expect(w.text()).toContain('GET https://x.test')
    expect(w.text()).toContain('HTTP Status')
    expect(w.text()).toContain('500')
    expect(w.text()).toContain('Timing breakdown')
    expect(w.text()).toContain('1240 ms')
  })

  it('renders Impact callout when error_summary or root_cause_hint is present', () => {
    const w = mount(DiagnosticsPanel, {
      global: { stubs },
      props: { diagnostics: diag({ root_cause_hint: 'Upstream database unreachable' }) },
    })
    expect(w.text()).toContain('Impact')
    expect(w.text()).toContain('Server returned 500')
  })

  it('renders ICMP section when icmp_available is set', () => {
    const w = mount(DiagnosticsPanel, {
      global: { stubs },
      props: { diagnostics: diag({ icmp_available: true }) },
    })
    expect(w.text()).toContain('ICMP probe')
    expect(w.text()).toContain('ICMP available')
  })

  it('renders Keyword section when keyword is set', () => {
    const w = mount(DiagnosticsPanel, {
      global: { stubs },
      props: { diagnostics: diag({ keyword: 'health' }) },
    })
    expect(w.text()).toContain('Keyword check')
    expect(w.text()).toContain('health')
  })

  it('shows empty state when no diagnostics', () => {
    const w = mount(DiagnosticsPanel, { global: { stubs }, props: { diagnostics: null } })
    expect(w.text()).toContain('No diagnostics available')
  })

  it('truncates response body > 5 KB and exposes Show full toggle', async () => {
    const longBody = 'x'.repeat(10_000)
    const w = mount(DiagnosticsPanel, {
      global: { stubs },
      props: { diagnostics: diag({ response_body: longBody }) },
    })
    expect(w.text()).toContain('Show full')
    expect(w.text()).toContain('KB total')
    const showBtn = w.findAll('button').find((b) => b.text() === 'Show full')
    await showBtn?.trigger('click')
    expect(w.text()).toContain('Show less')
  })
})

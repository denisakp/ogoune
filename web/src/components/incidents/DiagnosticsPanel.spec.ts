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

// --- Database health at failure (spec 088, US2) ---

describe('DiagnosticsPanel database health', () => {
  const withDb = (over: Record<string, unknown> = {}) =>
    mount(DiagnosticsPanel, {
      props: {
        diagnostics: {
          ...diag(),
          db_connections_active: 200,
          db_connections_max: 200,
          db_longest_query_seconds: 890.2,
          db_replication_lag_seconds: 4.5,
          ...over,
        },
      },
      global: { stubs },
    })

  // T036 -- the rows render when the incident carries them.
  it('shows what the database was doing when the check broke', () => {
    const block = withDb().get('[data-test="diagnostics-db-health"]')
    expect(block.text()).toContain('200 / 200')
    expect(block.text()).toContain('15 min')
    expect(block.text()).toContain('4.5 s')
  })

  // T036 / FR-019 -- absent means absent. A non-database incident renders no
  // block at all: not zeroed, not placeholdered, not an empty container.
  it('renders no block at all when the incident is not a database one', () => {
    const w = mount(DiagnosticsPanel, { props: { diagnostics: diag() }, global: { stubs } })
    expect(w.find('[data-test="diagnostics-db-health"]').exists()).toBe(false)
    expect(w.text()).not.toContain('Database at failure')
  })

  // Each field is independently nullable: a check that read only saturation shows
  // only saturation.
  it('shows only the fields the check actually read', () => {
    const block = withDb({
      db_longest_query_seconds: null,
      db_replication_lag_seconds: null,
    }).get('[data-test="diagnostics-db-health"]')
    expect(block.text()).toContain('200 / 200')
    expect(block.text()).not.toContain('Longest query')
    expect(block.text()).not.toContain('Replication lag')
  })

  // The saturation row needs both halves, exactly as on the monitor page.
  it('hides saturation when only half the pair survived', () => {
    const w = withDb({ db_connections_max: null })
    expect(w.find('[data-test="diagnostics-db-health"]').exists()).toBe(true)
    expect(w.get('[data-test="diagnostics-db-health"]').text()).not.toContain('Connections')
  })
})

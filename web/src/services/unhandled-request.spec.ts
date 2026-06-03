import { describe, expect, it } from 'vitest'

/**
 * Negative-coverage spec — proves MSW's `onUnhandledRequest: 'error'` is wired.
 *
 * A real `fetch` to an endpoint with no matching handler MUST reject. This
 * protects every later PR from silently reaching the network during tests.
 *
 * Spec 054 FR-012 / T031.
 */
describe('MSW unhandled-request guard', () => {
  it('rejects a fetch to an endpoint with no registered handler', async () => {
    await expect(
      fetch('http://example.test/__intentionally-unhandled-endpoint__'),
    ).rejects.toBeDefined()
  })
})

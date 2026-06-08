import type { App } from 'vue'
import type { Router } from 'vue-router'
import { createSyntheticIncident } from '@/views/errors/syntheticIncident'

let isHandlingError = false

const FALLBACK_HTML = `
<div style="min-height:100vh;display:flex;align-items:center;justify-content:center;font-family:system-ui,sans-serif;background:#fff;color:#0f172a;padding:24px;text-align:center">
  <div>
    <h1 style="font-size:24px;margin:0 0 8px">Something went very wrong.</h1>
    <p style="font-size:14px;color:#64748b;margin:0 0 16px">Please reload the page. If the issue persists, contact hello@ogoune.com.</p>
    <button onclick="location.reload()" style="padding:8px 16px;border-radius:6px;border:0;background:#4f46e5;color:#fff;font-weight:500;cursor:pointer">Reload</button>
  </div>
</div>
`

/**
 * Install a global Vue error handler that routes uncaught errors to the
 * branded 500 view, with re-entrancy protection.
 *
 * - First error: log + navigate to /error-500 with the synthetic incident.
 * - Second error during render of the 500 view itself: write a static HTML
 *   fallback into document.body to avoid an infinite render loop (FR-026).
 */
export function installErrorBoundary(app: App, router: Router): void {
  app.config.errorHandler = (err, _instance, info) => {
    const original = err instanceof Error ? err : new Error(String(err))

    // Always surface to console — this IS the existing client logger today.
    // (Replace with a structured logger if/when one is introduced.)
    console.error('[errorBoundary]', original, info)

    if (isHandlingError) {
      // Second-level failure: degrade to static HTML and stop.
      document.body.innerHTML = FALLBACK_HTML
      isHandlingError = false
      return
    }

    isHandlingError = true
    const incident = createSyntheticIncident(original.message)
    router
      .push({
        name: 'Error500',
        state: {
          incidentId: incident.id,
          occurredAt: incident.at.toISOString(),
          originalMessage: incident.originalMessage,
        },
      })
      .finally(() => {
        // Release the lock on next tick so future errors can be handled.
        setTimeout(() => {
          isHandlingError = false
        }, 0)
      })
  }
}

// Test-only helper.
export function __resetErrorBoundaryForTests(): void {
  isHandlingError = false
}

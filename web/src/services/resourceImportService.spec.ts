import { describe, expect, it } from 'vitest'
import { http, HttpResponse } from 'msw'
import { dryRunImport, importManifest, exportManifest } from './resourceImportService'
import { server } from '@/test/msw/server'

const manifest = 'version: 1\nresources:\n  - name: A\n    type: http\n    target: https://a.example.com\n'

describe('resourceImportService', () => {
  it('dryRunImport unwraps the { data } report envelope', async () => {
    server.use(
      http.post('*/v1/monitors/import', ({ request }) => {
        const url = new URL(request.url)
        expect(url.searchParams.get('dryRun')).toBe('true')
        return HttpResponse.json({
          data: {
            dry_run: true,
            total: 1,
            created: 0,
            skipped: 0,
            failed: 0,
            rows: [{ index: 0, name: 'A', valid: true, action: 'create' }],
          },
        })
      }),
    )
    const report = await dryRunImport(manifest)
    expect(report.dry_run).toBe(true)
    expect(report.rows).toHaveLength(1)
    expect(report.rows[0]?.action).toBe('create')
  })

  it('importManifest reads the 422 report body without throwing', async () => {
    server.use(
      http.post('*/v1/monitors/import', () =>
        HttpResponse.json(
          {
            data: {
              dry_run: false,
              total: 1,
              created: 0,
              skipped: 0,
              failed: 1,
              rows: [{ index: 0, name: 'A', valid: false, action: 'error', errors: ['target is required'] }],
            },
          },
          { status: 422 },
        ),
      ),
    )
    const report = await importManifest(manifest)
    expect(report.failed).toBe(1)
    expect(report.rows[0]?.errors).toContain('target is required')
  })

  it('exportManifest returns raw YAML text', async () => {
    server.use(
      http.get('*/v1/monitors/export', () =>
        HttpResponse.text('version: 1\nresources: []\n', {
          headers: { 'Content-Type': 'text/yaml' },
        }),
      ),
    )
    const yaml = await exportManifest()
    expect(yaml).toContain('version: 1')
  })
})

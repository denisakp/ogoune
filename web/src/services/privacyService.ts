import { HTTPError } from 'ky'
import { getAuthenticatedClient, request } from '@/core/http/client'
import type { PrivacySummary } from '@/types'

// Personal data (spec 094). The summary is a plain read; the export
// re-authenticates and comes back as a file.

/** Wrong password, missing or wrong two-factor code -- deliberately one error. */
export class InvalidCredentialsError extends Error {
  constructor() {
    super('The password or the two-factor code is not correct.')
    this.name = 'InvalidCredentialsError'
  }
}

export interface PersonalDataFile {
  blob: Blob
  filename: string
}

const SKIP_TOASTS = { 'x-skip-success-toast': '1', 'x-skip-error-toast': '1' }

function fallbackFilename(): string {
  return `ogoune-personal-data-${new Date().toISOString().slice(0, 10)}.json`
}

/** The server names the file; cross-origin setups may hide the header, so fall back to the same pattern. */
function filenameFrom(disposition: string | null): string {
  const m = disposition?.match(/filename="([^"]+)"/)
  return m?.[1] ?? fallbackFilename()
}

export async function getPrivacySummary(): Promise<PrivacySummary> {
  const r = await request<{ data: PrivacySummary }>(getAuthenticatedClient(), 'v1/me/privacy', {
    headers: { 'x-skip-success-toast': '1' },
  })
  return r.data
}

/**
 * Downloads the export. A failed re-authentication is a 422 -- never a 401,
 * which the client would treat as an expired session and sign the user out --
 * and becomes InvalidCredentialsError, with no toast: the form shows it.
 */
export async function exportPersonalData(input: {
  password: string
  code?: string
}): Promise<PersonalDataFile> {
  try {
    const res = await getAuthenticatedClient().post('v1/me/privacy/export', {
      json: { password: input.password, code: input.code ?? '' },
      headers: SKIP_TOASTS,
    })
    return { blob: await res.blob(), filename: filenameFrom(res.headers.get('content-disposition')) }
  } catch (error) {
    if (error instanceof HTTPError && error.response.status === 422) {
      throw new InvalidCredentialsError()
    }
    throw error
  }
}

const privacyService = { getPrivacySummary, exportPersonalData }
export default privacyService

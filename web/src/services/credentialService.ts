import { getAuthenticatedClient, request } from '@/core/http/client'
import { NotFoundError, type ApiError } from '@/core/errors'
import type {
  CredentialCreatePayload,
  CredentialResponse,
  TestConnectionResponse,
} from '@/types'

/**
 * Wrapper format used by all v1 endpoints: { data, meta }.
 */
interface V1Envelope<T> {
  data: T
}

/**
 * Fetch credential metadata for a resource. Password in the response is always
 * the mask string `••••••••` — the plaintext value never leaves the server.
 * Throws a `NotFoundError` when no credential is configured.
 */
export const fetchCredential = async (resourceId: string): Promise<CredentialResponse> => {
  const envelope = await request<V1Envelope<CredentialResponse>>(
    getAuthenticatedClient(),
    `v1/resources/${resourceId}/credentials`,
    { headers: { 'x-skip-error-toast': '1' } },
  )
  return envelope.data
}

/**
 * Create or atomically replace the credential for a resource.
 * Returns the masked response.
 */
export const setCredential = async (
  resourceId: string,
  payload: CredentialCreatePayload,
): Promise<CredentialResponse> => {
  const envelope = await request<V1Envelope<CredentialResponse>>(
    getAuthenticatedClient(),
    `v1/resources/${resourceId}/credentials`,
    { method: 'POST', json: payload },
  )
  return envelope.data
}

/**
 * Remove the credential for a resource. After this call the resource reverts
 * to the no-auth check behavior.
 */
export const deleteCredential = async (resourceId: string): Promise<void> => {
  await request<void>(
    getAuthenticatedClient(),
    `v1/resources/${resourceId}/credentials`,
    { method: 'DELETE' },
  )
}

/**
 * Live-test a credential payload without persisting it.
 * Rate-limited to 10 requests per minute per user (HTTP 429 on overflow).
 */
export const testCredential = async (
  resourceId: string,
  payload: CredentialCreatePayload,
): Promise<TestConnectionResponse> => {
  const envelope = await request<V1Envelope<TestConnectionResponse>>(
    getAuthenticatedClient(),
    `v1/resources/${resourceId}/credentials/test`,
    { method: 'POST', json: payload },
  )
  return envelope.data
}

/**
 * True when the thrown error is a "no credential configured" 404.
 */
export const isCredentialNotFound = (err: unknown): boolean => {
  return err instanceof NotFoundError
}

/**
 * Extract `Retry-After` seconds from a 429 typed error. Returns null otherwise.
 */
export const retryAfterSeconds = (err: unknown): number | null => {
  const apiErr = err as ApiError | undefined
  if (apiErr?.status !== 429) return null
  return apiErr.retryAfterSec ?? null
}

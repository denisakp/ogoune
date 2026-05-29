import type { AxiosError } from 'axios'

import axiosHelper from '../libs/axios.helper'
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
 * Throws an axios error with status 404 when no credential is configured.
 */
export const fetchCredential = async (resourceId: string): Promise<CredentialResponse> => {
  const { data } = await axiosHelper.get<V1Envelope<CredentialResponse>>(
    `/v1/resources/${resourceId}/credentials`,
    { skipErrorToast: true } as never,
  )
  return data.data
}

/**
 * Create or atomically replace the credential for a resource.
 * Returns the masked response.
 */
export const setCredential = async (
  resourceId: string,
  payload: CredentialCreatePayload,
): Promise<CredentialResponse> => {
  const { data } = await axiosHelper.post<V1Envelope<CredentialResponse>>(
    `/v1/resources/${resourceId}/credentials`,
    payload,
  )
  return data.data
}

/**
 * Remove the credential for a resource. After this call the resource reverts
 * to the no-auth check behavior.
 */
export const deleteCredential = async (resourceId: string): Promise<void> => {
  await axiosHelper.delete(`/v1/resources/${resourceId}/credentials`)
}

/**
 * Live-test a credential payload without persisting it.
 * Rate-limited to 10 requests per minute per user (HTTP 429 on overflow).
 */
export const testCredential = async (
  resourceId: string,
  payload: CredentialCreatePayload,
): Promise<TestConnectionResponse> => {
  const { data } = await axiosHelper.post<V1Envelope<TestConnectionResponse>>(
    `/v1/resources/${resourceId}/credentials/test`,
    payload,
  )
  return data.data
}

/**
 * Helper: true when an axios error is a "no credential configured" 404 from
 * the credentials GET endpoint. Lets callers treat the absence as a normal
 * empty state rather than a hard failure.
 */
export const isCredentialNotFound = (err: unknown): boolean => {
  const axiosErr = err as AxiosError | undefined
  return axiosErr?.response?.status === 404
}

/**
 * Helper: extract Retry-After header (seconds) from a 429 response.
 */
export const retryAfterSeconds = (err: unknown): number | null => {
  const axiosErr = err as AxiosError | undefined
  if (axiosErr?.response?.status !== 429) return null
  const raw = axiosErr.response.headers['retry-after']
  if (!raw) return null
  const n = Number.parseInt(String(raw), 10)
  return Number.isFinite(n) ? n : null
}

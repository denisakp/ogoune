import { getAuthenticatedClient, request } from '@/core/http/client'

const SKIP_SUCCESS = { headers: { 'x-skip-success-toast': '1' } }

export interface EscalationStep {
  id?: string
  delay_minutes: number
  channel_ids: string[]
}

export type EscalationScope = { kind: 'component'; value: string } | { kind: 'tag'; value: string }

export interface EscalationPolicy {
  id: string
  name: string
  scope: EscalationScope
  is_active: boolean
  priority: number
  steps: EscalationStep[]
}

export type EscalationPolicyInput = Omit<EscalationPolicy, 'id' | 'priority'>

interface Envelope<T> {
  data: T
}

const escalationService = {
  async list(): Promise<EscalationPolicy[]> {
    const r = await request<Envelope<EscalationPolicy[]>>(
      getAuthenticatedClient(),
      'escalation-policies',
      SKIP_SUCCESS,
    )
    return r.data
  },

  async create(payload: EscalationPolicyInput): Promise<EscalationPolicy> {
    const r = await request<Envelope<EscalationPolicy>>(
      getAuthenticatedClient(),
      'escalation-policies',
      { method: 'POST', json: payload },
    )
    return r.data
  },

  async update(id: string, payload: EscalationPolicyInput): Promise<EscalationPolicy> {
    const r = await request<Envelope<EscalationPolicy>>(
      getAuthenticatedClient(),
      `escalation-policies/${id}`,
      { method: 'PATCH', json: payload },
    )
    return r.data
  },

  async delete(id: string): Promise<void> {
    await request<void>(getAuthenticatedClient(), `escalation-policies/${id}`, {
      method: 'DELETE',
    })
  },

  async reorder(order: string[]): Promise<EscalationPolicy[]> {
    const r = await request<Envelope<EscalationPolicy[]>>(
      getAuthenticatedClient(),
      'escalation-policies/reorder',
      { method: 'PATCH', json: { order } },
    )
    return r.data
  },
}

export default escalationService

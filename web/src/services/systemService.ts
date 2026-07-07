import { http, request } from '@/core/http/client'

interface HasAccountsResponse {
  has_accounts: boolean
}

const SKIP_ERR = { headers: { 'x-skip-error-toast': '1' } }

const systemService = {
  async hasAccounts(): Promise<boolean> {
    const r = await request<HasAccountsResponse>(http, 'system/has-accounts', SKIP_ERR)
    return r.has_accounts
  },
}

export default systemService

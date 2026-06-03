import { setupServer } from 'msw/node'
import { baselineHandlers } from './handlers'

/**
 * Shared MSW server for the Vitest jsdom runtime. Lifecycle is wired in
 * `src/test/setup.ts`. Per-spec overrides go through `server.use(...)`.
 *
 * Contract: specs/054-slice-http-migration-axios/contracts/mock-server.md
 */
export const server = setupServer(...baselineHandlers)

/**
 * Client-side synthetic incident reference shown on the 500 page.
 * Cosmetic only — never persisted server-side. Gives the user something
 * concrete to quote when reporting the issue.
 */

export interface SyntheticIncidentRef {
  id: string
  at: Date
  originalMessage: string
}

const BASE32_ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'

function random6Base32(): string {
  let out = ''
  for (let i = 0; i < 6; i++) {
    out += BASE32_ALPHABET[Math.floor(Math.random() * BASE32_ALPHABET.length)]
  }
  return out
}

export function createSyntheticIncident(originalMessage: string): SyntheticIncidentRef {
  const at = new Date()
  const id = `INC-${at.getUTCFullYear()}-${random6Base32()}`
  return { id, at, originalMessage }
}

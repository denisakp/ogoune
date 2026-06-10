export type KeyboardShortcutSection = 'navigation' | 'actions' | 'view'

export type KeyboardShortcutKind = 'single' | 'combo' | 'chord'

export interface KeyboardShortcut {
  id: string
  section: KeyboardShortcutSection
  keys: string[]
  kind: KeyboardShortcutKind
  label: string
  handler: () => void
  whenInputFocused: boolean
}

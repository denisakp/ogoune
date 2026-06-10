import { ref } from 'vue'
import type { Router } from 'vue-router'
import type { KeyboardShortcut } from '@/types'
import { useSearchPalette } from '@/composables/useSearchPalette'

const CHORD_TIMEOUT_MS = 1200

const modalOpen = ref(false)

let shortcuts: readonly KeyboardShortcut[] = []
let listenerInstalled = false
let chordBuffer = ''
let chordTimer: ReturnType<typeof setTimeout> | null = null
let currentListener: ((e: KeyboardEvent) => void) | null = null

function isFormFieldFocused(): boolean {
  const el = document.activeElement as HTMLElement | null
  if (!el) return false
  if (['INPUT', 'TEXTAREA', 'SELECT'].includes(el.tagName)) return true
  if (el.isContentEditable === true) return true
  return false
}

function clearChord() {
  chordBuffer = ''
  if (chordTimer) {
    clearTimeout(chordTimer)
    chordTimer = null
  }
}

function armChord(key: string) {
  chordBuffer = key
  if (chordTimer) clearTimeout(chordTimer)
  chordTimer = setTimeout(clearChord, CHORD_TIMEOUT_MS)
}

function buildRegistry(router: Router): readonly KeyboardShortcut[] {
  const palette = useSearchPalette()
  const nav = (name: string) => () => {
    void router.push({ name })
  }
  return Object.freeze([
    {
      id: 'palette.open',
      section: 'navigation',
      keys: ['⌘', 'K'],
      kind: 'combo',
      label: 'Open search',
      handler: () => palette.toggle(),
      whenInputFocused: true,
    },
    {
      id: 'nav.overview',
      section: 'navigation',
      keys: ['G', 'O'],
      kind: 'chord',
      label: 'Go to Overview',
      handler: nav('Overview'),
      whenInputFocused: false,
    },
    {
      id: 'nav.resources',
      section: 'navigation',
      keys: ['G', 'R'],
      kind: 'chord',
      label: 'Go to Resources',
      handler: nav('Resources'),
      whenInputFocused: false,
    },
    {
      id: 'nav.incidents',
      section: 'navigation',
      keys: ['G', 'I'],
      kind: 'chord',
      label: 'Go to Incidents',
      handler: nav('Incidents'),
      whenInputFocused: false,
    },
    {
      id: 'nav.status',
      section: 'navigation',
      keys: ['G', 'S'],
      kind: 'chord',
      label: 'Go to Status pages',
      handler: nav('SettingsStatusPage'),
      whenInputFocused: false,
    },
    {
      id: 'shortcuts.open',
      section: 'view',
      keys: ['?'],
      kind: 'single',
      label: 'Show keyboard shortcuts',
      handler: () => {
        modalOpen.value = true
      },
      whenInputFocused: false,
    },
    {
      id: 'view.close',
      section: 'view',
      keys: ['Esc'],
      kind: 'single',
      label: 'Close modal / overlay',
      handler: () => {
        modalOpen.value = false
      },
      whenInputFocused: true,
    },
  ])
}

function findChordShortcut(seq: string): KeyboardShortcut | undefined {
  return shortcuts.find((s) => s.kind === 'chord' && s.keys.join('').toUpperCase() === seq.toUpperCase())
}

function makeListener(): (e: KeyboardEvent) => void {
  return (event: KeyboardEvent) => {
    // ⌘K / Ctrl+K — always available, even in inputs (FR-011).
    if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
      const s = shortcuts.find((x) => x.id === 'palette.open')
      if (s) {
        event.preventDefault()
        s.handler()
      }
      return
    }

    const inField = isFormFieldFocused()

    // Esc: close modal even when input focused (so it can dismiss overlays
    // mounted over an input).
    if (event.key === 'Escape' && modalOpen.value) {
      event.preventDefault()
      modalOpen.value = false
      clearChord()
      return
    }

    if (inField) {
      clearChord()
      return
    }

    // Single-key shortcuts (?).
    if (event.key === '?') {
      const s = shortcuts.find((x) => x.id === 'shortcuts.open')
      if (s) {
        event.preventDefault()
        s.handler()
        clearChord()
      }
      return
    }

    // Chord handling: leading G then nav key.
    const key = event.key.toUpperCase()
    if (chordBuffer === '' && key === 'G') {
      armChord('G')
      return
    }
    if (chordBuffer === 'G' && ['O', 'R', 'I', 'S'].includes(key)) {
      const match = findChordShortcut(`G${key}`)
      if (match) {
        event.preventDefault()
        match.handler()
      }
      clearChord()
      return
    }
    if (chordBuffer !== '') {
      // Any unexpected key after a leading chord resets the buffer.
      clearChord()
    }
  }
}

export function useKeyboardShortcuts() {
  return {
    modalOpen,
    shortcuts: () => shortcuts,
    open: () => {
      modalOpen.value = true
    },
    close: () => {
      modalOpen.value = false
    },
  }
}

export function installKeyboardShortcuts(router: Router): () => void {
  if (listenerInstalled) return () => undefined
  shortcuts = buildRegistry(router)
  currentListener = makeListener()
  document.addEventListener('keydown', currentListener)
  listenerInstalled = true
  return () => {
    if (currentListener) {
      document.removeEventListener('keydown', currentListener)
      currentListener = null
    }
    listenerInstalled = false
    shortcuts = []
    clearChord()
  }
}

// Test-only reset.
export function __resetKeyboardShortcutsForTests(): void {
  if (currentListener) {
    document.removeEventListener('keydown', currentListener)
    currentListener = null
  }
  listenerInstalled = false
  shortcuts = []
  modalOpen.value = false
  clearChord()
}

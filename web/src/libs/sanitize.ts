import DOMPurify from 'dompurify'

/**
 * The one sanitizer for rich text the app renders as HTML: incident updates,
 * the public status timeline, the editor preview. Three copies of this config
 * used to live in three components, each allowing a bare `<input>` -- the tag
 * the task-list checkbox needs, and also the tag a DOM-based attack wants.
 *
 * `<input>` is admitted only as a disabled checkbox. Anything else that
 * arrives as an input -- a text field, a file picker, a submit button -- is
 * dropped by the hook below before DOMPurify decides on attributes, so no
 * attribute allowance can widen it back.
 */
const ALLOWED_TAGS = [
  'p',
  'br',
  'strong',
  'em',
  'code',
  'a',
  'ul',
  'ol',
  'li',
  'h1',
  'h2',
  'input',
  'label',
  'div',
]

const ALLOWED_ATTR = [
  'href',
  'rel',
  'target',
  'type',
  'checked',
  'disabled',
  'data-checked',
  'data-type',
  'class',
]

const purifier = DOMPurify()

purifier.addHook('uponSanitizeElement', (node, data) => {
  if (data.tagName !== 'input') return
  const el = node as Element
  if (el.getAttribute('type')?.toLowerCase() !== 'checkbox') {
    el.parentNode?.removeChild(el)
    return
  }
  // A rendered task item is read-only; the editor owns the interactive one.
  el.setAttribute('disabled', '')
})

export function sanitizeRichText(html: string): string {
  return purifier.sanitize(html, { ALLOWED_TAGS, ALLOWED_ATTR })
}

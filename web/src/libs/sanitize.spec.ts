import { describe, expect, it } from 'vitest'
import { sanitizeRichText } from './sanitize'

describe('sanitizeRichText', () => {
  it('keeps the rich-text vocabulary', () => {
    const html = '<p><strong>a</strong> <em>b</em> <code>c</code> <a href="https://x" rel="noopener" target="_blank">l</a></p><ul><li>i</li></ul>'
    expect(sanitizeRichText(html)).toBe(html)
  })

  it('keeps a task-list checkbox, and makes it read-only', () => {
    const out = sanitizeRichText('<ul data-type="taskList"><li data-checked="true"><label><input type="checkbox" checked></label><div><p>t</p></div></li></ul>')
    expect(out).toContain('<input type="checkbox" checked="" disabled="">')
    expect(out).toContain('data-checked="true"')
  })

  it('drops every other kind of input', () => {
    for (const bad of [
      '<input type="text" value="x">',
      '<input type="file">',
      '<input type="submit" value="go">',
      '<input type="image" src="x" onerror="alert(1)">',
      '<input>',
      '<input type="CHECKBOX " onfocus="alert(1)" autofocus>',
    ]) {
      const out = sanitizeRichText(`<p>${bad}</p>`)
      expect(out, bad).not.toContain('<input')
    }
  })

  it('strips scripts and event handlers as before', () => {
    expect(sanitizeRichText('<p onclick="x()">a</p><script>1</script><img src=x onerror=1>')).toBe('<p>a</p>')
  })
})

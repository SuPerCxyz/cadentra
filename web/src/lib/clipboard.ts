export async function copyText(value: string): Promise<void> {
  if (!value) throw new Error('nothing to copy')

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value)
      return
    }
  } catch {
    // Fall through for HTTP origins and browsers that deny clipboard permission.
  }

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', '')
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  textarea.style.pointerEvents = 'none'
  document.body.appendChild(textarea)
  const copied = (() => {
    try {
      textarea.select()
      textarea.setSelectionRange(0, textarea.value.length)
      return document.execCommand('copy')
    } finally {
      textarea.remove()
    }
  })()
  if (!copied) throw new Error('clipboard copy failed')
}

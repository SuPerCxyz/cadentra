import { describe, expect, it } from 'vitest'
import { isSignInPath, safeRedirect } from './navigation'

describe('navigation redirects', () => {
  it('does not loop back to the sign-in page', () => {
    expect(isSignInPath('/sign-in')).toBe(true)
    expect(isSignInPath('/sign-in?redirect=%2Ftasks')).toBe(true)
    expect(safeRedirect('/sign-in?redirect=%2Ftasks')).toBe('/')
  })

  it('keeps same-origin paths and rejects external URLs', () => {
    expect(safeRedirect('/tasks?status=running#top')).toBe(
      '/tasks?status=running#top'
    )
    expect(safeRedirect('https://example.com/tasks')).toBe('/')
    expect(safeRedirect('//example.com/tasks')).toBe('/')
  })
})

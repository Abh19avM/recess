import { describe, it, expect, beforeEach } from 'vitest'
import { useToastStore, toast } from './toastStore'

describe('toastStore', () => {
  beforeEach(() => {
    useToastStore.setState({ toasts: [] })
  })

  it('adds a success toast notification', () => {
    toast.success('Match Won!', 'Great victory against classmate')
    const toasts = useToastStore.getState().toasts
    expect(toasts.length).toBe(1)
    expect(toasts[0].title).toBe('Match Won!')
    expect(toasts[0].type).toBe('success')
  })

  it('adds an error toast notification', () => {
    toast.error('Invalid Move', 'Square already occupied')
    const toasts = useToastStore.getState().toasts
    expect(toasts.length).toBe(1)
    expect(toasts[0].title).toBe('Invalid Move')
    expect(toasts[0].type).toBe('error')
  })

  it('removes a specific toast notification by id', () => {
    toast.info('Class Note', 'Recess bell in 5 minutes')
    const toastId = useToastStore.getState().toasts[0].id

    useToastStore.getState().removeToast(toastId)
    expect(useToastStore.getState().toasts.length).toBe(0)
  })
})

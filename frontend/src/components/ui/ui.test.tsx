import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Button } from './Button'
import { Input } from './Input'
import { Badge, Stamp } from './Badge'
import { PaperCard } from './PaperCard'
import { Avatar } from './Avatar'
import { EmptyState } from './EmptyState'
import { ErrorState } from './ErrorState'
import { LoadingState } from './LoadingState'

describe('Schoolhouse UI Components', () => {
  it('renders primary button correctly', () => {
    render(<Button variant="primary">Ring Bell</Button>)
    const button = screen.getByRole('button', { name: /ring bell/i })
    expect(button).toBeInTheDocument()
  })

  it('renders input with label and teacher error note', () => {
    render(<Input label="Student Handle" error="Handle is already taken" />)
    expect(screen.getByText('Student Handle')).toBeInTheDocument()
    expect(screen.getByText(/handle is already taken/i)).toBeInTheDocument()
  })

  it('renders rubber stamp with proper tone', () => {
    render(<Stamp tone="green">APPROVED A+</Stamp>)
    expect(screen.getByText('APPROVED A+')).toBeInTheDocument()
  })

  it('renders classroom badge properly', () => {
    render(<Badge variant="ink-blue">2 Players</Badge>)
    expect(screen.getByText('2 Players')).toBeInTheDocument()
  })

  it('renders ruled paper card with margin', () => {
    const { container } = render(<PaperCard variant="ruled">Notebook Content</PaperCard>)
    expect(container.firstChild).toHaveClass('paper-ruled')
  })

  it('renders student avatar with initials or preset', () => {
    render(<Avatar username="Abhinav" />)
    expect(screen.getByText('AB')).toBeInTheDocument()
  })

  it('renders EmptyState with action button', () => {
    const onAction = vi.fn()
    render(
      <EmptyState
        title="No Desks Found"
        description="There are currently no active games in this room."
        actionLabel="Create Desk"
        onAction={onAction}
      />
    )
    expect(screen.getByText('No Desks Found')).toBeInTheDocument()
    const btn = screen.getByRole('button', { name: /create desk/i })
    fireEvent.click(btn)
    expect(onAction).toHaveBeenCalledTimes(1)
  })

  it('renders ErrorState with teacher note and retry', () => {
    const onRetry = vi.fn()
    render(<ErrorState message="Connection lost with blackboard" onRetry={onRetry} />)
    expect(screen.getByText(/Connection lost with blackboard/i)).toBeInTheDocument()
    const retryBtn = screen.getByRole('button', { name: /try again/i })
    fireEvent.click(retryBtn)
    expect(onRetry).toHaveBeenCalledTimes(1)
  })

  it('renders LoadingState spinner', () => {
    render(<LoadingState message="Loading syllabus..." />)
    expect(screen.getByText('Loading syllabus...')).toBeInTheDocument()
  })
})

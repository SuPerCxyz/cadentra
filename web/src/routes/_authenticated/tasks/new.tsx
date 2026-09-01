import { createFileRoute } from '@tanstack/react-router'
import { TaskEditor } from '@/features/tasks'

export const Route = createFileRoute('/_authenticated/tasks/new')({
  component: TaskEditor,
})

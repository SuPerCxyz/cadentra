import { createFileRoute } from '@tanstack/react-router'
import { TaskEditor } from '@/features/tasks'

export const Route = createFileRoute('/_authenticated/tasks/$taskId/edit')({
  component: TaskEditor,
})

import { createFileRoute } from '@tanstack/react-router'
import { RunTask } from '@/features/tasks'

export const Route = createFileRoute('/_authenticated/tasks/$taskId/run')({
  component: RunTask,
})

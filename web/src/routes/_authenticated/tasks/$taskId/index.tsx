import { createFileRoute } from '@tanstack/react-router'
import { TaskDetail } from '@/features/tasks'

export const Route = createFileRoute('/_authenticated/tasks/$taskId/')({
  component: TaskDetail,
})

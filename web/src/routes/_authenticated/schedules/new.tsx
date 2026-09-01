import { createFileRoute } from '@tanstack/react-router'
import { ScheduleEditor } from '@/features/editors'

export const Route = createFileRoute('/_authenticated/schedules/new')({
  component: ScheduleEditor,
})

import { createFileRoute } from '@tanstack/react-router'
import { Schedules } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/schedules/')({
  component: Schedules,
})

import { createFileRoute } from '@tanstack/react-router'
import { Dashboard } from '@/features/overview'

export const Route = createFileRoute('/_authenticated/')({
  component: Dashboard,
})

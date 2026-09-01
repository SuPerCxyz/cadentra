import { createFileRoute } from '@tanstack/react-router'
import { Applications } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/applications/')({
  component: Applications,
})

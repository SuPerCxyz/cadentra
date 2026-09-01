import { createFileRoute } from '@tanstack/react-router'
import { Users } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/users/')({
  component: Users,
})

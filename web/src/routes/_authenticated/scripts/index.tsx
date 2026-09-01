import { createFileRoute } from '@tanstack/react-router'
import { Scripts } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/scripts/')({
  component: Scripts,
})

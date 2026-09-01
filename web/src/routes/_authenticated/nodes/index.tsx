import { createFileRoute } from '@tanstack/react-router'
import { Agents } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/nodes/')({
  component: Agents,
})

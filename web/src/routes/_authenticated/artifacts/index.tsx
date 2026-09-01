import { createFileRoute } from '@tanstack/react-router'
import { Artifacts } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/artifacts/')({
  component: Artifacts,
})

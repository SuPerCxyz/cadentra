import { createFileRoute } from '@tanstack/react-router'
import { Groups } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/groups/')({
  component: Groups,
})

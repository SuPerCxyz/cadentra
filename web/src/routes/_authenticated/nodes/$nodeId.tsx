import { createFileRoute } from '@tanstack/react-router'
import { AgentDetail } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/nodes/$nodeId')({
  component: AgentDetail,
})

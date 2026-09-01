import { createFileRoute } from '@tanstack/react-router'
import { AgentDetail } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/agents/$agentId')({
  component: AgentDetail,
})

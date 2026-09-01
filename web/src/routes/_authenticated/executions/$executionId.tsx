import { createFileRoute } from '@tanstack/react-router'
import { ExecutionDetail } from '@/features/executions'

export const Route = createFileRoute('/_authenticated/executions/$executionId')(
  { component: ExecutionDetail }
)

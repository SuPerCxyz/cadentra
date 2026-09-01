import { createFileRoute } from '@tanstack/react-router'
import { Audit } from '@/features/catalog'

export const Route = createFileRoute('/_authenticated/audit/')({
  component: Audit,
})

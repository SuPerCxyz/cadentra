import { createFileRoute } from '@tanstack/react-router'
import { FileTransfers } from '@/features/transfers'

export const Route = createFileRoute('/_authenticated/transfers/')({
  component: FileTransfers,
})

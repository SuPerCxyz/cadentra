import { createFileRoute } from '@tanstack/react-router'
import { ApplicationEditor } from '@/features/editors'

export const Route = createFileRoute('/_authenticated/applications/new')({
  component: ApplicationEditor,
})

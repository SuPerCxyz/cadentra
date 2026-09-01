import { createFileRoute } from '@tanstack/react-router'
import { ScriptEditor } from '@/features/editors'

export const Route = createFileRoute('/_authenticated/scripts/new')({
  component: ScriptEditor,
})

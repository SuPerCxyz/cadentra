import { createFileRoute } from '@tanstack/react-router'
import { ArtifactEditor } from '@/features/editors'

export const Route = createFileRoute('/_authenticated/artifacts/new')({
  component: ArtifactEditor,
})

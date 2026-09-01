import { createFileRoute } from '@tanstack/react-router'
import { GroupEditor } from '@/features/editors'

export const Route = createFileRoute('/_authenticated/groups/$groupId')({
  component: GroupEditor,
})

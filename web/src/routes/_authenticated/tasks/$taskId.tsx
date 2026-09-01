import { Outlet, createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/tasks/$taskId')({
  component: () => <Outlet />,
})

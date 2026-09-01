import { type QueryClient } from '@tanstack/react-query'
import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import { Toaster } from '@/components/ui/sonner'
import { NavigationProgress } from '@/components/navigation-progress'
import { GeneralError } from '@/features/system/general-error'
import { NotFoundError } from '@/features/system/not-found-error'

export const Route = createRootRouteWithContext<{ queryClient: QueryClient }>()(
  {
    component: () => (
      <>
        <NavigationProgress />
        <Outlet />
        <Toaster duration={5000} />
      </>
    ),
    notFoundComponent: NotFoundError,
    errorComponent: GeneralError,
  }
)

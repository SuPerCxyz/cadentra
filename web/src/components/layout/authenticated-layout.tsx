import { useEffect, useRef, useState } from 'react'
import { Outlet, useLocation, useNavigate } from '@tanstack/react-router'
import { sessionToUser, useAuthStore } from '@/stores/auth-store'
import { api, type SessionInfo } from '@/lib/api'
import { getCookie } from '@/lib/cookies'
import { currentPath } from '@/lib/navigation'
import { cn } from '@/lib/utils'
import { LayoutProvider } from '@/context/layout-provider'
import { SidebarInset, SidebarProvider } from '@/components/ui/sidebar'
import { Skeleton } from '@/components/ui/skeleton'
import { AppSidebar } from '@/components/layout/app-sidebar'
import { SkipToMain } from '@/components/skip-to-main'

type AuthenticatedLayoutProps = {
  children?: React.ReactNode
}

export function AuthenticatedLayout({ children }: AuthenticatedLayoutProps) {
  const defaultOpen = getCookie('sidebar_state') !== 'false'
  const navigate = useNavigate()
  const location = useLocation()
  const accessToken = useAuthStore((state) => state.auth.accessToken)
  const setUser = useAuthStore((state) => state.auth.setUser)
  const reset = useAuthStore((state) => state.auth.reset)
  const [checking, setChecking] = useState(true)
  const redirecting = useRef(false)

  useEffect(() => {
    let mounted = true
    const redirectToSignIn = () => {
      if (redirecting.current) return
      redirecting.current = true
      void navigate({
        to: '/sign-in',
        search: { redirect: currentPath() },
        replace: true,
      })
    }

    if (!accessToken) {
      redirectToSignIn()
      return () => {
        mounted = false
      }
    }

    api
      .get<SessionInfo>('/me')
      .then((session) => {
        if (mounted) setUser(sessionToUser(session))
      })
      .catch(() => {
        reset()
        redirectToSignIn()
      })
      .finally(() => {
        if (mounted) setChecking(false)
      })

    return () => {
      mounted = false
    }
  }, [
    accessToken,
    location.pathname,
    location.search,
    location.hash,
    navigate,
    reset,
    setUser,
  ])

  if (checking) {
    return (
      <div className='flex min-h-svh items-center justify-center p-6'>
        <Skeleton className='h-8 w-48' />
      </div>
    )
  }

  return (
    <LayoutProvider>
      <SidebarProvider defaultOpen={defaultOpen}>
        <SkipToMain />
        <AppSidebar />
        <SidebarInset
          className={cn(
            // Set content container, so we can use container queries
            '@container/content',

            // If layout is fixed, set the height
            // to 100svh to prevent overflow
            'has-data-[layout=fixed]:h-svh',

            // If layout is fixed and sidebar is inset,
            // set the height to 100svh - spacing (total margins) to prevent overflow
            'peer-data-[variant=inset]:has-data-[layout=fixed]:h-[calc(100svh-(var(--spacing)*4))]'
          )}
        >
          {children ?? <Outlet />}
        </SidebarInset>
      </SidebarProvider>
    </LayoutProvider>
  )
}

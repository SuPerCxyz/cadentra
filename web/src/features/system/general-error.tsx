import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'

export function GeneralError({ error }: { error?: Error }) {
  return (
    <div className='container flex min-h-svh flex-col items-center justify-center gap-4 text-center'>
      <h1 className='text-3xl font-semibold'>Unable to load this page</h1>
      <p className='max-w-md text-muted-foreground'>
        {error?.message || 'The request could not be completed.'}
      </p>
      <Button asChild>
        <Link to='/'>Back to overview</Link>
      </Button>
    </div>
  )
}

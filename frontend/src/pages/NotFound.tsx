import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { usePageTitle } from '@/hooks/use-page-title'

export default function NotFound() {
  usePageTitle('404')

  return (
    <div className="min-h-[60vh] flex flex-col items-center justify-center px-6 text-center">
      <p className="font-mono text-xs text-muted-foreground mb-4 tracking-wider">
        ERROR 404
      </p>
      <h1 className="text-3xl sm:text-4xl font-bold tracking-tight mb-3">
        These aren't the pages you're looking for.
      </h1>
      <p className="text-muted-foreground mb-8 max-w-md">
        *waves hand* Move along.
      </p>
      <Button size="lg" className="cursor-pointer" render={<Link to="/" />}>
        Back to home
      </Button>
    </div>
  )
}

import { Button } from '@/components/ui/button'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { useToast } from '@/hooks/use-toast'
import { AlertCircle, CheckCircle2 } from 'lucide-react'

export function ShadcnExample() {
  const { toast } = useToast()

  const handleToast = () => {
    toast({
      title: "Success!",
      description: "This is a shadcn/ui toast notification.",
      variant: "default",
    })
  }

  return (
    <div className="max-w-md mx-auto p-6 space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>shadcn/ui Components</CardTitle>
          <CardDescription>
            Examples of the installed shadcn/ui components working with your existing Tailwind setup.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" type="email" placeholder="Enter your email" />
          </div>
          
          <div className="flex gap-2">
            <Badge variant="default">Default</Badge>
            <Badge variant="secondary">Secondary</Badge>
            <Badge variant="outline">Outline</Badge>
            <Badge variant="destructive">Destructive</Badge>
          </div>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              This is an info alert using shadcn/ui components.
            </AlertDescription>
          </Alert>
        </CardContent>
        <CardFooter className="flex gap-2">
          <Button onClick={handleToast}>
            <CheckCircle2 className="h-4 w-4 mr-2" />
            Show Toast
          </Button>
          <Button variant="outline">Cancel</Button>
        </CardFooter>
      </Card>
    </div>
  )
}
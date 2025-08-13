import { useState } from 'react'
import { useAuthStore } from '../stores/auth'
import { User, Key, Bell, Shield } from 'lucide-react'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs'
import { Label } from '../components/ui/label'

export default function SettingsPage() {
  const { user, logout } = useAuthStore()

  return (
    <div className="page-content">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-gray-900">Settings</h1>
      </div>

      <Card>
        <Tabs defaultValue="profile" className="w-full">
          <CardHeader>
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="profile" className="flex items-center gap-2">
                <User className="h-4 w-4" />
                Profile
              </TabsTrigger>
              <TabsTrigger value="api-keys" className="flex items-center gap-2">
                <Key className="h-4 w-4" />
                API Keys
              </TabsTrigger>
              <TabsTrigger value="notifications" className="flex items-center gap-2">
                <Bell className="h-4 w-4" />
                Notifications
              </TabsTrigger>
              <TabsTrigger value="security" className="flex items-center gap-2">
                <Shield className="h-4 w-4" />
                Security
              </TabsTrigger>
            </TabsList>
          </CardHeader>

          <CardContent>
            <TabsContent value="profile" className="mt-0">
              <div className="space-y-6">
                <div>
                  <CardTitle className="text-lg mb-4">Profile Information</CardTitle>
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label className="text-sm font-medium">Email</Label>
                      <p className="text-sm text-gray-900">{user?.email}</p>
                    </div>
                    <div className="space-y-2">
                      <Label className="text-sm font-medium">User ID</Label>
                      <p className="text-sm text-gray-900 font-mono">{user?.id}</p>
                    </div>
                    <div className="space-y-2">
                      <Label className="text-sm font-medium">Member Since</Label>
                      <p className="text-sm text-gray-900">
                        {user?.created_at ? new Date(user.created_at).toLocaleDateString() : 'Unknown'}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="pt-6 border-t">
                  <Button
                    variant="destructive"
                    size="sm"
                    onClick={logout}
                  >
                    Sign out of your account
                  </Button>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="api-keys" className="mt-0">
              <div className="space-y-4">
                <CardTitle className="text-lg">API Keys</CardTitle>
                <CardDescription>
                  Manage your personal API keys. Project-specific keys can be found in each project's settings.
                </CardDescription>
                <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
                  <p className="text-sm text-yellow-800">
                    Personal API keys are coming soon. Use project-specific keys for now.
                  </p>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="notifications" className="mt-0">
              <div className="space-y-4">
                <CardTitle className="text-lg">Notification Preferences</CardTitle>
                <div className="space-y-4">
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="email-errors"
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      defaultChecked
                    />
                    <Label htmlFor="email-errors" className="text-sm">
                      Email notifications for errors
                    </Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="weekly-reports"
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      defaultChecked
                    />
                    <Label htmlFor="weekly-reports" className="text-sm">
                      Weekly summary reports
                    </Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="real-time-alerts"
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <Label htmlFor="real-time-alerts" className="text-sm">
                      Real-time alerts
                    </Label>
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="security" className="mt-0">
              <div className="space-y-4">
                <CardTitle className="text-lg">Security Settings</CardTitle>
                <div className="space-y-4">
                  <Button variant="outline" size="sm">
                    Change password
                  </Button>
                  <Button variant="outline" size="sm">
                    Enable two-factor authentication
                  </Button>
                  <Button variant="outline" size="sm">
                    View active sessions
                  </Button>
                </div>
              </div>
            </TabsContent>
          </CardContent>
        </Tabs>
      </Card>
    </div>
  )
}
import { ReactNode, useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { 
  BarChart3, 
  Settings, 
  FolderOpen,
  Menu,
  X,
  LogOut,
  User,
  ChevronLeft,
  ChevronRight
} from 'lucide-react'
import { useAuthStore } from '../stores/auth'

interface LayoutProps {
  children: ReactNode
}

const navigation = [
  { name: 'Dashboard', href: '/dashboard', icon: BarChart3 },
  { name: 'Projects', href: '/projects', icon: FolderOpen },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export default function Layout({ children }: LayoutProps) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const location = useLocation()
  const { user, logout } = useAuthStore()

  // Load collapsed state from localStorage on mount
  useEffect(() => {
    const savedState = localStorage.getItem('sidebar-collapsed')
    if (savedState) {
      setSidebarCollapsed(JSON.parse(savedState))
    }
  }, [])

  // Save collapsed state to localStorage whenever it changes
  const toggleSidebarCollapsed = () => {
    const newState = !sidebarCollapsed
    setSidebarCollapsed(newState)
    localStorage.setItem('sidebar-collapsed', JSON.stringify(newState))
  }

  const isActive = (href: string) => {
    return location.pathname === href || location.pathname.startsWith(href + '/')
  }

  return (
    <div className="h-screen flex overflow-hidden bg-gray-100">
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div 
          className="fixed inset-0 flex z-40 md:hidden"
          onClick={() => setSidebarOpen(false)}
        >
          <div className="fixed inset-0 bg-gray-600 bg-opacity-75" />
          <div className="relative flex-1 flex flex-col max-w-xs w-full bg-white">
            <div className="absolute top-0 right-0 -mr-12 pt-2">
              <button
                type="button"
                className="ml-1 flex items-center justify-center h-10 w-10 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white"
                onClick={() => setSidebarOpen(false)}
              >
                <X className="h-6 w-6 text-white" />
              </button>
            </div>
            <Sidebar collapsed={false} onToggleCollapse={toggleSidebarCollapsed} />
          </div>
        </div>
      )}

      {/* Desktop sidebar */}
      <div className="hidden md:flex md:flex-shrink-0">
        <div className={`flex flex-col transition-all duration-300 ${sidebarCollapsed ? 'w-16' : 'w-64'}`}>
          <Sidebar collapsed={sidebarCollapsed} onToggleCollapse={toggleSidebarCollapsed} />
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-col w-0 flex-1 overflow-hidden">
        {/* Mobile menu button - floating */}
        <button
          type="button"
          className="fixed top-4 left-4 z-20 p-2 rounded-md bg-white shadow-lg text-gray-500 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500 md:hidden"
          onClick={() => setSidebarOpen(true)}
        >
          <Menu className="h-6 w-6" />
        </button>

        {/* Page content - no header bar */}
        <main className="flex-1 relative overflow-y-auto focus:outline-none">
          {children}
        </main>
      </div>
    </div>
  )
}

interface SidebarProps {
  collapsed: boolean
  onToggleCollapse: () => void
}

function Sidebar({ collapsed, onToggleCollapse }: SidebarProps) {
  const location = useLocation()
  const { user, logout } = useAuthStore()

  const isActive = (href: string) => {
    return location.pathname === href || location.pathname.startsWith(href + '/')
  }

  return (
    <div className="flex flex-col h-full pt-5 bg-white overflow-y-auto">
      {/* Logo and toggle */}
      <div className={`flex items-center flex-shrink-0 mb-8 transition-all duration-300 ${collapsed ? 'px-2 justify-center' : 'px-4 justify-between'}`}>
        <div className={`flex items-center ${collapsed ? '' : 'space-x-2'}`}>
          <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center flex-shrink-0">
            <BarChart3 className="w-5 h-5 text-white" />
          </div>
          {!collapsed && (
            <div className="text-xl font-bold text-gray-900 whitespace-nowrap">EvalForge</div>
          )}
        </div>
        {/* Desktop sidebar toggle button */}
        <button
          type="button"
          className="hidden md:flex text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-primary-500 items-center transition-colors p-1 rounded"
          onClick={onToggleCollapse}
          title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <ChevronLeft className="h-4 w-4" />
          )}
        </button>
      </div>

      {/* Navigation */}
      <nav className={`flex-1 space-y-1 transition-all duration-300 ${collapsed ? 'px-2' : 'px-2'}`}>
        {navigation.map((item) => {
          const Icon = item.icon
          const active = isActive(item.href)
          
          return (
            <Link
              key={item.name}
              to={item.href}
              className={`group flex items-center px-2 py-2 text-sm font-medium rounded-md transition-all duration-200 ${
                active
                  ? 'bg-primary-100 text-primary-900 border-r-2 border-primary-500'
                  : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
              } ${collapsed ? 'justify-center' : ''}`}
              title={collapsed ? item.name : undefined}
            >
              <Icon className={`flex-shrink-0 h-5 w-5 ${collapsed ? '' : 'mr-3'} ${
                active ? 'text-primary-500' : 'text-gray-400 group-hover:text-gray-500'
              }`} />
              {!collapsed && (
                <span className="whitespace-nowrap">{item.name}</span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* User info and logout */}
      <div className={`flex-shrink-0 border-t border-gray-200 transition-all duration-300 ${collapsed ? 'px-2 py-3' : 'p-4'}`}>
        <div className={`flex flex-col ${collapsed ? 'items-center space-y-2' : 'space-y-3'}`}>
          <Link
            to="/settings"
            className={`flex items-center text-sm text-gray-600 hover:text-gray-900 transition-colors ${collapsed ? 'p-2' : 'px-3 py-2 rounded-md hover:bg-gray-50 w-full'}`}
            title="User Settings"
          >
            <User className="h-5 w-5" />
            {!collapsed && <span className="ml-3">Account</span>}
          </Link>
          <button
            onClick={logout}
            className={`flex items-center text-sm text-gray-600 hover:text-gray-900 transition-colors ${collapsed ? 'p-2' : 'px-3 py-2 rounded-md hover:bg-gray-50 w-full'}`}
            title="Logout"
          >
            <LogOut className="h-5 w-5" />
            {!collapsed && <span className="ml-3">Logout</span>}
          </button>
          {!collapsed && (
            <div className="text-xs text-gray-400 text-center pt-2">
              EvalForge v0.1.0
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
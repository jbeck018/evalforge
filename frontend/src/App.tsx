import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import { Toaster } from 'sonner'

import Layout from './components/Layout'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import ImprovedDashboard from './pages/ImprovedDashboard'
import ProjectsPage from './pages/ProjectsPage'
import ProjectDetailPage from './pages/ProjectDetailPage'
import TracesPage from './pages/TracesPage'
import AnalyticsPage from './pages/AnalyticsPage'
import SettingsPage from './pages/SettingsPage'
import EvaluationsPage from './pages/EvaluationsPage'
import EvaluationDetailPage from './pages/EvaluationDetailPage'
import ProjectSettingsPage from './pages/ProjectSettingsPage'
import LLMConfigPage from './pages/LLMConfigPage'
import AgentMonitoringPage from './pages/AgentMonitoringPage'

function App() {
  const { isAuthenticated } = useAuthStore()

  return (
    <>
      <Routes>
        {/* Public routes */}
        <Route 
          path="/login" 
          element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <LoginPage />} 
        />
        <Route 
          path="/register" 
          element={isAuthenticated ? <Navigate to="/dashboard" replace /> : <RegisterPage />} 
        />
        
        {/* Protected routes */}
        <Route 
          path="/*" 
          element={
            isAuthenticated ? (
              <Layout>
                <Routes>
                  <Route path="/" element={<Navigate to="/dashboard" replace />} />
                  <Route path="/dashboard" element={<ImprovedDashboard />} />
                  <Route path="/dashboard-old" element={<DashboardPage />} />
                  <Route path="/projects" element={<ProjectsPage />} />
                  <Route path="/projects/:id" element={<ProjectDetailPage />} />
                  <Route path="/projects/:id/evaluations" element={<EvaluationsPage />} />
                  <Route path="/projects/:projectId/evaluations/:evaluationId" element={<EvaluationDetailPage />} />
                  <Route path="/projects/:id/traces" element={<TracesPage />} />
                  <Route path="/projects/:id/analytics" element={<AnalyticsPage />} />
                  <Route path="/projects/:id/settings" element={<ProjectSettingsPage />} />
                  <Route path="/evaluations" element={<EvaluationsPage />} />
                  <Route path="/agents" element={<AgentMonitoringPage />} />
                  <Route path="/llm-config" element={<LLMConfigPage />} />
                  <Route path="/settings" element={<SettingsPage />} />
                  <Route path="*" element={<Navigate to="/dashboard" replace />} />
                </Routes>
              </Layout>
            ) : (
              <Navigate to="/login" replace />
            )
          } 
        />
      </Routes>
      <Toaster />
    </>
  )
}

export default App
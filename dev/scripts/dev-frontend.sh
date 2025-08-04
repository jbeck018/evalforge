#!/bin/bash
# Frontend development script with hot reloading

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🎨 Starting EvalForge Frontend Development Server${NC}"
echo -e "${BLUE}================================================${NC}"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo -e "${RED}❌ Node.js is required but not installed${NC}"
    echo -e "${YELLOW}Please install Node.js 18+ from https://nodejs.org/${NC}"
    exit 1
fi

# Check Node version
NODE_VERSION=$(node -v | cut -c 2-)
NODE_MAJOR_VERSION=$(echo $NODE_VERSION | cut -d. -f1)

if [ "$NODE_MAJOR_VERSION" -lt "18" ]; then
    echo -e "${RED}❌ Node.js version $NODE_VERSION is not supported${NC}"
    echo -e "${YELLOW}Please upgrade to Node.js 18+ from https://nodejs.org/${NC}"
    exit 1
fi

# Create frontend directory if it doesn't exist
mkdir -p frontend

# Create package.json if it doesn't exist
if [ ! -f frontend/package.json ]; then
    echo -e "${BLUE}📝 Creating React application...${NC}"
    cd frontend
    
    # Create package.json
    cat > package.json << 'EOF'
{
  "name": "evalforge-frontend",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "lint": "eslint src --ext .js,.jsx,.ts,.tsx",
    "lint:fix": "eslint src --ext .js,.jsx,.ts,.tsx --fix",
    "format": "prettier --write src/**/*.{js,jsx,ts,tsx,css,md}",
    "type-check": "tsc --noEmit"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.8.0",
    "@tanstack/react-query": "^4.24.4",
    "recharts": "^2.5.0",
    "date-fns": "^2.29.3",
    "clsx": "^1.2.1",
    "axios": "^1.3.0"
  },
  "devDependencies": {
    "@types/react": "^18.0.27",
    "@types/react-dom": "^18.0.10",
    "@vitejs/plugin-react": "^3.1.0",
    "vite": "^4.1.0",
    "typescript": "^4.9.4",
    "eslint": "^8.34.0",
    "@typescript-eslint/eslint-plugin": "^5.52.0",
    "@typescript-eslint/parser": "^5.52.0",
    "eslint-plugin-react": "^7.32.2",
    "eslint-plugin-react-hooks": "^4.6.0",
    "prettier": "^2.8.4",
    "autoprefixer": "^10.4.13",
    "postcss": "^8.4.21",
    "tailwindcss": "^3.2.6"
  }
}
EOF

    # Install dependencies
    echo -e "${BLUE}📦 Installing dependencies...${NC}"
    npm install

    cd ..
fi

# Create Vite configuration if it doesn't exist
if [ ! -f frontend/vite.config.ts ]; then
    echo -e "${BLUE}⚙️  Creating Vite configuration...${NC}"
    cat > frontend/vite.config.ts << 'EOF'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    host: true, // Allow external connections
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true,
        rewrite: (path) => path
      },
      '/ws': {
        target: 'ws://localhost:8000',
        ws: true,
        changeOrigin: true
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@pages': path.resolve(__dirname, './src/pages'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils'),
      '@api': path.resolve(__dirname, './src/api'),
      '@types': path.resolve(__dirname, './src/types')
    }
  },
  build: {
    sourcemap: true,
    outDir: 'dist',
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'query-vendor': ['@tanstack/react-query'],
          'chart-vendor': ['recharts'],
          'util-vendor': ['date-fns', 'clsx', 'axios']
        }
      }
    }
  },
  optimizeDeps: {
    include: ['react', 'react-dom', 'react-router-dom']
  }
})
EOF
fi

# Create TypeScript configuration if it doesn't exist
if [ ! -f frontend/tsconfig.json ]; then
    echo -e "${BLUE}📝 Creating TypeScript configuration...${NC}"
    cat > frontend/tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "ESNext",
    "lib": ["dom", "dom.iterable", "es6"],
    "allowJs": true,
    "skipLibCheck": true,
    "esModuleInterop": true,
    "allowSyntheticDefaultImports": true,
    "strict": true,
    "forceConsistentCasingInFileNames": true,
    "module": "ESNext",
    "moduleResolution": "node",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"],
      "@components/*": ["./src/components/*"],
      "@pages/*": ["./src/pages/*"],
      "@hooks/*": ["./src/hooks/*"],
      "@utils/*": ["./src/utils/*"],
      "@api/*": ["./src/api/*"],
      "@types/*": ["./src/types/*"]
    }
  },
  "include": [
    "src"
  ],
  "exclude": [
    "node_modules"
  ]
}
EOF
fi

# Create basic React app structure if it doesn't exist
if [ ! -d frontend/src ]; then
    echo -e "${BLUE}📁 Creating React application structure...${NC}"
    mkdir -p frontend/src/{components,pages,hooks,utils,api,types,styles}
    
    # Create index.html
    cat > frontend/index.html << 'EOF'
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/vite.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>EvalForge - LLM Observability Platform</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
EOF

    # Create main.tsx
    cat > frontend/src/main.tsx << 'EOF'
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './styles/index.css'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
EOF

    # Create App.tsx
    cat > frontend/src/App.tsx << 'EOF'
import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import Dashboard from './pages/Dashboard'
import Projects from './pages/Projects'
import Analytics from './pages/Analytics'
import Navigation from './components/Navigation'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="min-h-screen bg-gray-50">
          <Navigation />
          <main className="container mx-auto px-4 py-8">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/projects" element={<Projects />} />
              <Route path="/analytics" element={<Analytics />} />
            </Routes>
          </main>
        </div>
      </Router>
    </QueryClientProvider>
  )
}

export default App
EOF

    # Create basic components
    mkdir -p frontend/src/components
    cat > frontend/src/components/Navigation.tsx << 'EOF'
import React from 'react'
import { Link, useLocation } from 'react-router-dom'

const Navigation: React.FC = () => {
  const location = useLocation()

  const navItems = [
    { path: '/', label: 'Dashboard' },
    { path: '/projects', label: 'Projects' },
    { path: '/analytics', label: 'Analytics' },
  ]

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center space-x-8">
            <div className="text-xl font-bold text-blue-600">EvalForge</div>
            <div className="flex space-x-4">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    location.pathname === item.path
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>
        </div>
      </div>
    </nav>
  )
}

export default Navigation
EOF

    # Create basic pages
    mkdir -p frontend/src/pages
    cat > frontend/src/pages/Dashboard.tsx << 'EOF'
import React from 'react'
import { useQuery } from '@tanstack/react-query'

const Dashboard: React.FC = () => {
  const { data: healthData, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: async () => {
      const response = await fetch('/api/health')
      if (!response.ok) {
        throw new Error('Health check failed')
      }
      return response.json()
    },
  })

  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Dashboard</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-900">API Status</h3>
          <p className={`text-2xl font-bold ${isLoading ? 'text-yellow-600' : healthData ? 'text-green-600' : 'text-red-600'}`}>
            {isLoading ? 'Checking...' : healthData ? 'Healthy' : 'Offline'}
          </p>
        </div>
        
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-900">Total Events</h3>
          <p className="text-2xl font-bold text-blue-600">1,234</p>
        </div>
        
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-900">Active Projects</h3>
          <p className="text-2xl font-bold text-purple-600">5</p>
        </div>
        
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold text-gray-900">Cost Today</h3>
          <p className="text-2xl font-bold text-green-600">$12.45</p>
        </div>
      </div>

      <div className="bg-white p-6 rounded-lg shadow">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">Welcome to EvalForge</h2>
        <p className="text-gray-600 mb-4">
          Your LLM observability platform is running! This is a development environment 
          with hot reloading enabled.
        </p>
        <div className="bg-blue-50 p-4 rounded-md">
          <h3 className="text-sm font-medium text-blue-800">Quick Links:</h3>
          <ul className="mt-2 text-sm text-blue-700">
            <li>• API Documentation: <a href="http://localhost:8089" target="_blank" rel="noopener noreferrer" className="underline">localhost:8089</a></li>
            <li>• ClickHouse Console: <a href="http://localhost:8123" target="_blank" rel="noopener noreferrer" className="underline">localhost:8123</a></li>
            <li>• Grafana Dashboard: <a href="http://localhost:3001" target="_blank" rel="noopener noreferrer" className="underline">localhost:3001</a></li>
          </ul>
        </div>
      </div>
    </div>
  )
}

export default Dashboard
EOF

    cat > frontend/src/pages/Projects.tsx << 'EOF'
import React from 'react'

const Projects: React.FC = () => {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Projects</h1>
      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600">Project management interface coming soon...</p>
      </div>
    </div>
  )
}

export default Projects
EOF

    cat > frontend/src/pages/Analytics.tsx << 'EOF'
import React from 'react'

const Analytics: React.FC = () => {
  return (
    <div>
      <h1 className="text-3xl font-bold text-gray-900 mb-8">Analytics</h1>
      <div className="bg-white p-6 rounded-lg shadow">
        <p className="text-gray-600">Analytics dashboard coming soon...</p>
      </div>
    </div>
  )
}

export default Analytics
EOF

    # Create basic CSS
    cat > frontend/src/styles/index.css << 'EOF'
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  * {
    box-sizing: border-box;
  }
  
  body {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
      'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }
}

@layer components {
  .container {
    @apply max-w-7xl mx-auto;
  }
}
EOF

    # Create Tailwind config
    cat > frontend/tailwind.config.js << 'EOF'
/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#eff6ff',
          500: '#3b82f6',
          600: '#2563eb',
          700: '#1d4ed8',
        }
      }
    },
  },
  plugins: [],
}
EOF

    # Create PostCSS config
    cat > frontend/postcss.config.js << 'EOF'
export default {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
EOF
fi

# Install dependencies if node_modules doesn't exist
if [ ! -d frontend/node_modules ]; then
    echo -e "${BLUE}📦 Installing Node.js dependencies...${NC}"
    cd frontend
    npm install
    cd ..
fi

# Change to frontend directory and start development server
cd frontend

echo -e "${GREEN}🎨 Starting hot-reload development server...${NC}"
echo -e "${YELLOW}💡 The app will automatically reload when you change files${NC}"
echo -e "${YELLOW}💡 Press Ctrl+C to stop the server${NC}"
echo -e "${BLUE}🌐 Frontend will be available at: http://localhost:3000${NC}"
echo ""

# Start Vite development server
npm run dev
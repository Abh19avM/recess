import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AppLayout } from './components/layout/AppLayout'
import { LandingPage } from './pages/LandingPage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { DashboardPage } from './pages/DashboardPage'
import { GamesPage } from './pages/GamesPage'
import { ProfilePage } from './pages/ProfilePage'
import { WebSocketTestPage } from './pages/WebSocketTestPage'
import { XOArenaPage } from './pages/XOArenaPage'
import { HandCricketArenaPage } from './pages/HandCricketArenaPage'
import { DotsBoxesArenaPage } from './pages/DotsBoxesArenaPage'
import { Connect4ArenaPage } from './pages/Connect4ArenaPage'
import { PaperFootballArenaPage } from './pages/PaperFootballArenaPage'
import { NPATArenaPage } from './pages/NPATArenaPage'
import { SpectatorArenaPage } from './pages/SpectatorArenaPage'
import { ProtectedRoute } from './routes/ProtectedRoute'
import { EmptyState } from './components/ui/EmptyState'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
    },
  },
})

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Routes>
          <Route element={<AppLayout />}>
            <Route path="/" element={<LandingPage />} />
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/games" element={<GamesPage />} />
            <Route path="/games/xo" element={<XOArenaPage />} />
            <Route path="/games/xo/:roomId" element={<XOArenaPage />} />
            <Route path="/games/hand-cricket" element={<HandCricketArenaPage />} />
            <Route path="/games/hand-cricket/:roomId" element={<HandCricketArenaPage />} />
            <Route path="/games/dots-and-boxes" element={<DotsBoxesArenaPage />} />
            <Route path="/games/dots-and-boxes/:roomId" element={<DotsBoxesArenaPage />} />
            <Route path="/games/connect-4" element={<Connect4ArenaPage />} />
            <Route path="/games/connect-4/:roomId" element={<Connect4ArenaPage />} />
            <Route path="/games/paper-football" element={<PaperFootballArenaPage />} />
            <Route path="/games/paper-football/:roomId" element={<PaperFootballArenaPage />} />
            <Route path="/games/npat" element={<NPATArenaPage />} />
            <Route path="/games/npat/:roomId" element={<NPATArenaPage />} />
            <Route path="/spectate/:roomId" element={<SpectatorArenaPage />} />
            <Route path="/spectate/:gameType/:roomId" element={<SpectatorArenaPage />} />
            <Route path="/ws-test" element={<WebSocketTestPage />} />
            
            {/* Protected Routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <DashboardPage />
                </ProtectedRoute>
              }
            />
            <Route
              path="/profile"
              element={
                <ProtectedRoute>
                  <ProfilePage />
                </ProtectedRoute>
              }
            />

            {/* 404 Fallback */}
            <Route
              path="*"
              element={
                <EmptyState
                  title="Page Off-Syllabus (404)"
                  description="The page you are looking for has been moved or does not exist on the classroom notice board."
                  actionLabel="Return to Classroom"
                  onAction={() => window.location.replace('/')}
                />
              }
            />
          </Route>
        </Routes>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

export default App

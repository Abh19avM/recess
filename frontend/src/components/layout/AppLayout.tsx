import React from 'react'
import { Outlet } from 'react-router-dom'
import { Navbar } from './Navbar'
import { Footer } from './Footer'
import { ToastContainer } from '../ui/ToastContainer'

export const AppLayout: React.FC = () => {
  return (
    <div className="min-h-screen flex flex-col bg-[#FBF9F3] dark:bg-[#101713] text-[#1E242B] dark:text-[#F1F5F9] antialiased transition-colors duration-200">
      <Navbar />
      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <Outlet />
      </main>
      <Footer />
      <ToastContainer />
    </div>
  )
}

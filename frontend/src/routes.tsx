import { createBrowserRouter, Navigate } from 'react-router-dom'
import { AuthLayout } from '@/components/layout/AuthLayout'
import { MainLayout } from '@/components/layout/MainLayout'
import LoginPage from '@/pages/auth/LoginPage'
import SignupPage from '@/pages/auth/SignupPage'
import HomePage from '@/pages/home/HomePage'
import ProfilePage from '@/pages/profile/ProfilePage'
import EditProfilePage from '@/pages/profile/EditProfilePage'
import PostDetailPage from '@/pages/post/PostDetailPage'
import GroupsPage from '@/pages/groups/GroupsPage'
import GroupDetailPage from '@/pages/groups/GroupDetailPage'
import CreateGroupPage from '@/pages/groups/CreateGroupPage'
import ChatPage from '@/pages/chat/ChatPage'
import NotificationsPage from '@/pages/notifications/NotificationsPage'
import FollowersPage from '@/pages/followers/FollowersPage'
import FollowingPage from '@/pages/followers/FollowingPage'
import SearchPage from '@/pages/search/SearchPage'
import NotFoundPage from '@/pages/errors/NotFoundPage'
import ErrorPage from '@/pages/errors/ErrorPage'

export const router = createBrowserRouter([
  {
    element: <AuthLayout />,
    errorElement: <ErrorPage />,
    children: [
      { path: '/login', element: <LoginPage /> },
      { path: '/signup', element: <SignupPage /> },
    ],
  },
  {
    element: <MainLayout />,
    errorElement: <ErrorPage />,
    children: [
      { index: true, element: <Navigate to="/home" replace /> },
      { path: '/home', element: <HomePage /> },
      { path: '/profile/:uuid', element: <ProfilePage /> },
      { path: '/profile/me', element: <EditProfilePage /> },
      { path: '/posts/:uuid', element: <PostDetailPage /> },
      { path: '/groups', element: <GroupsPage /> },
      { path: '/groups/new', element: <CreateGroupPage /> },
      { path: '/groups/:uuid', element: <GroupDetailPage /> },
      { path: '/chat', element: <ChatPage /> },
      { path: '/chat/:id', element: <ChatPage /> },
      { path: '/notifications', element: <NotificationsPage /> },
      { path: '/followers', element: <FollowersPage /> },
      { path: '/following', element: <FollowingPage /> },
      { path: '/search', element: <SearchPage /> },
    ],
  },
  { path: '/404', element: <NotFoundPage /> },
  { path: '*', element: <NotFoundPage /> },
])

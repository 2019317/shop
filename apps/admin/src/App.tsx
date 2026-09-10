import { Navigate, Route, Routes } from 'react-router-dom'
import Login from './pages/Login'
import AdminLayout from './layouts/AdminLayout'
import Products from './pages/Products'
import ProductEdit from './pages/ProductEdit'
import Categories from './pages/Categories'
import Orders from './pages/Orders'
import Coupons from './pages/Coupons'
import Shipping from './pages/Shipping'
import Logs from './pages/Logs'
import { getToken } from './api/client'

function RequireAuth({ children }: { children: React.ReactNode }) {
  if (!getToken()) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <RequireAuth>
            <AdminLayout />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="/products" replace />} />
        <Route path="products" element={<Products />} />
        <Route path="products/new" element={<ProductEdit />} />
        <Route path="products/:id" element={<ProductEdit />} />
        <Route path="categories" element={<Categories />} />
        <Route path="orders" element={<Orders />} />
        <Route path="coupons" element={<Coupons />} />
        <Route path="shipping" element={<Shipping />} />
        <Route path="logs" element={<Logs />} />
      </Route>
    </Routes>
  )
}

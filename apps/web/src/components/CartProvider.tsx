'use client'

import { createContext, useContext, useEffect, useMemo, useState } from 'react'

export interface CartItem {
  variantId: string
  skuCode: string
  title: string
  options: Record<string, string>
  priceCents: number
  imageUrl: string
  qty: number
}

interface CartContextValue {
  items: CartItem[]
  add: (item: Omit<CartItem, 'qty'>, qty?: number) => void
  remove: (variantId: string) => void
  updateQty: (variantId: string, qty: number) => void
  clear: () => void
  count: number
  subtotalCents: number
}

const CartContext = createContext<CartContextValue | null>(null)
const STORAGE_KEY = 'shop_cart'

export default function CartProvider({ children }: { children: React.ReactNode }) {
  const [items, setItems] = useState<CartItem[]>([])

  // 购物车存本地：访客也能加购，降低海外用户流失
  useEffect(() => {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) {
      try {
        setItems(JSON.parse(raw))
      } catch {
        /* ignore invalid cache */
      }
    }
  }, [])

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(items))
  }, [items])

  const value = useMemo<CartContextValue>(() => {
    const add: CartContextValue['add'] = (item, qty = 1) => {
      setItems((prev) => {
        const found = prev.find((i) => i.variantId === item.variantId)
        if (found) {
          return prev.map((i) =>
            i.variantId === item.variantId ? { ...i, qty: i.qty + qty } : i,
          )
        }
        return [...prev, { ...item, qty }]
      })
    }
    const remove: CartContextValue['remove'] = (variantId) =>
      setItems((prev) => prev.filter((i) => i.variantId !== variantId))
    const updateQty: CartContextValue['updateQty'] = (variantId, qty) =>
      setItems((prev) =>
        prev.map((i) => (i.variantId === variantId ? { ...i, qty: Math.max(1, qty) } : i)),
      )

    return {
      items,
      add,
      remove,
      updateQty,
      clear: () => setItems([]),
      count: items.reduce((sum, i) => sum + i.qty, 0),
      subtotalCents: items.reduce((sum, i) => sum + i.priceCents * i.qty, 0),
    }
  }, [items])

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}

export function useCart() {
  const ctx = useContext(CartContext)
  if (!ctx) throw new Error('useCart must be used within CartProvider')
  return ctx
}

import type { Locale } from './config'

// 文案字典。新增语言时在此扩展，类型系统会强制校验完整性
const en = {
  nav: {
    shopAll: 'Shop All',
    journals: 'Journals',
    stickers: 'Stickers',
    cart: 'Cart',
  },
  home: {
    title: 'Paper goods made for slow living',
    subtitle: 'Handcrafted journals, planner inserts and stickers — shipped worldwide.',
    viewAll: 'View all products',
  },
  products: {
    shopAll: 'Shop All',
    searchPrefix: 'Search',
    count: 'products',
    empty: 'No products found.',
    sortNewest: 'Newest',
    sortPriceAsc: 'Price: Low to High',
    sortPriceDesc: 'Price: High to Low',
  },
  product: {
    qty: 'Qty',
    addToCart: 'Add to cart',
    outOfStock: 'Out of stock',
    description: 'Description',
  },
  cart: {
    title: 'Cart',
    empty: 'Your cart is empty.',
    startShopping: 'Start shopping',
    subtotal: 'Subtotal',
    checkout: 'Checkout',
    remove: 'Remove',
    shippingNote: 'Shipping is calculated at checkout.',
  },
  checkout: {
    title: 'Checkout',
    contact: 'Contact',
    shippingAddress: 'Shipping address',
    email: 'Email',
    recipient: 'Recipient',
    country: 'Country',
    state: 'State / Province',
    city: 'City',
    address: 'Address',
    apartment: 'Apartment, suite, etc.',
    postalCode: 'Postal code',
    phone: 'Phone',
    orderNote: 'Order note',
    placeOrder: 'Place order',
    placing: 'Placing order...',
    summary: 'Order summary',
    emptyCart: 'Your cart is empty',
    fillRequired: 'Please fill in required fields',
    serverShipping: 'Shipping calculated by destination at the server.',
  },
  order: {
    title: 'Order',
    items: 'Items',
    subtotal: 'Subtotal',
    shipping: 'Shipping',
    discount: 'Discount',
    total: 'Total',
    shipTo: 'Shipping to',
    tracking: 'Tracking',
    trackPackage: 'Track package',
    notFound: 'Order not found.',
    placedAt: 'Placed at',
    confirmEmail: 'A confirmation email will be sent once payment is confirmed.',
    status: {
      pending: 'Awaiting payment',
      paid: 'Paid',
      fulfilled: 'Shipped',
      cancelled: 'Cancelled',
      refunded: 'Refunded',
    },
  },
  footer: {
    rights: 'All rights reserved.',
  },
  language: 'Language',
}

// 中文翻译：结构与 en 完全一致
const zh: typeof en = {
  nav: {
    shopAll: '全部商品',
    journals: '手账本',
    stickers: '贴纸',
    cart: '购物车',
  },
  home: {
    title: '为慢生活而作的纸品',
    subtitle: '手工手账本、内芯与贴纸 —— 全球配送。',
    viewAll: '查看全部商品',
  },
  products: {
    shopAll: '全部商品',
    searchPrefix: '搜索',
    count: '件商品',
    empty: '未找到商品。',
    sortNewest: '最新',
    sortPriceAsc: '价格：从低到高',
    sortPriceDesc: '价格：从高到低',
  },
  product: {
    qty: '数量',
    addToCart: '加入购物车',
    outOfStock: '缺货',
    description: '商品描述',
  },
  cart: {
    title: '购物车',
    empty: '购物车是空的。',
    startShopping: '去购物',
    subtotal: '小计',
    checkout: '去结算',
    remove: '移除',
    shippingNote: '运费将在结算时计算。',
  },
  checkout: {
    title: '结算',
    contact: '联系方式',
    shippingAddress: '收货地址',
    email: '邮箱',
    recipient: '收件人',
    country: '国家/地区',
    state: '省/州',
    city: '城市',
    address: '详细地址',
    apartment: '门牌号、单元等',
    postalCode: '邮政编码',
    phone: '电话',
    orderNote: '订单备注',
    placeOrder: '提交订单',
    placing: '提交中...',
    summary: '订单摘要',
    emptyCart: '购物车是空的',
    fillRequired: '请填写必填项',
    serverShipping: '运费将根据目的地在服务端计算。',
  },
  order: {
    title: '订单',
    items: '商品',
    subtotal: '小计',
    shipping: '运费',
    discount: '优惠',
    total: '合计',
    shipTo: '收货信息',
    tracking: '物流',
    trackPackage: '查询物流',
    notFound: '未找到订单。',
    placedAt: '下单时间',
    confirmEmail: '支付确认后将发送确认邮件。',
    status: {
      pending: '待付款',
      paid: '已付款',
      fulfilled: '已发货',
      cancelled: '已取消',
      refunded: '已退款',
    },
  },
  footer: {
    rights: '保留所有权利。',
  },
  language: '语言',
}

const dictionaries = { en, zh }

export type Dictionary = typeof en

export function getDictionarySync(locale: Locale): Dictionary {
  return dictionaries[locale] ?? dictionaries.en
}

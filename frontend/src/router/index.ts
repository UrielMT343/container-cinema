import { createRouter, createWebHistory } from 'vue-router'
import ShowtimeList from '@/components/ShowtimeList.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: ShowtimeList,
    },
    {
      path: '/showtimes/:id/seats',
      name: 'showtime-seats',
      component: () => import('@/views/SeatView.vue'),
    },
    {
      path: '/checkout/pay',
      name: 'checkout-pay',
      component: () => import('@/views/CheckoutView.vue'),
    },
  ],
})

export default router

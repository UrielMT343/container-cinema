import { createRouter, createWebHistory } from "vue-router";
import ShowtimeList from "@/components/ShowtimeList.vue";
import { isSessionExpired } from "@/composables/useSessionExpired";

const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        {
            path: "/",
            name: "home",
            component: ShowtimeList,
        },
        {
            path: "/showtimes/:id/seats",
            name: "showtime-seats",
            component: () => import("@/views/SeatView.vue"),
        },
        {
            path: "/checkout/pay",
            name: "checkout-pay",
            component: () => import("@/views/CheckoutView.vue"),
        },
    ],
});

router.beforeEach((to, from, next) => {
    const rawSession = sessionStorage.getItem("checkout-session");

    if (!rawSession) {
        return next();
    }

    let session;
    try {
        session = JSON.parse(rawSession);
    } catch {
        sessionStorage.removeItem("checkout-session");
        return next();
    }

    // Wipe legacy/corrupt sessions that don't match the current format
    if (
        !session.expiresAt ||
        !session.step ||
        (session.step !== "SEATS" && session.step !== "PAYMENT")
    ) {
        sessionStorage.removeItem("checkout-session");
        return next();
    }

    if (new Date() > new Date(session.expiresAt)) {
        sessionStorage.removeItem("checkout-session");
        isSessionExpired.value = true;
        if (to.path === "/") {
            return next();
        }
        return next("/");
    }

    if (session.step === "PAYMENT") {
        if (to.path === "/checkout/pay") {
            return next();
        }
        return next("/checkout/pay");
    }

    if (session.step === "SEATS") {
        if (to.path === "/") {
            return next();
        }

        if (to.name === "showtime-seats") {
            const targetShowtimeId = Number(to.params.id);
            if (targetShowtimeId === session.showtimeId) {
                return next();
            }
            sessionStorage.removeItem("checkout-session");
            return next();
        }

        return next("/");
    }

    next();
});

export default router;

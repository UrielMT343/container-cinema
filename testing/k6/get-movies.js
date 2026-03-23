import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  stages: [
    { duration: "30s", target: 1200 },
    { duration: "2m", target: 1200 },
    { duration: "30s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<1500"],
    http_req_failed: ["rate<0.05"],
  },
};

export default function () {
  const baseUrl = "http://api.cinema.local";
  const targetUrl = `${baseUrl}/movies`;

  const res = http.get(targetUrl);

  check(res, {
    "is status 200": (r) => r.status === 200,
  });

  sleep(0.1);
}

import http from 'k6/http';
import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';

const movies = new SharedArray('all movies', function () {
    return JSON.parse(open('./data/movies.json'));
});

export const options = {
    vus: 50,
    duration: '1m',
};

export default function () {
    const baseUrl = __ENV.API_BASE_URL || 'http://api.cinema.local';
    const targetUrl = `${baseUrl}/movies`;

    const randomMovie = movies[Math.floor(Math.random() * movies.length)];

    const payload = JSON.stringify(randomMovie);

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(targetUrl, payload, params);

    check(res, {
        'is status 201 or 200': (r) => r.status === 201 || r.status === 200,
    });

    sleep(0.1);
}
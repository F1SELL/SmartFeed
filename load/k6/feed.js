import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(95)<1500'],
  },
};

const baseUrl = __ENV.BASE_URL || 'http://localhost:8080';

export default function () {
  const uid = `${__VU}_${__ITER}`;
  const username = `u_${uid}`;
  const password = 'password123';

  const regPayload = JSON.stringify({
    username,
    email: `${username}@example.com`,
    password,
  });

  const regRes = http.post(`${baseUrl}/api/v1/auth/register`, regPayload, {
    headers: { 'Content-Type': 'application/json' },
    responseCallback: http.expectedStatuses(201, 400),
  });
  check(regRes, {
    'register status': (r) => r.status === 201 || r.status === 400,
  });

  const loginRes = http.post(
    `${baseUrl}/api/v1/auth/login`,
    JSON.stringify({ username, password }),
    {
      headers: { 'Content-Type': 'application/json' },
      responseCallback: http.expectedStatuses(200),
    }
  );
  check(loginRes, {
    'login status 200': (r) => r.status === 200,
  });

  if (loginRes.status !== 200) {
    return;
  }

  const token = loginRes.json('token');
  const authHeaders = {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
  };

  const postRes = http.post(
    `${baseUrl}/api/v1/posts`,
    JSON.stringify({ content: `post from ${username}` }),
    {
      ...authHeaders,
      responseCallback: http.expectedStatuses(201),
    }
  );
  check(postRes, {
    'post created': (r) => r.status === 201,
  });

  sleep(0.2);

  const feedRes = http.get(`${baseUrl}/api/v1/feed?limit=20&offset=0`, {
    ...authHeaders,
    responseCallback: http.expectedStatuses(200),
  });
  check(feedRes, {
    'feed status 200': (r) => r.status === 200,
  });

  sleep(0.2);
}


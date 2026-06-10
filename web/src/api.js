// toda a comunicação com a API fica centralizada aqui

function apiUrl() {
  return (window.ENV && window.ENV.API_URL) || "http://localhost:8080";
}

async function request(path, { method = "GET", body, token } = {}) {
  const headers = { "Content-Type": "application/json" };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${apiUrl()}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    // erro especial para o app derrubar a sessão
    const error = new Error("sessão expirada");
    error.status = 401;
    throw error;
  }

  const data = await res.json().catch(() => null);

  if (!res.ok) {
    throw new Error((data && data.error) || `erro na requisição (${res.status})`);
  }

  return data;
}

export function login(email, password) {
  return request("/api/v1/auth/login", { method: "POST", body: { email, password } });
}

export function registerUser(name, email, password) {
  return request("/api/v1/auth/register", { method: "POST", body: { name, email, password } });
}

export function fetchMenu() {
  return request("/api/v1/menu");
}

export function fetchOrders(page = 1, status = "") {
  const params = new URLSearchParams({ page: String(page) });
  if (status) {
    params.set("status", status);
  }
  return request(`/api/v1/orders?${params.toString()}`);
}

export function fetchOrder(id) {
  return request(`/api/v1/orders/${id}`);
}

export function createOrder(token, order) {
  return request("/api/v1/orders", { method: "POST", body: order, token });
}

export function fetchStats() {
  return request("/api/v1/stats");
}

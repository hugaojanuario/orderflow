import { useCallback, useEffect, useState } from "react";
import { fetchOrders, fetchOrder } from "../api";
import OrderCard from "../components/OrderCard";

const POLL_INTERVAL_MS = 5000;

export default function OrdersScreen({ onApiError }) {
  const [orders, setOrders] = useState([]);
  const [statusFilter, setStatusFilter] = useState("");
  const [expanded, setExpanded] = useState(null);
  const [error, setError] = useState("");

  const load = useCallback(() => {
    fetchOrders(1, statusFilter)
      .then((response) => {
        setOrders(response.orders);
        setError("");
      })
      .catch((err) => {
        setError(err.message);
        onApiError(err);
      });
  }, [statusFilter, onApiError]);

  // polling a cada 5s para acompanhar o status mudando
  useEffect(() => {
    load();
    const interval = setInterval(load, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [load]);

  async function toggleDetails(id) {
    if (expanded && expanded.id === id) {
      setExpanded(null);
      return;
    }
    try {
      const order = await fetchOrder(id);
      setExpanded(order);
    } catch (err) {
      setError(err.message);
      onApiError(err);
    }
  }

  return (
    <div className="orders-screen">
      <div className="orders-header">
        <h2>Pedidos</h2>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Todos os status</option>
          <option value="received">Recebido</option>
          <option value="preparing">Em preparo</option>
          <option value="ready">Pronto</option>
          <option value="delivered">Entregue</option>
        </select>
      </div>

      {error && <p className="error">{error}</p>}
      {orders.length === 0 && !error && <p className="muted">Nenhum pedido encontrado.</p>}

      <div className="orders-list">
        {orders.map((order) => (
          <OrderCard
            key={order.id}
            order={order}
            expanded={expanded && expanded.id === order.id ? expanded : null}
            onToggleDetails={toggleDetails}
          />
        ))}
      </div>
    </div>
  );
}

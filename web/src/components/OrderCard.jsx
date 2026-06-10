const STATUS_LABELS = {
  received: "Recebido",
  preparing: "Em preparo",
  ready: "Pronto",
  delivered: "Entregue",
};

export default function OrderCard({ order, expanded, onToggleDetails }) {
  return (
    <div className="order-card">
      <div className="order-card-header">
        <div>
          <strong>#{order.id}</strong> — {order.customer_name}
        </div>
        <span className={`status status-${order.status}`}>
          {STATUS_LABELS[order.status] || order.status}
        </span>
      </div>
      <div className="order-card-body">
        <span>R$ {order.total.toFixed(2)}</span>
        <button className="link" onClick={() => onToggleDetails(order.id)}>
          {expanded ? "Ocultar histórico" : "Ver histórico"}
        </button>
      </div>
      {expanded && (
        <div className="order-history">
          <ul>
            {(expanded.history || []).map((entry) => (
              <li key={entry.id}>
                {STATUS_LABELS[entry.status] || entry.status} —{" "}
                {new Date(entry.created_at).toLocaleTimeString("pt-BR")}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

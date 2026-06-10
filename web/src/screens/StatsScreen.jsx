import { useCallback, useEffect, useState } from "react";
import { fetchStats } from "../api";
import StatCard from "../components/StatCard";

const POLL_INTERVAL_MS = 10000;

const STATUS_LABELS = {
  received: "Recebidos",
  preparing: "Em preparo",
  ready: "Prontos",
  delivered: "Entregues",
};

export default function StatsScreen({ onApiError }) {
  const [stats, setStats] = useState(null);
  const [error, setError] = useState("");

  const load = useCallback(() => {
    fetchStats()
      .then((response) => {
        setStats(response);
        setError("");
      })
      .catch((err) => {
        setError(err.message);
        onApiError(err);
      });
  }, [onApiError]);

  useEffect(() => {
    load();
    const interval = setInterval(load, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [load]);

  if (error) {
    return <p className="error">{error}</p>;
  }
  if (!stats) {
    return <p className="muted">Carregando...</p>;
  }

  return (
    <div className="stats-screen">
      <h2>Stats do dia — {stats.date}</h2>
      <div className="stats-grid">
        <StatCard label="Total de pedidos" value={stats.total_orders} />
        <StatCard label="Faturamento" value={`R$ ${stats.revenue.toFixed(2)}`} />
        <StatCard label="Tempo médio de preparo" value={`${Math.round(stats.avg_prep_seconds)}s`} />
        {Object.entries(STATUS_LABELS).map(([status, label]) => (
          <StatCard key={status} label={label} value={stats.orders_by_status[status] || 0} />
        ))}
      </div>
    </div>
  );
}

import { useEffect, useState } from "react";
import { fetchMenu, createOrder } from "../api";
import MenuItemCard from "../components/MenuItemCard";

export default function MenuScreen({ token, onApiError }) {
  const [items, setItems] = useState([]);
  const [cart, setCart] = useState({});
  const [customerName, setCustomerName] = useState("");
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    fetchMenu()
      .then(setItems)
      .catch((err) => {
        setError(err.message);
        onApiError(err);
      });
  }, [onApiError]);

  function addItem(id) {
    setCart((current) => ({ ...current, [id]: (current[id] || 0) + 1 }));
  }

  function removeItem(id) {
    setCart((current) => {
      const next = { ...current };
      if (next[id] > 1) {
        next[id] -= 1;
      } else {
        delete next[id];
      }
      return next;
    });
  }

  const cartItems = Object.entries(cart).map(([id, quantity]) => {
    const item = items.find((i) => i.id === Number(id));
    return { item, quantity };
  });

  const total = cartItems.reduce((sum, { item, quantity }) => sum + (item ? item.price * quantity : 0), 0);

  async function handleSubmit(event) {
    event.preventDefault();
    setError("");
    setSuccess("");
    setLoading(true);

    try {
      const order = await createOrder(token, {
        customer_name: customerName,
        items: Object.entries(cart).map(([id, quantity]) => ({
          menu_item_id: Number(id),
          quantity,
        })),
      });
      setSuccess(`Pedido #${order.id} criado! Acompanhe na aba Pedidos.`);
      setCart({});
      setCustomerName("");
    } catch (err) {
      setError(err.message);
      onApiError(err);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="menu-screen">
      <section className="menu-list">
        <h2>Cardápio</h2>
        <div className="menu-grid">
          {items.map((item) => (
            <MenuItemCard
              key={item.id}
              item={item}
              quantity={cart[item.id] || 0}
              onAdd={addItem}
              onRemove={removeItem}
            />
          ))}
        </div>
      </section>

      <aside className="cart">
        <h2>Pedido</h2>
        {cartItems.length === 0 && <p className="muted">Nenhum item selecionado.</p>}
        <ul>
          {cartItems.map(({ item, quantity }) =>
            item ? (
              <li key={item.id}>
                {quantity}x {item.name} — R$ {(item.price * quantity).toFixed(2)}
              </li>
            ) : null
          )}
        </ul>
        <p className="cart-total">Total: R$ {total.toFixed(2)}</p>

        <form onSubmit={handleSubmit}>
          <input
            type="text"
            placeholder="Nome do cliente"
            value={customerName}
            onChange={(e) => setCustomerName(e.target.value)}
            required
          />
          <button type="submit" disabled={loading || cartItems.length === 0}>
            {loading ? "Enviando..." : "Fazer pedido"}
          </button>
        </form>

        {error && <p className="error">{error}</p>}
        {success && <p className="success">{success}</p>}
      </aside>
    </div>
  );
}

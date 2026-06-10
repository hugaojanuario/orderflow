export default function MenuItemCard({ item, quantity, onAdd, onRemove }) {
  return (
    <div className="menu-item-card">
      <h3>{item.name}</h3>
      <p className="muted">{item.description}</p>
      <p className="price">R$ {item.price.toFixed(2)}</p>
      <div className="quantity-controls">
        <button onClick={() => onRemove(item.id)} disabled={quantity === 0}>
          −
        </button>
        <span>{quantity}</span>
        <button onClick={() => onAdd(item.id)}>+</button>
      </div>
    </div>
  );
}

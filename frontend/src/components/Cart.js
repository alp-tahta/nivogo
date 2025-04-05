import React, { useState, useEffect } from 'react';
import CartItem from './CartItem';
import '../styles/Cart.css';

const Cart = () => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await fetch('http://localhost:8081/product?ids=1,2');
        if (!response.ok) {
          throw new Error('Failed to fetch products');
        }
        const products = await response.json();
        // Initialize products with quantity 1
        setItems(products.map(product => ({
          ...product,
          quantity: 1
        })));
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const handleQuantityChange = (id, newQuantity) => {
    if (newQuantity < 1) return;
    setItems(items.map(item =>
      item.id === id ? { ...item, quantity: newQuantity } : item
    ));
  };

  const handleRemoveItem = (id) => {
    setItems(items.filter(item => item.id !== id));
  };

  const calculateSubtotal = () => {
    return items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  };

  const subtotal = calculateSubtotal();

  if (loading) {
    return <div className="cart-container">Loading...</div>;
  }

  if (error) {
    return <div className="cart-container">Error: {error}</div>;
  }

  return (
    <div className="cart-container">
      <h1>Cart</h1>
      
      <div className="cart-content">
        <div className="cart-items">
          <div className="cart-header">
            <span>Product</span>
            <span>Total</span>
          </div>
          
          {items.map(item => (
            <CartItem
              key={item.id}
              {...item}
              total={item.price * item.quantity}
              onQuantityChange={(newQuantity) => handleQuantityChange(item.id, newQuantity)}
              onRemove={() => handleRemoveItem(item.id)}
            />
          ))}
        </div>

        <div className="cart-summary">
          <h2>Cart totals</h2>
          <div className="coupon-section">
            <button className="coupon-button">Add a coupon</button>
          </div>
          <div className="totals">
            <div className="subtotal">
              <span>Subtotal</span>
              <span>${subtotal.toFixed(2)}</span>
            </div>
            <div className="total">
              <span>Total</span>
              <span>${subtotal.toFixed(2)}</span>
            </div>
          </div>
          <button className="checkout-button">
            Proceed to Checkout
          </button>
        </div>
      </div>
    </div>
  );
};

export default Cart; 
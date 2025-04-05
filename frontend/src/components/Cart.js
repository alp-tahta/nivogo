import React, { useState } from 'react';
import CartItem from './CartItem';
import '../styles/Cart.css';

const Cart = () => {
  const [items, setItems] = useState([
    {
      id: 1,
      name: "Bird's Nest Fern",
      price: 22.00,
      description: "The Bird's Nest Fern is a tropical plant known for its vibrant green, wavy fronds...",
      image: "/images/birds-nest-fern.jpg",
      quantity: 3
    },
    {
      id: 2,
      name: "Ctenanthe",
      price: 45.00,
      description: "The Ctenanthe, also known as the Prayer Plant, is a stunning tropical plant with bold...",
      image: "/images/ctenanthe.jpg",
      quantity: 1
    }
  ]);

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
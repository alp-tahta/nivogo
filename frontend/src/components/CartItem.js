import React from 'react';
import '../styles/CartItem.css';

const CartItem = ({ name, price = 0, description, quantity = 1, total = 0, onQuantityChange, onRemove }) => {

  const formatPrice = (value) => {
    return typeof value === 'number' ? value.toFixed(2) : '0.00';
  };

  return (
    <div className="cart-item">
      <div className="cart-item-main">
        <div className="cart-item-details">
          <h3 className="cart-item-name">{name || 'Unnamed Product'}</h3>
          <p className="cart-item-price">${formatPrice(price)}</p>
          <p className="cart-item-description">{description || 'No description available'}</p>
          
          <div className="cart-item-controls">
            <div className="quantity-controls">
              <button 
                onClick={() => onQuantityChange(quantity - 1)}
                className="quantity-button"
              >
                −
              </button>
              <span className="quantity">{quantity}</span>
              <button 
                onClick={() => onQuantityChange(quantity + 1)}
                className="quantity-button"
              >
                +
              </button>
            </div>
            <button onClick={onRemove} className="remove-button">
              Remove item
            </button>
          </div>
        </div>
      </div>
      <div className="cart-item-total">
        ${formatPrice(total)}
      </div>
    </div>
  );
};

export default CartItem; 
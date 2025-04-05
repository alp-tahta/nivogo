import React from 'react';
import '../styles/CartItem.css';

const CartItem = ({ image, name, price, description, quantity, total, onQuantityChange, onRemove }) => {
  return (
    <div className="cart-item">
      <div className="cart-item-main">
        <img src={image} alt={name} className="cart-item-image" />
        <div className="cart-item-details">
          <h3 className="cart-item-name">{name}</h3>
          <p className="cart-item-price">${price.toFixed(2)}</p>
          <p className="cart-item-description">{description}</p>
          
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
        ${total.toFixed(2)}
      </div>
    </div>
  );
};

export default CartItem; 
import React, { useState, useEffect } from 'react';
import CartItem from './CartItem';
import '../styles/Cart.css';

const API_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081';

const Cart = () => {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [checkoutLoading, setCheckoutLoading] = useState(false);
  const [checkoutMessage, setCheckoutMessage] = useState('');
  const [isLoadingDone, setIsLoadingDone] = useState(false);
  const [results, setResults] = useState(null);

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        const response = await fetch(`${API_URL}/product?ids=1,2`);
        
        if (!response.ok) {
          throw new Error(`Failed to fetch products: ${response.status}`);
        }
        
        const products = await response.json();
        
        if (!Array.isArray(products)) {
          throw new Error('Invalid response format: expected an array');
        }
        
        // Create sample cart items if API returns empty or invalid data
        if (products.length === 0) {
          const sampleItems = [
            {
              id: 'sample-1',
              name: 'Sample Product 1',
              price: 19.99,
              description: 'This is a sample product description',
              quantity: 1
            },
            {
              id: 'sample-2',
              name: 'Sample Product 2',
              price: 29.99,
              description: 'This is another sample product description',
              quantity: 1
            }
          ];
          
          setItems(sampleItems);
          setLoading(false);
          return;
        }
        
        // Map only essential fields from API response - using capitalized property names
        const cartItems = products.map((product, index) => {
          // Check if product has the expected structure
          if (!product || typeof product !== 'object') {
            return null;
          }
          
          return {
            id: product.ID || `temp-id-${index}`,
            name: product.Name || 'Unnamed Product',
            price: typeof product.Price === 'number' ? product.Price : 0,
            description: product.Description || 'No description available',
            quantity: 1
          };
        }).filter(Boolean); // Remove any null items
        
        // If no valid items were created, use sample data
        if (cartItems.length === 0) {
          const sampleItems = [
            {
              id: 'sample-1',
              name: 'Sample Product 1',
              price: 19.99,
              description: 'This is a sample product description',
              quantity: 1
            },
            {
              id: 'sample-2',
              name: 'Sample Product 2',
              price: 29.99,
              description: 'This is another sample product description',
              quantity: 1
            }
          ];
          
          setItems(sampleItems);
        } else {
          setItems(cartItems);
        }
      } catch (err) {
        console.error('Error fetching products:', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const handleQuantityChange = (id, newQuantity) => {
    if (newQuantity < 1) return;
    const updatedItems = items.map(item =>
      item.id === id ? { ...item, quantity: newQuantity } : item
    );
    setItems(updatedItems);
  };

  const handleRemoveItem = (id) => {
    const updatedItems = items.filter(item => item.id !== id);
    setItems(updatedItems);
  };

  const orderData = {
    items: items.map(item => ({
      product: {
        id: item.id,
        name: item.name,
        price: item.price
      },
      quantity: item.quantity
    }))
  };

  useEffect(() => {
    if (isLoadingDone) {
      console.log('Loading is done, sending order to API');
      const sendOrder = async () => {
        try {
          const response = await fetch(`http://localhost:10081/orders`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json'
            },
            mode: 'cors',
            credentials: 'include',
            body: JSON.stringify(orderData)
          });
          
          if (!response.ok) {
            throw new Error(`Failed to submit order: ${response.status}`);
          }
          
          const result = await response.json();
          console.log('Order submitted successfully:', result);
          setResults(result);
          setCheckoutMessage('Order submitted successfully!');
        } catch (err) {
          console.error('Error submitting order:', err);
          setCheckoutMessage('Failed to submit order. Please try again.');
        }
      };
      
      sendOrder();
    }
  }, [isLoadingDone]);

  const handleCheckout = () => {
    setCheckoutLoading(true);
    setCheckoutMessage('');
    
    // Simulate a checkout process with a 2-second delay
    setTimeout(() => {
      setCheckoutLoading(false);
      setIsLoadingDone(true);
    }, 2000);
  };

  const calculateSubtotal = () => {
    return items.reduce((sum, item) => {
      const price = typeof item.price === 'number' ? item.price : 0;
      const quantity = typeof item.quantity === 'number' ? item.quantity : 0;
      return sum + (price * quantity);
    }, 0);
  };

  const subtotal = calculateSubtotal();

  if (loading) {
    return <div className="cart-container">Loading...</div>;
  }

  if (error) {
    return <div className="cart-container">Error: {error}</div>;
  }

  if (items.length === 0) {
    return <div className="cart-container">Your cart is empty</div>;
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
              total={(typeof item.price === 'number' ? item.price : 0) * (typeof item.quantity === 'number' ? item.quantity : 0)}
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
          <button 
            className="checkout-button"
            onClick={handleCheckout}
            disabled={checkoutLoading}
          >
            {checkoutLoading ? 'Processing...' : 'Proceed to Checkout'}
          </button>
          {checkoutMessage && (
            <div className={`checkout-message ${checkoutMessage.includes('Failed') ? 'error' : 'success'}`}>
              {checkoutMessage}
            </div>
          )}
        </div>
      </div>
      
      {/* Loading Modal */}
      {checkoutLoading && (
        <div className="loading-modal">
          <div className="loading-modal-content">
            <div className="loading-spinner"></div>
            <p>Processing your order...</p>
          </div>
        </div>
      )}
    </div>
  );
};

export default Cart; 
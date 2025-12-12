
document.addEventListener('DOMContentLoaded', () => {
    // Keep non-Vue related modal initializations and other variables for now
    const cartButton = document.getElementById('cart-button');
    const cartModal = new bootstrap.Modal(document.getElementById('cart-modal'));
    const cartItems = document.getElementById('cart-items');
    const cartCount = document.getElementById('cart-count');
    const cartTotal = document.getElementById('cart-total');
    const loginButton = document.getElementById('login-button');
    const loginModal = new bootstrap.Modal(document.getElementById('login-modal'));
    const loginForm = document.getElementById('login-form');
    const registerForm = document.getElementById('register-form');
    const showRegister = document.getElementById('show-register');
    const showLogin = document.getElementById('show-login');
    const ordersButton = document.getElementById('orders-button');
    const ordersModal = new bootstrap.Modal(document.getElementById('orders-modal'));
    const orderList = document.getElementById('order-list');
    const checkoutButton = document.getElementById('checkout-button');
    const searchBar = document.getElementById('search-bar');
    const categoryFilter = document.getElementById('category-filter');
    const productModal = new bootstrap.Modal(document.getElementById('product-modal'));
    const productModalTitle = document.getElementById('product-modal-title');
    const productModalBody = document.getElementById('product-modal-body');

    let cartId = localStorage.getItem('cartId');
    let loggedIn = false;
    let userId = null;

    const { createApp } = Vue;

    const app = createApp({
        data() {
            return {
                products: [],
                categories: [],
                searchTerm: '',
                selectedCategory: '',
                selectedProduct: null,
                reviews: [],
                reviewRating: 5,
                reviewComment: '',
                cartItems: [],
                cartTotal: 0,
                cartItemCount: 0,
                loggedIn: false,
                userId: null,
                loginUsername: '',
                loginPassword: '',
                registerUsername: '',
                registerPassword: ''
            };
        },
        methods: {
            // ... (existing methods: fetchProducts, fetchCategories, etc.) ...
            async fetchProducts() {
                const response = await fetch(`/api/products?search=${this.searchTerm}&category=${this.selectedCategory}`);
                const products = await response.json();
                this.products = products || [];
            },
            async fetchCategories() {
                const response = await fetch('/api/categories');
                const categories = await response.json();
                this.categories = categories || [];
            },
            async addToCart(productId) {
                console.log("Add to cart button clicked via Vue");
                const cartId = await getOrCreateCart();
                const response = await fetch(`/api/cart/${cartId}/items`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ productId: parseInt(productId), quantity: 1 })
                });
                if (response.ok) {
                    this.renderCartItems(); // Refresh cart state
                }
            },
            async showProductDetails(productId) {
                console.log(`Product card clicked for product ID: ${productId}`);
                const productResponse = await fetch(`/api/products/${productId}`);
                this.selectedProduct = await productResponse.json();

                const reviewsResponse = await fetch(`/api/products/${productId}/reviews`);
                this.reviews = await reviewsResponse.json() || [];
                
                this.reviewRating = 5;
                this.reviewComment = '';

                productModal.show();
            },
            async submitReview() {
                if (!this.loggedIn) {
                    alert('You must be logged in to leave a review.');
                    loginModal.show();
                    return;
                }
                if (!this.selectedProduct) return;

                const response = await fetch(`/api/products/${this.selectedProduct.id}/reviews`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ 
                        rating: this.reviewRating, 
                        comment: this.reviewComment 
                    })
                });

                if (response.ok) {
                    // Refresh reviews
                    const reviewsResponse = await fetch(`/api/products/${this.selectedProduct.id}/reviews`);
                    this.reviews = await reviewsResponse.json() || [];
                    this.reviewRating = 5;
                    this.reviewComment = '';
                } else {
                    alert('Failed to submit review.');
                }
            },
            async renderCartItems() {
                const cartId = await getOrCreateCart();
                const response = await fetch(`/api/cart/${cartId}`);
                const cart = await response.json();

                this.cartItems = cart.items || [];
                this.cartItemCount = this.cartItems.length;
                this.cartTotal = this.cartItems.reduce((total, item) => {
                    return total + (item.product.price * item.quantity);
                }, 0);
            },
            showCart() {
                this.renderCartItems();
                cartModal.show();
            },
            // --- Auth Methods ---
            async checkLoginStatus() {
                const response = await fetch('/api/me');
                const data = await response.json();
                this.loggedIn = data.loggedIn;
                this.userId = data.loggedIn ? data.userID : null;
            },
            async toggleLogin() {
                if (this.loggedIn) {
                    // Logout
                    await fetch('/api/logout', { method: 'POST' });
                    this.loggedIn = false;
                    this.userId = null;
                } else {
                    // Show login modal
                    this.loginUsername = '';
                    this.loginPassword = '';
                    loginModal.show();
                }
            },
            async handleLogin() {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username: this.loginUsername, password: this.loginPassword })
                });

                if (response.ok) {
                    alert('Login successful!');
                    await this.checkLoginStatus();
                    loginModal.hide();
                } else {
                    alert('Invalid credentials');
                }
            },
            async handleRegister() {
                const response = await fetch('/api/register', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ username: this.registerUsername, password: this.registerPassword })
                });

                if (response.ok) {
                    alert('Registration successful! Please log in.');
                    this.showLoginForm();
                } else {
                    alert('Registration failed.');
                }
            },
            showRegisterForm() {
                document.getElementById('login-form').style.display = 'none';
                document.getElementById('register-form').style.display = 'block';
            },
            showLoginForm() {
                document.getElementById('register-form').style.display = 'none';
                document.getElementById('login-form').style.display = 'block';
            }
        },
        mounted() {
            this.fetchProducts();
            this.fetchCategories();
            this.renderCartItems();
            this.checkLoginStatus();
        }
    }).mount('#app');


    // --- Functions to be refactored into Vue later ---

    // Function to check login status
    const checkLoginStatus = async () => {
        const response = await fetch('/api/me');
        const data = await response.json();
        loggedIn = data.loggedIn;
        if (loggedIn) {
            userId = data.userID;
            loginButton.textContent = 'Logout';
        } else {
            userId = null;
            loginButton.textContent = 'Login';
        }
    };

    // Function to get or create a cart

    // Show/hide login modal
    loginButton.addEventListener('click', () => {
        if (loggedIn) {
            // Logout
            fetch('/api/logout', { method: 'POST' }).then(() => {
                loggedIn = false;
                userId = null;
                loginButton.textContent = 'Login';
            });
        } else {
            loginModal.show();
        }
    });

    // Show/hide orders modal
    ordersButton.addEventListener('click', () => {
        if (!loggedIn) {
            alert('You must be logged in to view your orders.');
            loginModal.show();
            return;
        }
        renderOrderHistory();
        ordersModal.show();
    });

    // Switch between login and register forms
    showRegister.addEventListener('click', (e) => {
        e.preventDefault();
        loginForm.style.display = 'none';
        registerForm.style.display = 'block';
    });

    showLogin.addEventListener('click', (e) => {
        e.preventDefault();
        registerForm.style.display = 'none';
        loginForm.style.display = 'block';
    });

    // Handle login form submission
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('login-username').value;
        const password = document.getElementById('login-password').value;

        const response = await fetch('/api/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            alert('Login successful!');
            checkLoginStatus();
            loginModal.hide();
        } else {
            alert('Invalid credentials');
        }
    });

    // Handle register form submission
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const username = document.getElementById('register-username').value;
        const password = document.getElementById('register-password').value;

        const response = await fetch('/api/register', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        if (response.ok) {
            alert('Registration successful! Please log in.');
            registerForm.style.display = 'none';
            loginForm.style.display = 'block';
        } else {
            alert('Registration failed.');
        }
    });

    // Handle checkout
    checkoutButton.addEventListener('click', async () => {
        if (!loggedIn) {
            alert('You must be logged in to checkout.');
            loginModal.show();
            return;
        }

        const cartIdVal = await getOrCreateCart();
        const response = await fetch('/api/orders', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ cartId: parseInt(cartIdVal) })
        });

        if (response.ok) {
            alert('Order placed successfully!');
            cartModal.hide();
            localStorage.removeItem('cartId');
            cartId = null;
            updateCartCount();
        } else {
            alert('Failed to place order.');
        }
    });

    // Function to render order history
    const renderOrderHistory = async () => {
        const response = await fetch('/api/orders');
        const orders = await response.json();

        orderList.innerHTML = '';

        if (orders) {
            orders.forEach(order => {
                const orderDiv = document.createElement('div');
                orderDiv.className = 'accordion-item';

                let itemsHTML = '';
                let total = 0;
                order.items.forEach(item => {
                    itemsHTML += `<li>${item.quantity} x Product ID ${item.productId} at $${item.price.toFixed(2)}</li>`;
                    total += item.quantity * item.price;
                });

                orderDiv.innerHTML = `
                    <h2 class="accordion-header">
                        <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse" data-bs-target="#order-${order.id}">
                            Order #${order.id} - ${new Date(order.createdAt).toLocaleDateString()} - Total: $${total.toFixed(2)}
                        </button>
                    </h2>
                    <div id="order-${order.id}" class="accordion-collapse collapse">
                        <div class="accordion-body">
                            <ul>${itemsHTML}</ul>
                        </div>
                    </div>
                `;
                orderList.appendChild(orderDiv);
            });
        }
    };

    // Initial setup
    checkLoginStatus();
    updateCartCount();
});

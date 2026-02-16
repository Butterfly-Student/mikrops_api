// Payment Portal Application
class PaymentPortal {
    constructor() {
        this.customer = null;
        this.invoices = [];
        this.selectedInvoice = null;
        this.selectedPaymentMethod = null;
        this.init();
    }

    init() {
        this.bindEvents();
        this.checkAuth();
    }

    bindEvents() {
        // Login form
        document.getElementById('login-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        // Logout
        document.getElementById('logout-btn').addEventListener('click', () => this.handleLogout());

        // Payment methods
        document.querySelectorAll('.payment-method').forEach(method => {
            method.addEventListener('click', () => this.selectPaymentMethod(method));
        });

        // Pay now button
        document.getElementById('pay-now-btn').addEventListener('click', () => this.initiatePayment());

        // Navigation
        document.getElementById('back-to-dashboard').addEventListener('click', () => this.showPage('dashboard'));
        document.getElementById('back-to-dashboard-success').addEventListener('click', () => this.showPage('dashboard'));
        document.getElementById('back-to-dashboard-failed').addEventListener('click', () => this.showPage('dashboard'));

        // Modal
        document.getElementById('modal-close').addEventListener('click', () => this.closeModal());
        document.getElementById('modal-close-footer').addEventListener('click', () => this.closeModal());
        document.getElementById('modal-action').addEventListener('click', () => this.goToPayment());
        document.getElementById('modal-download').addEventListener('click', () => this.downloadInvoicePDF());
    }

    checkAuth() {
        const customerId = localStorage.getItem(CONFIG.STORAGE_KEYS.CUSTOMER_ID);
        if (customerId) {
            this.showPage('dashboard');
            this.loadCustomerData();
        } else {
            this.showPage('login');
        }
        this.hideLoading();
    }

    async handleLogin() {
        const customerCode = document.getElementById('customer-code').value.trim();
        if (!customerCode) {
            this.showError('login-error', 'Please enter customer code or phone number');
            return;
        }

        this.showLoading();
        try {
            const customer = await this.fetchCustomer(customerCode);
            if (customer) {
                this.customer = customer;
                localStorage.setItem(CONFIG.STORAGE_KEYS.CUSTOMER_ID, customer.id);
                localStorage.setItem(CONFIG.STORAGE_KEYS.CUSTOMER_CODE, customer.customer_code);
                localStorage.setItem(CONFIG.STORAGE_KEYS.CUSTOMER_NAME, customer.full_name);
                this.showPage('dashboard');
                this.loadCustomerData();
            } else {
                this.showError('login-error', 'Customer not found');
            }
        } catch (error) {
            this.showError('login-error', error.message || 'Failed to login');
        } finally {
            this.hideLoading();
        }
    }

    async fetchCustomer(customerCode) {
        try {
            const response = await fetch(`${CONFIG.API_BASE_URL}${CONFIG.ENDPOINTS.CUSTOMERS}/${customerCode}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error('Customer not found');
            }

            return await response.json();
        } catch (error) {
            console.error('Error fetching customer:', error);
            throw error;
        }
    }

    async loadCustomerData() {
        try {
            document.getElementById('customer-name').textContent = this.customer.full_name;
            await this.loadInvoices();
            this.updateStats();
        } catch (error) {
            console.error('Error loading customer data:', error);
        }
    }

    async loadInvoices() {
        try {
            const customerId = this.customer.id;
            const response = await fetch(`${CONFIG.API_BASE_URL}${CONFIG.ENDPOINTS.INVOICES}?customer_id=${customerId}`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                },
            });

            if (!response.ok) {
                throw new Error('Failed to load invoices');
            }

            const data = await response.json();
            this.invoices = data.data || [];
            this.renderInvoices();
        } catch (error) {
            console.error('Error loading invoices:', error);
            this.invoices = [];
            this.renderInvoices();
        }
    }

    renderInvoices() {
        const container = document.getElementById('invoices-container');
        
        if (this.invoices.length === 0) {
            container.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-file-invoice-dollar"></i>
                    <p>No invoices found</p>
                </div>
            `;
            return;
        }

        container.innerHTML = this.invoices.map(invoice => `
            <div class="invoice-card ${invoice.status.toLowerCase()}" data-invoice-id="${invoice.id}">
                <div class="invoice-header">
                    <span class="invoice-number">${invoice.invoice_number}</span>
                    <span class="invoice-status ${invoice.status.toLowerCase()}">${CONFIG.STATUS_LABELS[invoice.status] || invoice.status}</span>
                </div>
                <div class="invoice-body">
                    <div class="invoice-info">
                        <div class="invoice-date">
                            <i class="fas fa-calendar"></i>
                            <span>Due: ${this.formatDate(invoice.due_date)}</span>
                        </div>
                        <div class="invoice-amount">
                            <span class="label">Total:</span>
                            <span class="amount">${this.formatCurrency(invoice.total_amount)}</span>
                        </div>
                    </div>
                    <div class="invoice-actions">
                        <button class="btn btn-view" onclick="window.paymentPortal.showInvoiceDetails('${invoice.id}')">
                            <i class="fas fa-eye"></i> View
                        </button>
                        ${invoice.status !== 'paid' ? `
                            <button class="btn btn-pay" onclick="window.paymentPortal.selectInvoiceForPayment('${invoice.id}')">
                                <i class="fas fa-credit-card"></i> Pay
                            </button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `).join('');
    }

    updateStats() {
        const pending = this.invoices.filter(i => i.status === 'sent' || i.status === 'partial');
        const paid = this.invoices.filter(i => i.status === 'paid');
        const overdue = this.invoices.filter(i => i.status === 'overdue');

        const pendingAmount = pending.reduce((sum, i) => sum + (i.total_amount - i.paid_amount), 0);
        const paidAmount = paid.reduce((sum, i) => sum + i.paid_amount, 0);
        const overdueAmount = overdue.reduce((sum, i) => sum + (i.total_amount - i.paid_amount), 0);

        document.getElementById('pending-amount').textContent = this.formatCurrency(pendingAmount);
        document.getElementById('paid-amount').textContent = this.formatCurrency(paidAmount);
        document.getElementById('overdue-amount').textContent = this.formatCurrency(overdueAmount);
    }

    showInvoiceDetails(invoiceId) {
        const invoice = this.invoices.find(i => i.id === invoiceId);
        if (!invoice) return;

        this.selectedInvoice = invoice;

        const modalBody = document.getElementById('modal-body');
        modalBody.innerHTML = `
            <div class="invoice-detail-view">
                <div class="detail-row">
                    <span class="label">Invoice Number:</span>
                    <span class="value">${invoice.invoice_number}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Status:</span>
                    <span class="value invoice-status ${invoice.status.toLowerCase()}">${CONFIG.STATUS_LABELS[invoice.status] || invoice.status}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Issue Date:</span>
                    <span class="value">${this.formatDate(invoice.issue_date)}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Due Date:</span>
                    <span class="value ${invoice.status === 'overdue' ? 'overdue' : ''}">${this.formatDate(invoice.due_date)}</span>
                </div>
                <hr>
                <div class="detail-row">
                    <span class="label">Subtotal:</span>
                    <span class="value">${this.formatCurrency(invoice.subtotal)}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Tax (11%):</span>
                    <span class="value">${this.formatCurrency(invoice.tax_amount)}</span>
                </div>
                ${invoice.late_fee > 0 ? `
                    <div class="detail-row">
                        <span class="label">Late Fee:</span>
                        <span class="value warning">${this.formatCurrency(invoice.late_fee)}</span>
                    </div>
                ` : ''}
                <div class="detail-row total">
                    <span class="label">Total Amount:</span>
                    <span class="value large">${this.formatCurrency(invoice.total_amount)}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Paid Amount:</span>
                    <span class="value">${this.formatCurrency(invoice.paid_amount)}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Balance:</span>
                    <span class="value ${invoice.total_amount - invoice.paid_amount > 0 ? 'warning' : 'success'}">
                        ${this.formatCurrency(invoice.total_amount - invoice.paid_amount)}
                    </span>
                </div>
            </div>
        `;

        // Show/hide action buttons based on status
        const modalAction = document.getElementById('modal-action');
        const modalDownload = document.getElementById('modal-download');

        if (invoice.status !== 'paid') {
            modalAction.classList.remove('hidden');
        } else {
            modalAction.classList.add('hidden');
        }

        modalDownload.classList.remove('hidden');

        this.openModal();
    }

    selectInvoiceForPayment(invoiceId) {
        const invoice = this.invoices.find(i => i.id === invoiceId);
        if (!invoice) return;

        this.selectedInvoice = invoice;
        this.goToPayment();
    }

    goToPayment() {
        if (!this.selectedInvoice) return;

        this.showPage('payment');
        this.renderPaymentDetails();
        this.resetPaymentMethodSelection();
    }

    renderPaymentDetails() {
        const content = document.getElementById('invoice-details-content');
        const invoice = this.selectedInvoice;

        content.innerHTML = `
            <div class="payment-invoice-detail">
                <div class="detail-row">
                    <span class="label">Invoice Number:</span>
                    <span class="value">${invoice.invoice_number}</span>
                </div>
                <div class="detail-row">
                    <span class="label">Due Date:</span>
                    <span class="value">${this.formatDate(invoice.due_date)}</span>
                </div>
                <hr>
                <div class="detail-row total">
                    <span class="label">Amount Due:</span>
                    <span class="value large">${this.formatCurrency(invoice.total_amount - invoice.paid_amount)}</span>
                </div>
            </div>
        `;
    }

    selectPaymentMethod(element) {
        document.querySelectorAll('.payment-method').forEach(el => el.classList.remove('selected'));
        element.classList.add('selected');
        this.selectedPaymentMethod = element.dataset.method;
        document.getElementById('pay-now-btn').disabled = false;
    }

    resetPaymentMethodSelection() {
        document.querySelectorAll('.payment-method').forEach(el => el.classList.remove('selected'));
        this.selectedPaymentMethod = null;
        document.getElementById('pay-now-btn').disabled = true;
    }

    async initiatePayment() {
        if (!this.selectedInvoice || !this.selectedPaymentMethod) {
            alert('Please select a payment method');
            return;
        }

        this.showLoading();
        document.getElementById('payment-error').classList.add('hidden');

        try {
            const response = await fetch(`${CONFIG.API_BASE_URL}${CONFIG.ENDPOINTS.PAYMENT_CREATE}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    invoice_id: this.selectedInvoice.id,
                    payment_method: this.selectedPaymentMethod,
                    amount: this.selectedInvoice.total_amount - this.selectedInvoice.paid_amount,
                }),
            });

            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.message || 'Failed to create payment');
            }

            const result = await response.json();
            
            // Redirect to Xendit payment page
            if (result.data && result.data.payment_url) {
                window.location.href = result.data.payment_url;
            } else {
                throw new Error('Payment URL not received');
            }
        } catch (error) {
            console.error('Error initiating payment:', error);
            this.showError('payment-error', error.message || 'Failed to initiate payment');
        } finally {
            this.hideLoading();
        }
    }

    async downloadInvoicePDF() {
        if (!this.selectedInvoice) return;

        try {
            const response = await fetch(`${CONFIG.API_BASE_URL}${CONFIG.ENDPOINTS.INVOICES}/${this.selectedInvoice.id}/pdf`, {
                method: 'GET',
            });

            if (!response.ok) {
                throw new Error('Failed to download PDF');
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `invoice-${this.selectedInvoice.invoice_number}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (error) {
            console.error('Error downloading PDF:', error);
            alert('Failed to download invoice PDF');
        }
    }

    handleLogout() {
        localStorage.clear();
        this.customer = null;
        this.invoices = [];
        this.selectedInvoice = null;
        this.showPage('login');
    }

    // UI Helpers
    showPage(pageId) {
        document.querySelectorAll('.page').forEach(page => page.classList.add('hidden'));
        document.getElementById(`${pageId}-page`).classList.remove('hidden');
    }

    showLoading() {
        document.getElementById('loading-screen').classList.remove('hidden');
    }

    hideLoading() {
        document.getElementById('loading-screen').classList.add('hidden');
    }

    showError(elementId, message) {
        const element = document.getElementById(elementId);
        element.textContent = message;
        element.classList.remove('hidden');
    }

    openModal() {
        document.getElementById('modal').classList.remove('hidden');
    }

    closeModal() {
        document.getElementById('modal').classList.add('hidden');
    }

    formatCurrency(amount) {
        return new Intl.NumberFormat('id-ID', {
            style: 'currency',
            currency: 'IDR',
            minimumFractionDigits: 0,
            maximumFractionDigits: 0,
        }).format(amount);
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        return date.toLocaleDateString('id-ID', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
        });
    }
}

// Check for payment status in URL
function checkPaymentStatus() {
    const urlParams = new URLSearchParams(window.location.search);
    const status = urlParams.get('status');
    const invoiceId = urlParams.get('invoice_id');

    if (status && invoiceId) {
        if (status === 'success') {
            showPaymentSuccess(invoiceId);
        } else if (status === 'failed' || status === 'cancelled') {
            showPaymentFailed(invoiceId, status);
        }
    }
}

function showPaymentSuccess(invoiceId) {
    const successDetails = document.getElementById('success-details');
    successDetails.innerHTML = `
        <p>Invoice ID: ${invoiceId}</p>
        <p>Your payment has been processed successfully.</p>
    `;
    document.getElementById('dashboard-page').classList.add('hidden');
    document.getElementById('success-page').classList.remove('hidden');
}

function showPaymentFailed(invoiceId, status) {
    const failedDetails = document.getElementById('failed-details');
    failedDetails.innerHTML = `
        <p>Invoice ID: ${invoiceId}</p>
        <p>Status: ${status.toUpperCase()}</p>
        <p>Please try again or contact support if the problem persists.</p>
    `;
    document.getElementById('dashboard-page').classList.add('hidden');
    document.getElementById('failed-page').classList.remove('hidden');
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.paymentPortal = new PaymentPortal();
    checkPaymentStatus();
});

// Payment Portal Configuration
const CONFIG = {
    // API Base URL - change this to match your backend
    API_BASE_URL: window.location.origin,
    
    // API Endpoints
    ENDPOINTS: {
        LOGIN: '/auth/login',
        CUSTOMER: '/customers',
        INVOICES: '/billing/invoices',
        INVOICE: '/billing/invoices',
        PAYMENT_CREATE: '/payments/create',
        PAYMENT_STATUS: '/payments',
        PAYMENT_WEBHOOK: '/webhooks/xendit',
        PAYMENT_RECEIPT: '/payments',
        CUSTOMER_BILLING_INFO: '/customers',
    },
    
    // Payment Methods
    PAYMENT_METHODS: {
        VA: 'VA',
        GOPAY: 'GOPAY',
        OVO: 'OVO',
        DANA: 'DANA',
        QRIS: 'QRIS',
        BCA: 'BCA',
    },
    
    // Status Colors
    STATUS_COLORS: {
        draft: '#95a5a6',
        sent: '#f39c12',
        partial: '#e67e22',
        paid: '#27ae60',
        overdue: '#e74c3c',
        cancelled: '#95a5a6',
    },
    
    // Status Labels
    STATUS_LABELS: {
        draft: 'Draft',
        sent: 'Sent',
        partial: 'Partial',
        paid: 'Paid',
        overdue: 'Overdue',
        cancelled: 'Cancelled',
    },
    
    // Format Options
    CURRENCY: 'IDR',
    DATE_FORMAT: 'DD MMM YYYY',
    DATETIME_FORMAT: 'DD MMM YYYY HH:mm',
    
    // Local Storage Keys
    STORAGE_KEYS: {
        CUSTOMER_ID: 'portal_customer_id',
        CUSTOMER_CODE: 'portal_customer_code',
        CUSTOMER_NAME: 'portal_customer_name',
        AUTH_TOKEN: 'portal_auth_token',
    },
    
    // Retry Configuration
    RETRY: {
        MAX_ATTEMPTS: 3,
        DELAY: 1000, // 1 second
    },
    
    // Timeout Configuration
    TIMEOUT: {
        API: 30000, // 30 seconds
        PAYMENT: 60000, // 60 seconds
    },
};

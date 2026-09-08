import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles/index.scss';
import App from './App';
import Layout from './Layout';

const root = ReactDOM.createRoot(
    document.getElementById('root') as HTMLElement
);
root.render(
    <React.StrictMode>
        <Layout>
            <App />
        </Layout>
    </React.StrictMode>
);

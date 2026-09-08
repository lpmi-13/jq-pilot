import type { ReactNode } from 'react';

const Layout = ({ children }: { children: ReactNode }) => {
    return (
        <div role="main" className="container">
            {children}
        </div>
    );
};

export default Layout;

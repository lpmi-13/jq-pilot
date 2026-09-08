import { render, screen } from '@testing-library/react';
import { beforeAll, expect, test, vi } from 'vitest';
import App from './App';

// App opens an SSE connection on mount; jsdom has no EventSource, so stub it.
beforeAll(() => {
    global.EventSource = vi.fn(() => ({
        close: vi.fn(),
        onmessage: null,
        onerror: null,
    }));
});

test('renders the loading state before any question arrives', () => {
    render(<App />);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
});

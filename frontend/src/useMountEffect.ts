import { useEffect } from 'react';

// A named wrapper around useEffect(..., []) that makes "run once on mount,
// clean up on unmount" explicit and keeps ad-hoc effects out of components.
// Use it for one-time external sync only: DOM integration, third-party widget
// lifecycles, and browser API subscriptions.
export function useMountEffect(callback: () => void | (() => void)) {
    // Mount-only by contract; exhaustive-deps can't see through the wrapper.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(callback, []);
}

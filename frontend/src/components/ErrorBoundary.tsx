import { Component } from 'react';
import type { ErrorInfo, ReactNode } from 'react';
import { C } from '../theme';
import { Button } from 'antd'

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<
  { children: ReactNode },
  ErrorBoundaryState
> {
  constructor(props: { children: ReactNode }) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary] Uncaught render error:', error, info.componentStack);
  }

  render() {
    if (this.state.hasError) {
      return (
        <div style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--paper)',
          padding: 32,
        }}>
          <div style={{
            maxWidth: 480,
            textAlign: 'center',
            background: 'var(--panel)',
            border: '1px solid var(--rule)',
            padding: 40,
          }}>
            <div style={{ fontSize: 40, marginBottom: 16 }}>&#x26A0;</div>
            <h1 style={{ fontSize: 20, fontWeight: 600, color: 'var(--ink)', marginBottom: 8, fontFamily: C.sans }}>
              Something went wrong
            </h1>
            <p style={{ fontSize: 14, color: 'var(--mid)', marginBottom: 20, fontFamily: C.sans, lineHeight: 1.5 }}>
              This page encountered an error. Try refreshing or navigating back.
            </p>
            {this.state.error && (
              <pre style={{
                fontSize: 12,
                fontFamily: C.mono,
                color: '#f87171',
                background: 'var(--flag-soft)',
                border: '1px solid var(--rule)',
                padding: 16,
                textAlign: 'left',
                overflow: 'auto',
                maxHeight: 120,
                marginBottom: 20,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
              }}>
                {this.state.error.message}
              </pre>
            )}
            <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
              <Button
                onClick={() => window.location.reload()}
                style={{
                  padding: '10px 20px',
                  border: '1px solid var(--ink)',
                  background: 'var(--ink)',
                  color: 'var(--paper)',
                  fontWeight: 500,
                  fontSize: 13,
                  cursor: 'pointer',
                  fontFamily: C.mono,
                }}
              >
                Refresh Page
              </Button>
              <Button
                onClick={() => { window.location.href = '/'; }}
                style={{
                  padding: '10px 20px',
                  border: '1px solid var(--rule)',
                  background: 'transparent',
                  color: 'var(--mid)',
                  fontWeight: 500,
                  fontSize: 13,
                  cursor: 'pointer',
                  fontFamily: C.mono,
                }}
              >
                Go Home
              </Button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

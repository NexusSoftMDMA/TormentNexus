import React from 'react';
import { createRoot } from 'react-dom/client';
import { logMessage } from '@src/utils/helpers';
import {
  applyDarkMode,
  applyLightMode,
  injectTailwindToShadowDom,
} from '@src/utils/shadowDom';
import Spotlight from './Spotlight';

export class SpotlightManager {
  private shadowHost: HTMLDivElement | null = null;
  private shadowRoot: ShadowRoot | null = null;
  private container: HTMLDivElement | null = null;
  private root: ReturnType<typeof createRoot> | null = null;
  private _isInitialized = false;
  private _initializationPromise: Promise<void> | null = null;

  public async initialize(): Promise<void> {
    if (this._isInitialized) return Promise.resolve();
    if (this._initializationPromise) return this._initializationPromise;

    this._initializationPromise = new Promise<void>(async (resolve, reject) => {
      try {
        if (!document.body) {
          await new Promise<void>(res => {
            const checkBody = () => {
              if (document.body) res();
              else setTimeout(checkBody, 10);
            };
            checkBody();
          });
        }

        // Create shadow host
        this.shadowHost = document.createElement('div');
        this.shadowHost.id = 'mcp-spotlight-shadow-host';
        this.shadowHost.style.position = 'fixed';
        this.shadowHost.style.top = '0';
        this.shadowHost.style.left = '0';
        this.shadowHost.style.width = '100vw';
        this.shadowHost.style.height = '100vh';
        this.shadowHost.style.zIndex = '2147483647'; // Max z-index
        this.shadowHost.style.pointerEvents = 'none'; // Pass through clicks when closed
        
        document.body.appendChild(this.shadowHost);

        this.shadowRoot = this.shadowHost.attachShadow({ mode: 'open' });

        this.container = document.createElement('div');
        this.container.id = 'spotlight-container';
        this.container.style.width = '100%';
        this.container.style.height = '100%';
        this.container.style.pointerEvents = 'none'; // Will be enabled by Spotlight component
        
        this.shadowRoot.appendChild(this.container);

        try {
          await injectTailwindToShadowDom(this.shadowRoot);
        } catch (e) {
          logMessage(`[SpotlightManager] CSS injection failed: ${e}`);
        }

        this.applyThemeClass('system');

        this.root = createRoot(this.container);
        
        // Render the Spotlight globally
        this.root.render(React.createElement(Spotlight, { 
          hostPointerEventsToggle: (enable: boolean) => {
            if (this.shadowHost && this.container) {
              this.shadowHost.style.pointerEvents = enable ? 'auto' : 'none';
              this.container.style.pointerEvents = enable ? 'auto' : 'none';
            }
          }
        }));

        this._isInitialized = true;
        resolve();
      } catch (e) {
        this._isInitialized = false;
        reject(e);
      } finally {
        this._initializationPromise = null;
      }
    });

    return this._initializationPromise;
  }

  public applyThemeClass(theme: 'light' | 'dark' | 'system'): void {
    if (!this.shadowHost || !this.shadowRoot) return;
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    this.shadowHost.classList.remove('light', 'dark');

    if (theme === 'dark' || (theme === 'system' && prefersDark)) {
      this.shadowHost.classList.add('dark');
      applyDarkMode(this.shadowRoot);
    } else {
      this.shadowHost.classList.add('light');
      applyLightMode(this.shadowRoot);
    }
  }

  public destroy(): void {
    if (this.root) {
      this.root.unmount();
      this.root = null;
    }
    if (this.shadowHost && this.shadowHost.parentNode) {
      this.shadowHost.parentNode.removeChild(this.shadowHost);
    }
    this.shadowHost = null;
    this.shadowRoot = null;
    this.container = null;
    this._isInitialized = false;
  }
}

// Singleton instance
export const spotlightManager = new SpotlightManager();

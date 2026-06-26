/**
 * Helper Utilities
 *
 * This file contains helper functions for the content script.
 * You can add your utility functions here as needed.
 */

/**
 * Example utility function
 * @param message The message to log
 */
import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('logMessage');

export const logMessage = (message: string): void => {
  logger.debug(`: ${message}`);
};

/**
 * Injects CSS into a Shadow DOM with proper error handling
 *
 * @param shadowRoot The Shadow DOM root to inject styles into
 * @param cssPath The path to the CSS file relative to the extension root
 * @returns Promise that resolves when the CSS is injected or rejects with an error
 */
export const injectCSSIntoShadowDOM = async (shadowRoot: ShadowRoot, cssPath: string): Promise<void> => {
  if (!shadowRoot) {
    throw new Error('Shadow root is not available for style injection');
  }

  try {
    const cssUrl = chrome.runtime.getURL(cssPath);
    logMessage(`Fetching CSS from: ${cssUrl}`);

    const response = await fetch(cssUrl);
    if (!response.ok) {
      throw new Error(`Failed to fetch CSS: ${response.statusText} (URL: ${cssUrl})`);
    }

    const cssText = await response.text();
    if (cssText.length === 0) {
      throw new Error('CSS content is empty');
    }

    logMessage(`Fetched CSS content (${cssText.length} bytes)`);

    const styleElement = document.createElement('style');
    styleElement.textContent = cssText;
    shadowRoot.appendChild(styleElement);

    logMessage('Successfully injected CSS into Shadow DOM');
    return Promise.resolve();
  } catch (error) {
    logMessage(`Error injecting CSS into Shadow DOM: ${error instanceof Error ? error.message : String(error)}`);
    throw error;
  }
};

/**
 * Utility for debugging Shadow DOM styling issues
 * This helps identify which styles are being properly applied
 * Only use this in development mode
 *
 * @param shadowRoot The Shadow DOM root to debug
 */
export const debugShadowDomStyles = (shadowRoot: ShadowRoot): void => {
  if (!shadowRoot) {
    logMessage('Cannot debug styles: Shadow root is null');
    return;
  }

  // Count all style elements
  const styleElements = shadowRoot.querySelectorAll('style');
  logMessage(`Shadow DOM contains ${styleElements.length} style elements`);

  // Log CSS rule count
  let totalRules = 0;
  styleElements.forEach((style, index) => {
    if (style.sheet) {
      const ruleCount = style.sheet.cssRules.length;
      totalRules += ruleCount;
      logMessage(`Style element #${index + 1} has ${ruleCount} CSS rules`);
    } else {
      logMessage(`Style element #${index + 1} has no CSS sheet attached`);
    }
  });

  logMessage(`Total CSS rules in Shadow DOM: ${totalRules}`);
};

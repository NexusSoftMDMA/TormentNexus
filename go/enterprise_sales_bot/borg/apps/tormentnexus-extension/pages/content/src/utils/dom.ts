// DOM Utility Functions

/**
 * Creates an HTML element with specified attributes and children.
 * @param tag - The HTML tag name.
 * @param attrs - An object of attributes to set on the element.
 * @param children - An array of child nodes or strings to append.
 * @returns The created HTML element.
 */
import { createLogger } from '@extension/shared/lib/logger';

const logger = createLogger('DOMUtils');

export function createElement<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  attrs: Record<string, any> = {},
  children: (Node | string)[] = [],
): HTMLElementTagNameMap[K] {
  const element = document.createElement(tag);
  for (const key in attrs) {
    if (Object.prototype.hasOwnProperty.call(attrs, key)) {
      element.setAttribute(key, attrs[key]);
    }
  }
  children.forEach(child => {
    if (typeof child === 'string') {
      element.appendChild(document.createTextNode(child));
    } else {
      element.appendChild(child);
    }
  });
  logger.debug(`Created <${tag}> element.`);
  return element;
}

/**
 * Waits for an element to appear in the DOM.
 * @param selector - The CSS selector for the element.
 * @param timeout - Maximum time to wait in milliseconds.
 * @param root - The root element to search within (default: document).
 * @returns A promise that resolves with the element or null if not found within timeout.
 */
export function waitForElement(
  selector: string,
  timeout = 5000,
  root: Document | Element = document,
): Promise<HTMLElement | null> {
  logger.debug(`Waiting for selector: "${selector}" with timeout ${timeout}ms.`);
  return new Promise(resolve => {
    const startTime = Date.now();
    const observer = new MutationObserver((mutationsList, obs) => {
      const element = root.querySelector(selector) as HTMLElement | null;
      if (element) {
        obs.disconnect();
        logger.debug(`Element "${selector}" found.`);
        resolve(element);
        return;
      }
      if (Date.now() - startTime > timeout) {
        obs.disconnect();
        logger.warn(`Timeout waiting for element "${selector}".`);
        resolve(null);
      }
    });

    // Check if element already exists
    const existingElement = root.querySelector(selector) as HTMLElement | null;
    if (existingElement) {
      logger.debug(`Element "${selector}" already exists.`);
      resolve(existingElement);
      return;
    }

    observer.observe(root === document ? document.documentElement : root, {
      childList: true,
      subtree: true,
    });
  });
}

/**
 * Injects a CSS string into the document's head.
 * @param css - The CSS string to inject.
 * @param id - An optional ID for the style tag.
 * @returns The created style element.
 */
export function injectCSS(css: string, id?: string): HTMLStyleElement {
  const styleElement = document.createElement('style');
  if (id) {
    styleElement.id = id;
  }
  styleElement.textContent = css;
  document.head.appendChild(styleElement);
  logger.debug(`CSS injected${id ? ' with ID: ' + id : ''}.`);
  return styleElement;
}

/**
 * Observes DOM mutations on a target node.
 * @param targetNode - The node to observe.
 * @param callback - The function to call on mutations.
 * @param options - MutationObserverInit options.
 * @returns The MutationObserver instance.
 */
export function observeChanges(
  targetNode: Node,
  callback: MutationCallback,
  options: MutationObserverInit,
): MutationObserver {
  const observer = new MutationObserver(callback);
  observer.observe(targetNode, options);
  logger.debug('[Utils.observeChanges] Mutation observer started.');
  return observer;
}

export type FindActionButtonOptions = {
  actionLabels: string[];
  preferredSelectors?: string[];
  root?: ParentNode;
  near?: HTMLElement | null;
  iconPathHints?: string[];
};

const ACTION_MENU_HINT_PATTERN =
  /\b(menu|dropdown|options|more|model|voice|attach|upload|search|web search|tools?)\b|ask every time|always run/;

const normalizeText = (value: string | null | undefined): string => value?.trim().toLowerCase() ?? '';

const escapeRegExp = (value: string): string => value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

const getButtonDescriptor = (button: HTMLButtonElement): string =>
  normalizeText(
    [
      button.textContent,
      button.getAttribute('aria-label'),
      button.getAttribute('title'),
      button.getAttribute('name'),
      button.getAttribute('data-testid'),
      button.className,
    ]
      .filter(Boolean)
      .join(' '),
  );

export const isElementVisible = (element: HTMLElement): boolean => {
  const style = window.getComputedStyle(element);
  const rect = element.getBoundingClientRect();
  const hasRenderableContent =
    Boolean(element.textContent?.trim()) || Boolean(element.querySelector('svg, img, span, path'));

  return !(
    style.display === 'none' ||
    style.visibility === 'hidden' ||
    style.opacity === '0' ||
    element.hidden ||
    element.getAttribute('aria-hidden') === 'true' ||
    ((rect.width <= 0 || rect.height <= 0) && !hasRenderableContent)
  );
};

export const isButtonDisabled = (button: HTMLButtonElement): boolean =>
  button.disabled ||
  button.getAttribute('disabled') !== null ||
  button.getAttribute('aria-disabled') === 'true' ||
  button.classList.contains('disabled');

export const findBestActionButton = ({
  actionLabels,
  preferredSelectors = [],
  root = document,
  near = null,
  iconPathHints = [],
}: FindActionButtonOptions): HTMLButtonElement | null => {
  const labels = actionLabels.map(normalizeText).filter(Boolean);
  if (labels.length === 0) {
    return null;
  }

  const candidates = new Set<HTMLButtonElement>();
  const addCandidate = (element: Element | null | undefined) => {
    if (!element) {
      return;
    }

    const button = element instanceof HTMLButtonElement ? element : element.closest('button');
    if (button instanceof HTMLButtonElement) {
      candidates.add(button);
    }
  };

  for (const selector of preferredSelectors) {
    addCandidate(root.querySelector(selector));
  }

  const nearbyScopes = [
    near?.closest('form'),
    near?.parentElement,
    near?.closest('[role="group"]'),
    near?.closest('section'),
    near?.closest('div'),
  ].filter(Boolean) as Element[];

  for (const scope of nearbyScopes) {
    scope.querySelectorAll('button').forEach(button => addCandidate(button));
  }

  root.querySelectorAll('button').forEach(button => addCandidate(button));

  let bestButton: HTMLButtonElement | null = null;
  let bestScore = Number.NEGATIVE_INFINITY;

  for (const button of candidates) {
    if (!isElementVisible(button)) {
      continue;
    }

    const descriptor = getButtonDescriptor(button);
    let score = 0;

    if (!descriptor && button.type !== 'submit' && iconPathHints.length === 0) {
      continue;
    }

    if (
      button.getAttribute('aria-haspopup') !== null ||
      button.getAttribute('aria-controls')?.toLowerCase().includes('menu') ||
      button.getAttribute('role') === 'menuitem' ||
      ACTION_MENU_HINT_PATTERN.test(descriptor)
    ) {
      score -= 250;
    }

    for (const label of labels) {
      const exactPattern = new RegExp(`^${escapeRegExp(label)}$`, 'i');
      const wholeWordPattern = new RegExp(`\\b${escapeRegExp(label)}\\b`, 'i');

      if (exactPattern.test(normalizeText(button.getAttribute('aria-label')))) {
        score += 200;
      }
      if (exactPattern.test(normalizeText(button.textContent))) {
        score += 180;
      }
      if (wholeWordPattern.test(descriptor)) {
        score += 90;
      }
    }

    if (button.type === 'submit') {
      score += 40;
    }

    if (near && button.closest('form') && button.closest('form') === near.closest('form')) {
      score += 30;
    }

    if (near && near.parentElement && button.parentElement === near.parentElement) {
      score += 20;
    }

    if (iconPathHints.some(hint => button.querySelector(`svg path[d*="${hint}"]`))) {
      score += 60;
    }

    if (score > bestScore) {
      bestScore = score;
      bestButton = button;
    }
  }

  return bestScore > 0 ? bestButton : null;
};

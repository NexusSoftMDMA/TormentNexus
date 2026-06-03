/**
 * Services Index
 *
 * Centralized export point for all application services
 */

import { createLogger } from '@extension/shared/lib/logger';
const logger = createLogger('Services Index');

export {
  AutomationService,
} from './automation.service';

// Export initialization function for all services
export async function initializeAllServices(): Promise<void> {
  logger.debug('[Services] Initializing all application services...');

  try {
    // Initialize automation service
    const { AutomationService } = await import('./automation.service');
    AutomationService.getInstance();

    // Initialize memory capture service
    const { memoryCaptureService } = await import('./memory-capture.service');
    memoryCaptureService.setEnabled(true);

    // Initialize TormentNexus Kernel button on supported AI chat sites
    const { initTormentNexusKernelButton } = await import('./tormentnexus-kernel-button');
    await initTormentNexusKernelButton();

    logger.debug('[Services] All services initialized successfully');
  } catch (error) {
    logger.error('[Services] Error initializing services:', error);
    throw error;
  }
}

// Export cleanup function for all services
export async function cleanupAllServices(): Promise<void> {
  logger.debug('[Services] Cleaning up all application services...');

  try {
    // Cleanup automation service
    // const { cleanupAutomationService } = await import('./automation.service');
    // cleanupAutomationService();

    logger.debug('[Services] All services cleaned up successfully');
  } catch (error) {
    logger.error('[Services] Error cleaning up services:', error);
  }
}

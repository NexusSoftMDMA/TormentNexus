import { UiAutomationManager } from './ui_automation.js';

export class InputManager {
    private automation = new UiAutomationManager();

    async sendKeys(keys: string, windowTitle?: string) {
        try {
            return await this.automation.sendKeys(keys, windowTitle);
        } catch (error: any) {
            return `Error sending keys: ${error.message}`;
        }
    }
}

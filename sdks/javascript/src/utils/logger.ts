/**
 * Logging utilities
 */

import { Logger } from '../types';

class ConsoleLogger implements Logger {
  constructor(private enabled: boolean = false) {}

  debug(message: string, ...args: any[]): void {
    if (this.enabled) {
      console.debug(`[EvalForge DEBUG] ${message}`, ...args);
    }
  }

  info(message: string, ...args: any[]): void {
    if (this.enabled) {
      console.info(`[EvalForge INFO] ${message}`, ...args);
    }
  }

  warn(message: string, ...args: any[]): void {
    if (this.enabled) {
      console.warn(`[EvalForge WARN] ${message}`, ...args);
    }
  }

  error(message: string, ...args: any[]): void {
    if (this.enabled) {
      console.error(`[EvalForge ERROR] ${message}`, ...args);
    }
  }
}

class NoOpLogger implements Logger {
  debug(): void {}
  info(): void {}
  warn(): void {}
  error(): void {}
}

export function createLogger(debug: boolean = false): Logger {
  return debug ? new ConsoleLogger(true) : new NoOpLogger();
}
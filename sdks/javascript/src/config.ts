/**
 * Configuration management
 */

import { EvalForgeConfig } from './types';
import { EvalForge } from './client';

interface GlobalConfig {
  apiKey?: string;
  projectId?: number;
  baseUrl?: string;
  batchSize?: number;
  flushInterval?: number;
  maxRetries?: number;
  timeout?: number;
  debug?: boolean;
}

let globalConfig: GlobalConfig = {};

/**
 * Load configuration from environment variables
 */
function loadFromEnv(): GlobalConfig {
  const config: GlobalConfig = {};

  // Node.js environment variables
  if (typeof process !== 'undefined' && process.env) {
    config.apiKey = process.env.EVALFORGE_API_KEY;
    
    if (process.env.EVALFORGE_PROJECT_ID) {
      const projectId = parseInt(process.env.EVALFORGE_PROJECT_ID, 10);
      if (!isNaN(projectId)) {
        config.projectId = projectId;
      }
    }

    config.baseUrl = process.env.EVALFORGE_BASE_URL;

    if (process.env.EVALFORGE_BATCH_SIZE) {
      const batchSize = parseInt(process.env.EVALFORGE_BATCH_SIZE, 10);
      if (!isNaN(batchSize)) {
        config.batchSize = batchSize;
      }
    }

    if (process.env.EVALFORGE_FLUSH_INTERVAL) {
      const flushInterval = parseFloat(process.env.EVALFORGE_FLUSH_INTERVAL);
      if (!isNaN(flushInterval)) {
        config.flushInterval = flushInterval * 1000; // Convert to milliseconds
      }
    }

    if (process.env.EVALFORGE_MAX_RETRIES) {
      const maxRetries = parseInt(process.env.EVALFORGE_MAX_RETRIES, 10);
      if (!isNaN(maxRetries)) {
        config.maxRetries = maxRetries;
      }
    }

    if (process.env.EVALFORGE_TIMEOUT) {
      const timeout = parseFloat(process.env.EVALFORGE_TIMEOUT);
      if (!isNaN(timeout)) {
        config.timeout = timeout * 1000; // Convert to milliseconds
      }
    }

    if (process.env.EVALFORGE_DEBUG) {
      config.debug = ['true', '1', 'yes', 'on'].includes(
        process.env.EVALFORGE_DEBUG.toLowerCase()
      );
    }
  }

  return config;
}

/**
 * Configure global settings
 */
export function configure(config: Partial<EvalForgeConfig>): void {
  globalConfig = {
    ...globalConfig,
    ...config,
  };
}

/**
 * Get current configuration
 */
export function getConfig(): GlobalConfig {
  // Merge environment variables with explicit configuration
  const envConfig = loadFromEnv();
  return {
    ...envConfig,
    ...globalConfig,
  };
}

/**
 * Create a client with current global configuration
 */
export function createClient(overrides: Partial<EvalForgeConfig> = {}): EvalForge {
  const config = getConfig();
  
  const clientConfig: EvalForgeConfig = {
    apiKey: overrides.apiKey || config.apiKey || '',
    projectId: overrides.projectId || config.projectId || 0,
    baseUrl: overrides.baseUrl || config.baseUrl,
    batchSize: overrides.batchSize || config.batchSize,
    flushInterval: overrides.flushInterval || config.flushInterval,
    maxRetries: overrides.maxRetries || config.maxRetries,
    timeout: overrides.timeout || config.timeout,
    debug: overrides.debug !== undefined ? overrides.debug : config.debug,
  };

  if (!clientConfig.apiKey) {
    throw new Error('EvalForge API key is required');
  }

  if (!clientConfig.projectId) {
    throw new Error('EvalForge project ID is required');
  }

  return new EvalForge(clientConfig);
}

/**
 * Reset configuration to defaults
 */
export function resetConfig(): void {
  globalConfig = {};
}